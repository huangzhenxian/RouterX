package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/routex/routex/internal/service"
	"gorm.io/gorm"
)

type providerAPI struct {
	svc *service.ProviderService
}

func newProviderAPI(svc *service.ProviderService) *providerAPI { return &providerAPI{svc: svc} }

func (h *providerAPI) Create(c *gin.Context) {
	var in service.CreateProviderInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *providerAPI) List(c *gin.Context) {
	ps, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": ps, "total": len(ps)})
}

func (h *providerAPI) Get(c *gin.Context) {
	id, err := parseProviderID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *providerAPI) Delete(c *gin.Context) {
	id, err := parseProviderID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *providerAPI) Enable(c *gin.Context) {
	id, err := parseProviderID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Enable(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *providerAPI) Disable(c *gin.Context) {
	id, err := parseProviderID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Disable(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Test 立即跑一次健康检查（同步，等返回）。
// 前端 "测试" 按钮触发：响应里直接给延迟/错误，方便用户判断。
func (h *providerAPI) Test(c *gin.Context) {
	id, err := parseProviderID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	res := h.svc.Check(c.Request.Context(), p)
	// 不走 autoDisable 阈值：手动测试只关心当下，不应触发禁用
	_ = h.svc.PersistCheck(c.Request.Context(), p, res, 0)
	c.JSON(http.StatusOK, gin.H{
		"healthy":    res.Healthy,
		"latency_ms": res.LatencyMs,
		"error":      res.Err,
	})
}

func parseProviderID(c *gin.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}
