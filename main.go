package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"forge-api-client/app"
	"forge-api-client/internal/security"
	"forge-api-client/internal/storage"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed migrations/*.sql
var migrations embed.FS

func main() {
	configRoot, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	// Keep the legacy directory so existing users retain requests, environments and history after the rename.
	configDir := filepath.Join(configRoot, "ForgeAPIClient")
	if err = os.MkdirAll(configDir, 0700); err != nil {
		panic(err)
	}
	store, err := storage.Open(filepath.Join(configDir, "forge.db"), migrations)
	if err != nil {
		panic(err)
	}
	protector, err := security.NewProtector(configDir)
	if err != nil {
		panic(err)
	}
	frontend, err := fs.Sub(assets, "frontend/dist")
	if err != nil {
		panic(err)
	}
	application := app.New(store, protector)
	err = wails.Run(&options.App{Title: "Metriyon API", Width: 1440, Height: 900, MinWidth: 960, MinHeight: 640, Frameless: true, AssetServer: &assetserver.Options{Assets: frontend}, BackgroundColour: &options.RGBA{R: 10, G: 13, B: 18, A: 1}, OnStartup: application.Startup, OnShutdown: application.Shutdown, Bind: []interface{}{application}, Windows: &windows.Options{WebviewIsTransparent: false, WindowIsTranslucent: false, Theme: windows.Dark}})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Metriyon API:", err)
		os.Exit(1)
	}
}
