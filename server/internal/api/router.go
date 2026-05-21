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

type Deps struct {
	Cfg    *config.Config
	Log    *zap.Logger
	XC     *xray.Client
	Users  *service.UserService
	Admins *service.AdminService
	Nodes  *service.NodeService
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
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Node-Token"},
		AllowCredentials: true,
	}))

	v1 := r.Group("/v1")
	{
		v1.GET("/health", Health)

		// 公开：管理员登录
		v1.POST("/admin/login", newAdminAPI(d.Admins).Login)

		// 节点 agent 心跳：用 X-Node-Token 鉴权，不走 JWT
		nodeAPI := newNodeAPI(d.Nodes)
		nodeAuthed := v1.Group("")
		nodeAuthed.Use(middleware.NodeAuth(d.Nodes))
		nodeAuthed.POST("/nodes/heartbeat", nodeAPI.Heartbeat)

		// 后台管理：JWT 鉴权
		authed := v1.Group("")
		authed.Use(middleware.JWTAuth(d.Cfg.JWTSecret))
		{
			u := newUserAPI(d.Users)
			authed.POST("/users", u.Create)
			authed.GET("/users", u.List)
			authed.GET("/users/:id", u.Get)
			authed.DELETE("/users/:id", u.Delete)
			authed.POST("/users/:id/enable", u.Enable)
			authed.POST("/users/:id/disable", u.Disable)

			authed.POST("/nodes", nodeAPI.Create)
			authed.GET("/nodes", nodeAPI.List)
			authed.GET("/nodes/:id", nodeAPI.Get)
			authed.DELETE("/nodes/:id", nodeAPI.Delete)
		}

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
