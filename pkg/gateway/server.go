package gateway

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"harnsgateway/pkg/apis"
	"net/http"
)

func InstallHandler(group *gin.RouterGroup, mgr *Manager) {
	group.GET("/gatewayMeta", getGateway(mgr))

}

func getGateway(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		g, _ := mgr.GetGatewayMeta()
		c.Header(apis.ETag, fmt.Sprintf("%s", g.GetVersion()))
		c.JSON(http.StatusOK, g)
	}
}
