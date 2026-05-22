package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/routex/routex/internal/service"
	"gorm.io/gorm"
)

// routerAPI 是控制 vless-in → proxy-out 实际指向的接口，
// 不是 gin 的路由对象（那个在 router.go 的 NewRouter 里）。
type routerAPI struct {
	rs *service.RouterService
	ps *service.ProviderService
}

func newRouterAPI(rs *service.RouterService, ps *service.ProviderService) *routerAPI {
	return &routerAPI{rs: rs, ps: ps}
}

// Active 返回当前 proxy-out 实际指向的 provider 详情；为 0 时表示直连兜底。
func (h *routerAPI) Active(c *gin.Context) {
	id := h.rs.CurrentProviderID()
	if id == 0 {
		c.JSON(http.StatusOK, gin.H{
			"active":   false,
			"provider": nil,
			"mode":     "direct", // freedom 兜底
		})
		return
	}
	p, err := h.ps.Get(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{"active": true, "provider": nil, "mode": "stale"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"active":   true,
		"provider": p,
		"mode":     "provider",
	})
}

// Sync 手动触发一次最优 provider 重选 + Xray 切换，用于"我刚改了配置想立刻生效"场景。
func (h *routerAPI) Sync(c *gin.Context) {
	if err := h.rs.SyncBest(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "active_provider_id": h.rs.CurrentProviderID()})
}
