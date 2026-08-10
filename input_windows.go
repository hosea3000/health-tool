//go:build windows

package main

import (
	"health-tool/domain"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	whKeyboardLL = 13
	whMouseLL    = 14
	hcAction     = 0
	wmKeyDown    = 0x0100
	wmSysKeyDown = 0x0104
	wmLButton    = 0x0201
	wmRButton    = 0x0204
	wmMButton    = 0x0207
	wmMouseMove  = 0x0200
	wmMouseWheel = 0x020A
	wmQuit       = 0x0012
)

type point struct {
	x int32
	y int32
}

type mouseHookData struct {
	point point
	flags uint32
	data  uint32
	time  uint32
	info  uintptr
}

type windowsMessage struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	point   point
	private uint32
}

func startInputMonitor(emit func(domain.EffectiveActivity)) (func(), error) {
	ready := make(chan inputRuntime, 1)
	go runInputHooks(emit, ready)
	runtime := <-ready
	if runtime.err != nil {
		return func() {}, runtime.err
	}
	return runtime.stop, nil
}

type inputRuntime struct {
	stop func()
	err  error
}

func runInputHooks(emit func(domain.EffectiveActivity), ready chan<- inputRuntime) {
	var mu sync.Mutex
	events := make(chan domain.EffectiveActivity, 256)
	done := make(chan struct{})
	go func() {
		for {
			select {
			case event := <-events:
				emit(event)
			case <-done:
				return
			}
		}
	}()
	push := func(event domain.EffectiveActivity) {
		select {
		case events <- event:
		case <-done:
		default:
		}
	}
	lastPoint := point{}
	havePoint := false

	keyboardCallback := windows.NewCallback(func(code int, wParam, lParam uintptr) uintptr {
		if code == hcAction && (wParam == wmKeyDown || wParam == wmSysKeyDown) {
			push(domain.EffectiveActivity{Kind: domain.KeyPress, At: time.Now()})
		}
		return callNextHook(0, code, wParam, lParam)
	})
	mouseCallback := windows.NewCallback(func(code int, wParam, lParam uintptr) uintptr {
		if code == hcAction {
			data := (*mouseHookData)(unsafe.Pointer(lParam))
			switch wParam {
			case wmLButton, wmRButton, wmMButton:
				push(domain.EffectiveActivity{Kind: domain.Click, At: time.Now()})
			case wmMouseWheel:
				push(domain.EffectiveActivity{Kind: domain.Wheel, At: time.Now()})
			case wmMouseMove:
				mu.Lock()
				if havePoint {
					distance := abs(data.point.x-lastPoint.x) + abs(data.point.y-lastPoint.y)
					if significantMouseMovement(distance) {
						push(domain.EffectiveActivity{Kind: domain.MouseMove, At: time.Now()})
					}
				}
				lastPoint, havePoint = data.point, true
				mu.Unlock()
			}
		}
		return callNextHook(0, code, wParam, lParam)
	})

	keyboardHook, err := setHook(whKeyboardLL, keyboardCallback)
	if err != nil {
		close(done)
		ready <- inputRuntime{err: err}
		return
	}
	mouseHook, err := setHook(whMouseLL, mouseCallback)
	if err != nil {
		unhook(keyboardHook)
		close(done)
		ready <- inputRuntime{err: err}
		return
	}

	threadID := getCurrentThreadID()
	ready <- inputRuntime{stop: func() {
		unhook(keyboardHook)
		unhook(mouseHook)
		postThreadMessage(threadID, wmQuit)
		close(done)
	}}

	for {
		var message windowsMessage
		result, _, _ := getMessage.Call(uintptr(unsafe.Pointer(&message)), 0, 0, 0)
		if result == 0 || result == ^uintptr(0) {
			return
		}
		translateMessage.Call(uintptr(unsafe.Pointer(&message)))
		dispatchMessage.Call(uintptr(unsafe.Pointer(&message)))
	}
}

var (
	user32              = windows.NewLazySystemDLL("user32.dll")
	setWindowsHookEx    = user32.NewProc("SetWindowsHookExW")
	callNextWindowsHook = user32.NewProc("CallNextHookEx")
	unhookWindowsHook   = user32.NewProc("UnhookWindowsHookEx")
	getMessage          = user32.NewProc("GetMessageW")
	translateMessage    = user32.NewProc("TranslateMessage")
	dispatchMessage     = user32.NewProc("DispatchMessageW")
	postThread          = user32.NewProc("PostThreadMessageW")
	kernel32            = windows.NewLazySystemDLL("kernel32.dll")
	currentThread       = kernel32.NewProc("GetCurrentThreadId")
)

func setHook(kind int, callback uintptr) (windows.Handle, error) {
	hook, _, err := setWindowsHookEx.Call(uintptr(kind), callback, 0, 0)
	if hook == 0 {
		return 0, err
	}
	return windows.Handle(hook), nil
}

func unhook(hook windows.Handle) {
	unhookWindowsHook.Call(uintptr(hook))
}

func callNextHook(hook uintptr, code int, wParam, lParam uintptr) uintptr {
	result, _, _ := callNextWindowsHook.Call(hook, uintptr(code), wParam, lParam)
	return result
}

func getCurrentThreadID() uintptr {
	thread, _, _ := currentThread.Call()
	return thread
}

func postThreadMessage(thread, message uintptr) {
	postThread.Call(thread, message, 0, 0)
}

func abs(value int32) int32 {
	if value < 0 {
		return -value
	}
	return value
}
