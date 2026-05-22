package api

import (
	"embed"
	"encoding/json"
	"html/template"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/routex/routex/internal/service"
)

//go:embed templates/sub.html
var subFS embed.FS

var subTmpl = template.Must(template.ParseFS(subFS, "templates/sub.html"))

// subLink 是模板里渲染每条 vless 链接的视图模型。
// URLJSON 是 url 字符串的 JSON 安全编码（含引号），在 onclick={{.URLJSON}} 直接落到属性里不会被引号截断。
type subLink struct {
	NodeName string
	Host     string
	Port     int
	URL      string
	URLJSON  template.JS
}

type subView struct {
	Username string
	SubURL   string
	Base64   string
	Links    []subLink
}

// renderSubHTML 把订阅数据塞进 HTML 模板。
// 调用方负责判断 Accept 头是不是 text/html。
func (h *subAPI) renderSubHTML(c *gin.Context, payload *service.Payload, subURL string) {
	links := make([]subLink, 0, len(payload.Links))
	for _, l := range payload.Links {
		b, _ := json.Marshal(l.URL)
		links = append(links, subLink{
			NodeName: l.NodeName,
			Host:     l.Host,
			Port:     l.Port,
			URL:      l.URL,
			URLJSON:  template.JS(b),
		})
	}
	view := subView{
		Username: payload.Username,
		SubURL:   subURL,
		Base64:   payload.Base64,
		Links:    links,
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := subTmpl.Execute(c.Writer, view); err != nil {
		c.String(http.StatusInternalServerError, "render: "+err.Error())
	}
}

func wantsHTML(c *gin.Context) bool {
	accept := c.GetHeader("Accept")
	// 浏览器一般会发 text/html,application/xhtml+xml,...
	// 代理客户端通常 Accept: */* 或 text/plain
	return strings.Contains(accept, "text/html")
}
