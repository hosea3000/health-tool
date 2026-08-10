//go:build !windows

package main

func (a *App) startTray() {}

func updateTrayState(string) {}

func stopTray() {}
