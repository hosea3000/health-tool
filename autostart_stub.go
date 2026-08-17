//go:build !windows

package main

import "errors"

// setAutoStart 非 Windows 平台不支持开机自启动。
func setAutoStart(enabled bool) error {
	return errors.New("开机自启动仅支持 Windows 平台")
}

// autoStartEnabled 非 Windows 平台不支持查询自启状态。
func autoStartEnabled() (bool, error) {
	return false, errors.New("开机自启动仅支持 Windows 平台")
}
