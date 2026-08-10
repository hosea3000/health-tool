//go:build windows

package main

import (
	"fmt"
	"log"

	toast "git.sr.ht/~jackmordaunt/go-toast/v2"
)

func newNotifiers() (func(int), func(int)) {
	if err := toast.SetAppData(toast.AppData{AppID: "health-tool", GUID: "7d4f8d18-9c67-4c57-8ef4-7f4eb4f33e1b"}); err != nil {
		log.Printf("windows toast setup unavailable: %v", err)
	}
	reminder := func(restMinutes int) {
		notification := toast.Notification{
			AppID: "health-tool",
			Title: "久坐提醒",
			Body:  fmt.Sprintf("请休息 %d 分钟。", restMinutes),
			Audio: toast.Default,
		}
		if err := notification.Push(); err != nil {
			log.Printf("windows toast unavailable: %v", err)
		}
	}
	started := func(reminderMinutes int) {
		notification := toast.Notification{
			AppID: "health-tool",
			Title: "久坐提醒",
			Body:  fmt.Sprintf("新的工作段已开始，提醒将在 %d 分钟后触发。", reminderMinutes),
			Audio: toast.Default,
		}
		if err := notification.Push(); err != nil {
			log.Printf("windows toast unavailable: %v", err)
		}
	}
	return reminder, started
}
