package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"github.com/amirreza-azadmehr/dlyven/internal/appinfo"
	"github.com/amirreza-azadmehr/dlyven/internal/logx"
	"github.com/amirreza-azadmehr/dlyven/internal/settings"
	"github.com/amirreza-azadmehr/dlyven/internal/sysutil"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

func main() {
	dataDir, err := appinfo.DataDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "data dir:", err)
		os.Exit(1)
	}
	log, logFile, err := logx.New(filepath.Join(dataDir, "logs"), appinfo.ID, os.Getenv("DLYVEN_DEBUG") != "")
	if err != nil {
		fmt.Fprintln(os.Stderr, "log:", err)
	}
	st, err := settings.Load(filepath.Join(dataDir, "settings.json"))
	if err != nil {
		log.Error("settings load failed, using defaults", "err", err)
	}
	log.Info("starting", "version", appinfo.Version, "os", runtime.GOOS, "arch", runtime.GOARCH)

	// Translucency (Mica on Windows 11, vibrancy on macOS) needs a transparent window;
	// everywhere else we render a solid background.
	translucent := st.Get().Translucency && sysutil.SupportsTranslucency()

	app, err := NewApp(dataDir, st, log, logFile, translucent)
	if err != nil {
		log.Error("init failed", "err", err)
		fmt.Fprintln(os.Stderr, "init:", err)
		os.Exit(1)
	}

	bg := &options.RGBA{R: 11, G: 13, B: 18, A: 255}
	if translucent {
		bg = &options.RGBA{R: 0, G: 0, B: 0, A: 0}
	}

	err = wails.Run(&options.App{
		Title:     appinfo.Name,
		Width:     1180,
		Height:    780,
		MinWidth:  900,
		MinHeight: 600,
		// Custom titlebar on Windows/Linux; macOS keeps native traffic lights (hidden-inset).
		Frameless:        runtime.GOOS != "darwin",
		BackgroundColour: bg,
		AssetServer:      &assetserver.Options{Assets: assets},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind:             []interface{}{app},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: appinfo.BundleID,
			OnSecondInstanceLaunch: func(options.SecondInstanceData) {
				app.focusWindow()
			},
		},
		DragAndDrop: &options.DragAndDrop{
			// Links and .txt files are handled by the web layer (HTML5 drag & drop).
			EnableFileDrop:     false,
			DisableWebViewDrop: false,
		},
		Windows: &windows.Options{
			WebviewIsTransparent:              translucent,
			WindowIsTranslucent:               translucent,
			BackdropType:                      windows.Mica,
			Theme:                             windows.SystemDefault,
			DisableFramelessWindowDecorations: false,
		},
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarHiddenInset(),
			WebviewIsTransparent: translucent,
			WindowIsTranslucent:  translucent,
			About: &mac.AboutInfo{
				Title:   appinfo.Name,
				Message: "Version " + appinfo.Version,
				Icon:    icon,
			},
		},
		Linux: &linux.Options{
			Icon:             icon,
			ProgramName:      appinfo.ID,
			WebviewGpuPolicy: linux.WebviewGpuPolicyOnDemand,
		},
	})
	if err != nil {
		log.Error("wails run failed", "err", err)
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
