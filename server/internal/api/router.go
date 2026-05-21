package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/routex/routex/internal/config"
	"github.com/routex/routex/internal/middleware"
	"github.com/routex/routex/internal/service"
	"github.com/routex/routex/internal/xray"
	"go.uber.org/zap"
)

// Deps 路由层依赖的服务集合。比把零散参数串到 NewRouter 里清爽。
type Deps struct {
	Cfg    *config.Config
	Log    *zap.Logger
	XC     *xray.Client
	Users  *service.UserService
	Admins *service.AdminService
}

func NewRouter(d Deps) *gin.Engine {
	if d.Cfg.AppEnv == "prod" {
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

		// 公开接口：登录
		adminAPI := newAdminAPI(d.Admins)
		v1.POST("/admin/login", adminAPI.Login)

		// 受保护：JWT 鉴权
		auth := v1.Group("")
		auth.Use(middleware.JWTAuth(d.Cfg.JWTSecret))
		{
			userAPI := newUserAPI(d.Users)
			auth.POST("/users", userAPI.Create)
			auth.GET("/users", userAPI.List)
			auth.GET("/users/:id", userAPI.Get)
			auth.DELETE("/users/:id", userAPI.Delete)
			auth.POST("/users/:id/enable", userAPI.Enable)
			auth.POST("/users/:id/disable", userAPI.Disable)
		}

		// dev 环境的 Xray 联通调试接口，无鉴权
		if d.Cfg.AppEnv != "prod" {
			dbg := newDebugXray(d.XC)
			grp := v1.Group("/_debug/xray")
			grp.GET("/ping", dbg.Ping)
			grp.POST("/users", dbg.AddUser)
			grp.DELETE("/users/:email", dbg.RemoveUser)
			grp.GET("/users/:email/traffic", dbg.UserTraffic)
			grp.GET("/inbound/traffic", dbg.InboundTraffic)
		}
	}

	return r
}
