package main

import (
	"context"
	"embed"
	"log"

	"github.com/pulse/pulse/internal/config"
	"github.com/pulse/pulse/internal/event"
	"github.com/pulse/pulse/internal/repository/sqlite"
	"github.com/pulse/pulse/internal/service"
	"github.com/pulse/pulse/pkg/logger"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 初始化日志
	logger := logger.New()

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	logger.Info().Str("dataDir", cfg.App.DataDir).Msg("config loaded")

	// 初始化数据库
	db, err := sqlite.Open(cfg)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	if err := db.Migrate(); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
	logger.Info().Str("path", cfg.Database.SQLitePath).Msg("database ready")

	// 初始化事件总线
	bus := event.NewBus()

	// 初始化 Service 容器
	svc := service.NewContainer(db)
	logger.Info().Msg("services initialized")

	// 创建应用实例
	app := NewApp(cfg, logger, bus, db, svc)

	err = wails.Run(&options.App{
		Title:  "Pulse",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 15, G: 15, B: 15, A: 1},
		OnStartup: func(ctx context.Context) {
			app.Startup(ctx)
		},
		OnShutdown: func(ctx context.Context) {
			app.Shutdown(ctx)
		},
		Bind: []interface{}{
			app,
		},
		Mac: &mac.Options{
			TitleBar: mac.TitleBarHiddenInset(),
			Appearance: mac.NSAppearanceNameDarkAqua,
			About: &mac.AboutInfo{
				Title:   "Pulse",
				Message: "A modern open-source load testing tool.",
			},
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
		Linux: &linux.Options{
			ProgramName:         "pulse",
			WindowIsTranslucent: false,
		},
	})

	if err != nil {
		log.Fatalf("wails run failed: %v", err)
	}
}