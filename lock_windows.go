//go:build windows

package main

import "golang.org/x/sys/windows"

const desktopSwitch = 0x0100

var (
	openInputDesktop = windows.NewLazySystemDLL("user32.dll").NewProc("OpenInputDesktop")
	closeDesktop     = windows.NewLazySystemDLL("user32.dll").NewProc("CloseDesktop")
)

func workstationLocked() bool {
	desktop, _, _ := openInputDesktop.Call(0, 0, desktopSwitch)
	if desktop == 0 {
		return true
	}
	closeDesktop.Call(desktop)
	return false
}
