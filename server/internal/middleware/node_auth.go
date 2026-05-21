package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/routex/routex/internal/service"
)

const (
	CtxNodeID   = "node_id"
	CtxNodeName = "node_name"
)

// NodeAuth 校验 X-Node-Token 头部，匹配上一条 nodes.auth_token 才放行，
// 并把 node_id / node_name 注入到 gin.Context 供 handler 取用。
func NodeAuth(nodes *service.NodeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-Node-Token")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing X-Node-Token"})
			return
		}
		n, err := nodes.FindByToken(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid node token"})
			return
		}
		c.Set(CtxNodeID, n.ID)
		c.Set(CtxNodeName, n.Name)
		c.Next()
	}
}
