package main

import (
	"context"
	"embed"
	"fmt"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/zyriu/portfolio/backend"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {

	m := backend.NewManager()

	app := &options.App{
		Title:  "Portfolio",
		Width:  1600,
		Height: 900,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Bind: []any{m},
		// Disable focus grabbing during job execution
		DisableResize: false,
		Fullscreen:    false,
		OnStartup: func(ctx context.Context) {
			m.Startup(ctx)

			// Sync jobs with current settings (stops disabled jobs, starts enabled ones)
			if err := m.SyncJobsWithSettings(); err != nil {
				fmt.Printf("Warning: Failed to sync jobs with settings: %v\n", err)
			}
		},
	}

	if err := wails.Run(app); err != nil {
		log.Fatal(err)
	}
}
