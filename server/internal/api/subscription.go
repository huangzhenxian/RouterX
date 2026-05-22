package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/routex/routex/internal/service"
	"gorm.io/gorm"
)

type subAPI struct {
	subs  *service.SubscriptionService
	users *service.UserService
}

func newSubAPI(subs *service.SubscriptionService, users *service.UserService) *subAPI {
	return &subAPI{subs: subs, users: users}
}

// Public 订阅入口：token 本身就是凭证，无需 JWT。
//
// 内容协商：
//   - 浏览器（Accept: text/html）→ 渲染友好页面，显示二维码、客户端下载、复制按钮
//   - 代理客户端（Accept: */* 或 text/plain）→ 返回 base64 编码的 vless 链接列表
func (h *subAPI) Public(c *gin.Context) {
	token := c.Param("token")
	user, err := h.subs.FindByToken(c.Request.Context(), token)
	if err != nil {
		if errors.Is(err, service.ErrSubscriptionNotFound) {
			c.String(http.StatusNotFound, "subscription not found")
			return
		}
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if user.Status != 1 {
		c.String(http.StatusForbidden, "subscription disabled")
		return
	}
	payload, err := h.subs.BuildForUser(c.Request.Context(), user)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if wantsHTML(c) {
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		subURL := scheme + "://" + c.Request.Host + "/v1/sub/" + token
		h.renderSubHTML(c, payload, subURL)
		return
	}

	// 主流代理客户端识别 base64-encoded text/plain 订阅
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Profile-Update-Interval", "24")
	c.String(http.StatusOK, payload.Base64)
}

// AdminView 后台用：返回结构化数据（含明文 URL 列表 + base64），方便展示二维码。
func (h *subAPI) AdminView(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.users.Get(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	payload, err := h.subs.BuildForUser(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// 把订阅 endpoint 的绝对路径也一并返回，方便前端展示 / 生成二维码
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	subURL := scheme + "://" + c.Request.Host + "/v1/sub/" + user.SubscriptionToken
	c.JSON(http.StatusOK, gin.H{
		"user":             user,
		"subscription_url": subURL,
		"payload":          payload,
	})
}
