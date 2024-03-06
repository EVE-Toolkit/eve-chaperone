package main

import (
	"embed"
	"log"

	"github.com/joho/godotenv"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal(err)
	}

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "eve-chaperone",
		Width:  300,
		Height: 200,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:   app.startup,
		OnDomReady:  app.OnDomReady,
		AlwaysOnTop: true,
		Frameless:   false,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
