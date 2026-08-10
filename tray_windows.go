//go:build windows

package main

import (
	"sync/atomic"

	"github.com/getlantern/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var trayReady atomic.Bool

func (a *App) startTray() {
	go systray.Run(func() {
		trayReady.Store(true)
		systray.SetTitle("久坐提醒")
		systray.SetTooltip("久坐提醒")
		updateTrayState(a.Status().State)
		openItem := systray.AddMenuItem("打开主界面", "显示久坐提醒主界面")
		settingsItem := systray.AddMenuItem("设置", "调整工作段提醒时长")
		quitItem := systray.AddMenuItem("退出", "退出久坐提醒")
		go func() {
			for {
				select {
				case <-openItem.ClickedCh:
					runtime.WindowShow(a.ctx)
				case <-settingsItem.ClickedCh:
					runtime.WindowShow(a.ctx)
				case <-quitItem.ClickedCh:
					runtime.Quit(a.ctx)
					return
				}
			}
		}()
	}, func() {})
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
		systray.SetTooltip("久坐提醒 · " + label)
	}
}

func stopTray() {
	trayReady.Store(false)
	systray.Quit()
}
