package main

import (
	"embed"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

// version is set at build time via -ldflags "-X main.version=<value>"
var version = "dev"

// helperPin is the SHA-256 (hex) of the packaged bundle helper, injected at
// build time via -ldflags "-X main.helperPin=<sha256>" before the main binary
// is linked. It is empty in local development builds. The pin gates only the
// launch of a bundle helper copy (first install / legitimate update); it is
// never a precondition for launching an already-verified root-owned protected
// copy. It is defense-in-depth, not the trust anchor — the anchor is the
// root-owned protected copy, whose availability does not depend on this pin.
var helperPin string

//go:embed all:frontend/.output/public
var assets embed.FS

func main() {
	app := NewAppWithVersion(version)

	// Create application menu
	appMenu := menu.NewMenu()
	if runtime.GOOS == "darwin" {
		appMenu.Append(menu.AppMenu())
		appMenu.Append(menu.EditMenu())
	}

	err := wails.Run(&options.App{
		Title:     "LaunchPal",
		Width:     1024,
		Height:    768,
		MinWidth:  800,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 30, G: 30, B: 30, A: 1},
		OnStartup:        app.startup,
		Menu:             appMenu,
		Bind: []interface{}{
			app,
		},
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                  true,
				HideTitleBar:               false,
				FullSizeContent:            true,
				UseToolbar:                 false,
			},
			Appearance: mac.NSAppearanceNameDarkAqua,
			About: &mac.AboutInfo{
				Title:   "LaunchPal",
				Message: "LaunchAgent Manager for macOS\n© 2025",
			},
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
