package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

// version 是运行时版本号：发布构建由 CI 通过 -ldflags "-X main.version=<版本>" 注入，本地开发保持 dev。
var version = "dev"

func main() {
	app := NewApp()

	err := wails.Run(newAppOptions(app))
	if err != nil {
		println("Error:", err.Error())
	}
}

func newAppOptions(app *App) *options.App {
	return &options.App{
		Title:  "health-tool",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnBeforeClose:    app.beforeClose,
		OnShutdown:       app.shutdown,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "health-tool",
			OnSecondInstanceLaunch: func(options.SecondInstanceData) {
				app.showMainWindow()
			},
		},
		Bind: []interface{}{
			app,
		},
	}
}

func (a *App) showMainWindow() {
	a.mu.Lock()
	ctx := a.ctx
	a.mu.Unlock()
	if ctx == nil {
		return
	}
	runtime.WindowUnminimise(ctx)
	runtime.WindowShow(ctx)
}
