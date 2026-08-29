//go:build windows

package main

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows/registry"
)

// autoStartRunKey 是 Windows 当前用户的开机自启动注册表键。
const autoStartRunKey = `Software\Microsoft\Windows\CurrentVersion\Run`

// setAutoStart 开启时向 HKCU Run key 写入本应用自启命令，关闭时删除该值。
func setAutoStart(enabled bool) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, autoStartRunKey, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("打开注册表 Run key 失败: %w", err)
	}
	defer key.Close()
	if !enabled {
		// 已关闭时再次关闭视为成功（幂等）。
		if err := key.DeleteValue(autoStartValueName); err != nil && err != registry.ErrNotExist {
			return fmt.Errorf("删除自启注册表值失败: %w", err)
		}
		return nil
	}
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取程序路径失败: %w", err)
	}
	if err := key.SetStringValue(autoStartValueName, buildAutoStartCommand(exePath)); err != nil {
		return fmt.Errorf("写入自启注册表值失败: %w", err)
	}
	return nil
}

// autoStartEnabled 查询自启状态：Run key 下存在非空 health-tool 值即视为开启。
// Run key 或值本身不存在时返回未开启而非错误。
func autoStartEnabled() (bool, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, autoStartRunKey, registry.QUERY_VALUE)
	if err != nil {
		if err == registry.ErrNotExist {
			return false, nil
		}
		return false, fmt.Errorf("打开注册表 Run key 失败: %w", err)
	}
	defer key.Close()
	value, _, err := key.GetStringValue(autoStartValueName)
	if err != nil {
		if err == registry.ErrNotExist {
			return false, nil
		}
		return false, fmt.Errorf("读取自启注册表值失败: %w", err)
	}
	return value != "", nil
}
