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
	Cfg     *config.Config
	Log     *zap.Logger
	XC      *xray.Client
	Users   *service.UserService
	Admins  *service.AdminService
	Nodes   *service.NodeService
	Subs    *service.SubscriptionService
}

func NewRouter(d Deps) *gin.Engine {
	if d.Cfg.AppEnv == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8890"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Node-Token"},
		AllowCredentials: true,
	}))

	v1 := r.Group("/v1")
	{
		v1.GET("/health", Health)

		v1.POST("/admin/login", newAdminAPI(d.Admins).Login)

		// 公开订阅入口：token 即凭证，无 JWT
		subs := newSubAPI(d.Subs, d.Users)
		v1.GET("/sub/:token", subs.Public)

		// 节点 agent 心跳：X-Node-Token
		nodeAPI := newNodeAPI(d.Nodes)
		nodeAuthed := v1.Group("")
		nodeAuthed.Use(middleware.NodeAuth(d.Nodes))
		nodeAuthed.POST("/nodes/heartbeat", nodeAPI.Heartbeat)

		// 后台管理：JWT
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
			authed.GET("/users/:id/subscription", subs.AdminView)

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
