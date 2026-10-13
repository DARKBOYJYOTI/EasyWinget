package main

import (
	"context"
	_ "embed"
	"embed"
	"os"
	"strings"

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
	// Enable cross-integrity communication so standard-user taskbar shortcut clicks can wake up an elevated instance
	system.AllowUIPIMessages()

	startHidden := false
	initialTab := ""
	for _, arg := range os.Args[1:] {
		low := strings.ToLower(arg)
		if low == "-silent" || low == "--silent" {
			startHidden = true
		} else if strings.Contains(low, "updates") {
			initialTab = "updates"
		}
	}

	// Create an instance of the app structure
	app := NewApp()
	if initialTab != "" {
		app.SetInitialTab(initialTab)
	}

	// Create application with options
	err := wails.Run(&options.App{
		Title:             "EasyWinGet",
		Width:             1060,
		Height:            640,
		MinWidth:          880,
		MinHeight:         520,
		StartHidden:       startHidden,
		HideWindowOnClose: true,
		Frameless:         true,
		WindowStartState:  options.Maximised,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 32, G: 32, B: 32, A: 0},
		OnStartup: func(ctx context.Context) {
			system.InitAppIcon(appIconPng)
			system.CleanupProgramFilesJunk()
			app.startup(ctx)
			initTray(app, ctx)
		},

		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "easywinget-windows-native-single-instance-lock-guid",
			OnSecondInstanceLaunch: func(secondInstanceData options.SecondInstanceData) {
				app.RestoreWindow()
				for _, arg := range secondInstanceData.Args {
					if strings.Contains(strings.ToLower(arg), "updates") {
						app.NavigateToTab("updates")
						break
					}
				}
			},
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
			Theme:                windows.SystemDefault,
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
				app.RestoreWindow()
			})

			systray.SetOnDClick(func(menu systray.IMenu) {
				app.RestoreWindow()
			})

			mOpen.Click(func() {
				app.RestoreWindow()
			})

			mCheck.Click(func() {
				app.RestoreWindow()
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
