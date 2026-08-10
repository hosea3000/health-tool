//go:build !windows

package main

import "health-tool/domain"

func startInputMonitor(_ func(domain.EffectiveActivity)) (func(), error) {
	return func() {}, nil
}
