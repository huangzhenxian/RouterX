package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/routex/routex/internal/config"
	"github.com/routex/routex/internal/xray"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func NewRouter(cfg *config.Config, zlog *zap.Logger, db *gorm.DB, xc *xray.Client) *gin.Engine {
	if cfg.AppEnv == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	v1 := r.Group("/v1")
	{
		v1.GET("/health", Health)

		// 仅开发环境暴露的调试接口，用来手测 Xray gRPC 是否通。
		// 生产前需要加鉴权或移除。
		if cfg.AppEnv != "prod" {
			dbg := newDebugXray(xc)
			grp := v1.Group("/_debug/xray")
			grp.GET("/ping", dbg.Ping)
			grp.POST("/users", dbg.AddUser)
			grp.DELETE("/users/:email", dbg.RemoveUser)
			grp.GET("/users/:email/traffic", dbg.UserTraffic)
			grp.GET("/inbound/traffic", dbg.InboundTraffic)
		}
	}

	_ = db // 后续用户/节点 API 会用到，先留住引用避免 unused
	return r
}
