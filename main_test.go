package main

import (
	"testing"
	"time"
)

func TestAppOptionsEnableSingleInstanceActivation(t *testing.T) {
	app := newApp(func() time.Time { return time.Unix(0, 0) }, func() {})
	options := newAppOptions(app, false)

	if options.SingleInstanceLock == nil {
		t.Fatal("single-instance lock is not configured")
	}
	if options.SingleInstanceLock.UniqueId != "health-tool" {
		t.Fatalf("single-instance ID = %q, want health-tool", options.SingleInstanceLock.UniqueId)
	}
	if options.SingleInstanceLock.OnSecondInstanceLaunch == nil {
		t.Fatal("second-instance callback is not configured")
	}
	if len(options.Bind) != 1 || options.Bind[0] != app {
		t.Fatal("app binding was not preserved")
	}
}

func TestAppOptionsStartHidden(t *testing.T) {
	app := newApp(func() time.Time { return time.Unix(0, 0) }, func() {})
	if !newAppOptions(app, true).StartHidden {
		t.Fatal("StartHidden = false, want true when --hidden is passed")
	}
	if newAppOptions(app, false).StartHidden {
		t.Fatal("StartHidden = true, want false without --hidden")
	}
}
