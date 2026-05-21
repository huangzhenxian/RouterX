package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/routex/routex/internal/service"
)

type adminAPI struct {
	svc *service.AdminService
}

func newAdminAPI(svc *service.AdminService) *adminAPI { return &adminAPI{svc: svc} }

type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (a *adminAPI) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := a.svc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrLoginFailed) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}
