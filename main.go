package main

import (
	"context"
	_ "embed"
	"embed"
	"os"

	"easywinget/backend/system"

	"github.com/energye/systray"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/windows/icon.ico
var trayIcon []byte

//go:embed build/appicon.png
var appIconPng []byte


func main() {
	startHidden := false
	for _, arg := range os.Args[1:] {
		if arg == "-silent" || arg == "--silent" {
			startHidden = true
			break
		}
	}

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:             "EasyWinGet",
		Width:             1180,
		Height:            780,
		MinWidth:          960,
		MinHeight:         640,
		StartHidden:       startHidden,
		HideWindowOnClose: true,
		Frameless:         true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 32, G: 32, B: 32, A: 0},
		OnStartup: func(ctx context.Context) {
			system.InitAppIcon(appIconPng)
			app.startup(ctx)
			initTray(app, ctx)
		},

		OnShutdown: func(ctx context.Context) {
			systray.Quit()
		},
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			BackdropType:         windows.Mica,
			Theme:                windows.Dark,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func initTray(app *App, ctx context.Context) {
	go func() {
		systray.Run(func() {
			if len(trayIcon) > 0 {
				systray.SetIcon(trayIcon)
			}
			systray.SetTooltip("EasyWinGet - Windows Package Manager")

			mOpen := systray.AddMenuItem("Open EasyWinGet", "Show EasyWinGet window")
			mCheck := systray.AddMenuItem("Check for Updates", "Scan for application updates")
			systray.AddSeparator()
			mQuit := systray.AddMenuItem("Exit EasyWinGet", "Quit application completely")

			systray.SetOnClick(func(menu systray.IMenu) {
				wailsRuntime.WindowShow(ctx)
				wailsRuntime.WindowUnminimise(ctx)
			})

			systray.SetOnDClick(func(menu systray.IMenu) {
				wailsRuntime.WindowShow(ctx)
				wailsRuntime.WindowUnminimise(ctx)
			})

			mOpen.Click(func() {
				wailsRuntime.WindowShow(ctx)
				wailsRuntime.WindowUnminimise(ctx)
			})

			mCheck.Click(func() {
				wailsRuntime.WindowShow(ctx)
				wailsRuntime.WindowUnminimise(ctx)
				app.CheckForUpdatesStartup()
			})

			mQuit.Click(func() {
				systray.Quit()
				wailsRuntime.Quit(ctx)
			})
		}, func() {
			// OnExit
		})
	}()
}
