//go:build windows

package main

import (
	_ "embed"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows"
)

//go:embed build/windows/icon.ico
var trayIcon []byte

var trayReady atomic.Bool

// 托盘菜单命令 ID
const (
	cmdOpen     = 1
	cmdSettings = 2
	cmdQuit     = 3
)

const (
	wmTrayMsg       = 0x0401 // Shell_NotifyIcon 回调消息
	wmRButtonUp     = 0x0205
	wmLButtonUp     = 0x0202
	wmClose         = 0x0010
	wmDestroy       = 0x0002
	wmEndSession    = 0x0016
	wmPowerChange   = 0x0218
	wmWtsChange     = 0x02B1
	pbtResumeAuto   = 0x0012 // PBT_APMRESUMEAUTOMATIC
	pbtResume       = 0x0007 // PBT_APMRESUMESUSPEND
	wtsUnlock       = 0x0008 // WTS_SESSION_UNLOCK
	notifyThisSess  = 0     // NOTIFY_FOR_THIS_SESSION

	nimAdd    = 0
	nimModify = 1
	nimDelete = 2

	nifMessage = 0x01
	nifIcon    = 0x02
	nifTip     = 0x04

	tpmBottomAlign = 0x0020
	tpmLeftAlign   = 0x0000
	tpmReturnCmd   = 0x0100

	lrLoadFromFile = 0x10
	lrDefaultSize  = 0x40
	imgIcon        = 1
)

var (
	// user32 在 input_windows.go 中已声明，此处复用
	shell  = windows.NewLazySystemDLL("shell32.dll")
	wtsapi = windows.NewLazySystemDLL("wtsapi32.dll")

	procRegisterWindowMessageW = user32.NewProc("RegisterWindowMessageW")
	procRegisterClassExW       = user32.NewProc("RegisterClassExW")
	procUnregisterClassW       = user32.NewProc("UnregisterClassW")
	procCreateWindowExW        = user32.NewProc("CreateWindowExW")
	procDefWindowProcW         = user32.NewProc("DefWindowProcW")
	procDestroyWindow          = user32.NewProc("DestroyWindow")
	procPostMessageW           = user32.NewProc("PostMessageW")
	procPostQuitMessage        = user32.NewProc("PostQuitMessage")
	procGetMessageW            = user32.NewProc("GetMessageW")
	procTranslateMessage       = user32.NewProc("TranslateMessage")
	procDispatchMessageW       = user32.NewProc("DispatchMessageW")
	procGetCursorPos           = user32.NewProc("GetCursorPos")
	procSetForegroundWindow    = user32.NewProc("SetForegroundWindow")
	procCreatePopupMenu        = user32.NewProc("CreatePopupMenu")
	procDestroyMenu            = user32.NewProc("DestroyMenu")
	procAppendMenuW            = user32.NewProc("AppendMenuW")
	procTrackPopupMenu         = user32.NewProc("TrackPopupMenu")
	procLoadImageW             = user32.NewProc("LoadImageW")
	// kernel32 在 input_windows.go 中已声明，此处复用；GetModuleHandleW 属于 kernel32
	procGetModuleHandleW       = kernel32.NewProc("GetModuleHandleW")

	procShellNotifyIcon = shell.NewProc("Shell_NotifyIconW")

	procWtsRegisterNotification   = wtsapi.NewProc("WTSRegisterSessionNotification")
	procWtsUnregisterNotification = wtsapi.NewProc("WTSUnRegisterSessionNotification")
)

type wndClassEx struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   windows.Handle
	Icon       windows.Handle
	Cursor     windows.Handle
	Background windows.Handle
	MenuName   *uint16
	ClassName  *uint16
	IconSm     windows.Handle
}

// point 在 input_windows.go 中已声明，此处复用

type msg struct {
	hwnd     uintptr
	message  uint32
	wParam   uintptr
	lParam   uintptr
	time     uint32
	pt       point
	lPrivate uint32
}

type notifyIconData struct {
	cbSize           uint32
	hWnd             windows.Handle
	uID              uint32
	uFlags           uint32
	uCallbackMessage uint32
	hIcon            windows.Handle
	szTip            [128]uint16
	dwState          uint32
	dwStateMask      uint32
	szInfo           [256]uint16
	uVersion         uint32
	szInfoTitle      [64]uint16
	dwInfoFlags      uint32
	guidItem         windows.GUID
	hBalloonIcon     windows.Handle
}

type tray struct {
	app       *App
	wndProcFn uintptr
	window    windows.Handle
	menu      windows.Handle
	instance  windows.Handle
	className *uint16
	nid       notifyIconData
	nidMu     sync.Mutex
	wmTaskbar uint32
}

var wt tray

// startTray 创建托盘图标与菜单。消息泵运行在锁定的 OS 线程上，
// 保证 Win32 窗口消息始终被同一线程消费（锁屏/睡眠恢复后不失效）。
func (a *App) startTray() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				trayReady.Store(false)
				log.Printf("tray unavailable: %v", r)
			}
		}()
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		if err := wt.init(a); err != nil {
			log.Printf("tray init failed: %v", err)
			return
		}
		trayReady.Store(true)
		wt.setTooltip("久坐提醒")
		wt.pump()
	}()
}

func (t *tray) init(a *App) error {
	t.app = a

	// 注册 explorer 重启广播消息（TaskbarCreated）
	if name, err := windows.UTF16PtrFromString("TaskbarCreated"); err == nil {
		if res, _, _ := procRegisterWindowMessageW.Call(uintptr(unsafe.Pointer(name))); res != 0 {
			t.wmTaskbar = uint32(res)
		}
	}

	iconPath, err := t.tempIconPath()
	if err != nil {
		return err
	}
	res, _, _ := procGetModuleHandleW.Call(0)
	if res == 0 {
		return syscallError("GetModuleHandleW")
	}
	t.instance = windows.Handle(res)
	iconRes, _, _ := procLoadImageW.Call(0, uintptr(unsafe.Pointer(iconPath)), imgIcon, 0, 0, lrLoadFromFile|lrDefaultSize)
	if iconRes == 0 {
		return syscallError("LoadImageW")
	}

	t.wndProcFn = windows.NewCallback(t.wndProc)
	t.className, _ = windows.UTF16PtrFromString("HealthToolTrayClass")
	wc := wndClassEx{
		Style:     0,
		WndProc:   t.wndProcFn,
		Instance:  t.instance,
		Icon:      windows.Handle(iconRes),
		Cursor:    windows.Handle(0),
		ClassName: t.className,
	}
	wc.Size = uint32(unsafe.Sizeof(wc))
	if res, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc))); res == 0 {
		return syscallError("RegisterClassExW")
	}

	winName, _ := windows.UTF16PtrFromString("")
	hWnd, _, _ := procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(t.className)), uintptr(unsafe.Pointer(winName)), 0, 0, 0, 0, 0, 0, 0, uintptr(t.instance), 0)
	if hWnd == 0 {
		return syscallError("CreateWindowExW")
	}
	t.window = windows.Handle(hWnd)

	t.nid = notifyIconData{
		hWnd:             t.window,
		uID:              100,
		uFlags:           nifMessage | nifIcon | nifTip,
		uCallbackMessage: wmTrayMsg,
		hIcon:            windows.Handle(iconRes),
	}
	t.nid.cbSize = uint32(unsafe.Sizeof(t.nid))
	if err := t.addIcon(); err != nil {
		return err
	}

	if err := t.buildMenu(); err != nil {
		return err
	}

	// 监听会话解锁（锁屏/睡眠恢复后重新注册图标）
	procWtsRegisterNotification.Call(uintptr(t.window), notifyThisSess)
	return nil
}

// pump 消费窗口消息，直到收到退出消息（stopTray）。
func (t *tray) pump() {
	var m msg
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if ret == 0 || ret == ^uintptr(0) {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
	trayReady.Store(false)
	procWtsUnregisterNotification.Call(uintptr(t.window))
	procDestroyMenu.Call(uintptr(t.menu))
	procUnregisterClassW.Call(uintptr(unsafe.Pointer(t.className)), uintptr(t.instance))
}

func (t *tray) wndProc(hWnd windows.Handle, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case wmTrayMsg:
		switch lParam {
		case wmLButtonUp:
			t.app.showMainWindow()
		case wmRButtonUp:
			t.showMenu()
		}
	case t.wmTaskbar: // explorer.exe 重启
		t.addIcon()
	case wmPowerChange: // 睡眠恢复
		if wParam == pbtResumeAuto || wParam == pbtResume {
			t.addIcon()
		}
	case wmWtsChange: // 解锁
		if uint32(wParam) == wtsUnlock {
			t.addIcon()
		}
	case wmEndSession:
		// 不处理：会话结束/注销广播不影响托盘，保持窗口存活
	case wmClose:
		procDestroyWindow.Call(uintptr(t.window))
	case wmDestroy:
		procPostQuitMessage.Call(0)
	default:
		res, _, _ := procDefWindowProcW.Call(uintptr(hWnd), uintptr(message), wParam, lParam)
		return res
	}
	return 0
}

func (t *tray) showMenu() {
	var p point
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&p)))
	procSetForegroundWindow.Call(uintptr(t.window))
	res, _, _ := procTrackPopupMenu.Call(uintptr(t.menu), tpmBottomAlign|tpmLeftAlign|tpmReturnCmd, uintptr(p.x), uintptr(p.y), 0, uintptr(t.window), 0)
	switch uintptr(res) {
	case cmdOpen:
		t.app.showMainWindow()
	case cmdSettings:
		wailsruntime.WindowShow(t.app.ctx)
	case cmdQuit:
		t.app.requestQuit()
		wailsruntime.Quit(t.app.ctx)
	}
}

func (t *tray) buildMenu() error {
	hMenu, _, _ := procCreatePopupMenu.Call()
	if hMenu == 0 {
		return syscallError("CreatePopupMenu")
	}
	t.menu = windows.Handle(hMenu)
	items := []struct {
		cmd   uintptr
		title string
	}{
		{cmdOpen, "打开主界面"},
		{cmdSettings, "设置"},
		{cmdQuit, "退出"},
	}
	for _, item := range items {
		titlePtr, err := windows.UTF16PtrFromString(item.title)
		if err != nil {
			return err
		}
		if res, _, _ := procAppendMenuW.Call(uintptr(t.menu), 0, item.cmd, uintptr(unsafe.Pointer(titlePtr))); res == 0 {
			return syscallError("AppendMenuW")
		}
	}
	return nil
}

func (t *tray) addIcon() error {
	t.nidMu.Lock()
	defer t.nidMu.Unlock()
	if res, _, _ := procShellNotifyIcon.Call(nimAdd, uintptr(unsafe.Pointer(&t.nid))); res == 0 {
		return syscallError("Shell_NotifyIcon")
	}
	return nil
}

func (t *tray) setTooltip(text string) {
	tip, err := windows.UTF16FromString(text)
	if err != nil {
		return
	}
	t.nidMu.Lock()
	defer t.nidMu.Unlock()
	t.nid.uFlags |= nifTip
	copy(t.nid.szTip[:], tip)
	if len(tip) >= len(t.nid.szTip) {
		t.nid.szTip[len(t.nid.szTip)-1] = 0
	}
	procShellNotifyIcon.Call(nimModify, uintptr(unsafe.Pointer(&t.nid)))
}

// tempIconPath 把内嵌图标写到临时目录，供 LoadImageW 读取。
func (t *tray) tempIconPath() (*uint16, error) {
	path := filepath.Join(os.TempDir(), "health-tool-tray.ico")
	if _, err := os.Stat(path); err != nil {
		if err := os.WriteFile(path, trayIcon, 0o644); err != nil {
			return nil, err
		}
	}
	return windows.UTF16PtrFromString(path)
}

type syscallError string

func (e syscallError) Error() string {
	return string(e)
}

func updateTrayState(state string) {
	if !trayReady.Load() {
		return
	}
	labels := map[string]string{
		"waiting":     "待工作",
		"working":     "工作中",
		"idle-paused": "闲置暂停",
		"resting":     "休息中",
	}
	if label, ok := labels[state]; ok {
		wt.setTooltip("久坐提醒 · " + label)
	}
}

func stopTray() {
	if !trayReady.Load() {
		return
	}
	trayReady.Store(false)
	procPostMessageW.Call(uintptr(wt.window), wmClose, 0, 0)
}
