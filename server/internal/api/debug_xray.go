package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/routex/routex/internal/xray"
)

// debugXray 是一组手测 Xray gRPC 联通的接口，仅在 dev 启用。
type debugXray struct {
	xc *xray.Client
}

func newDebugXray(xc *xray.Client) *debugXray { return &debugXray{xc: xc} }

func (d *debugXray) Ping(c *gin.Context) {
	if err := d.xc.Ping(c.Request.Context()); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":      "ok",
		"inbound_tag": d.xc.InboundTag(),
	})
}

type addUserReq struct {
	Email    string `json:"email"    binding:"required"`
	UUID     string `json:"uuid"     binding:"required"`
	Flow     string `json:"flow"`
	Protocol string `json:"protocol"` // 默认 vless
	Level    uint32 `json:"level"`
}

func (d *debugXray) AddUser(c *gin.Context) {
	var req addUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := d.xc.AddUser(c.Request.Context(), xray.UserSpec{
		Email:    req.Email,
		UUID:     req.UUID,
		Flow:     req.Flow,
		Protocol: xray.Protocol(req.Protocol),
		Level:    req.Level,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (d *debugXray) RemoveUser(c *gin.Context) {
	email := c.Param("email")
	if err := d.xc.RemoveUser(c.Request.Context(), email); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (d *debugXray) UserTraffic(c *gin.Context) {
	email := c.Param("email")
	reset := c.Query("reset") == "true"
	stat, err := d.xc.QueryUserTraffic(c.Request.Context(), email, reset)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stat)
}

func (d *debugXray) InboundTraffic(c *gin.Context) {
	tag := c.DefaultQuery("tag", d.xc.InboundTag())
	reset := c.Query("reset") == "true"
	stat, err := d.xc.QueryInboundTraffic(c.Request.Context(), tag, reset)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tag": tag, "stat": stat})
}
