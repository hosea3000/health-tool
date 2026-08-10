//go:build !windows

package main

import (
	"fmt"

	"github.com/gen2brain/beeep"
)

func newNotifiers() (func(int), func(int)) {
	reminder := func(restMinutes int) {
		_ = beeep.Notify("久坐提醒", fmt.Sprintf("请休息 %d 分钟。", restMinutes), "")
	}
	started := func(reminderMinutes int) {
		_ = beeep.Notify("久坐提醒", fmt.Sprintf("新的工作段已开始，提醒将在 %d 分钟后触发。", reminderMinutes), "")
	}
	return reminder, started
}
