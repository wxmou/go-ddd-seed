// @title          Go DDD Seed API
// @version        1.0
// @description    通用 DDD 脚手架后端 API
// @termsOfService http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    https://github.com/go-ddd-seed/go-ddd-seed
// @contact.email  team@example.com

// @license.name   Apache 2.0
// @license.url    http://www.apache.org/licenses/LICENSE-2.0.html

// @host           localhost:8080
// @BasePath       /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in                         header
// @name                       Authorization
// @description               Bearer token authentication. Format: "Bearer <token>"
package main

import (
	"fmt"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/config"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/logger"
)

func main() {
	// 加载配置
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		logger.Fatal("配置加载失败", err)
	}

	// 初始化依赖（通过 Wire）
	app, cleanup, err := InitializeApp(cfg)
	if err != nil {
		logger.Fatal("应用初始化失败", err)
	}
	defer cleanup()

	// 启动服务
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	app.Logger.Info("Go DDD Seed API 服务启动", map[string]interface{}{
		"addr": addr,
		"mode": cfg.Server.Mode,
	})
	if err := app.Run(addr); err != nil {
		app.Logger.Fatal("服务启动失败", err)
	}
}
