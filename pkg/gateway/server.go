package gateway

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"harnsgateway/pkg/apis"
	"net/http"
)

func InstallHandler(group *gin.RouterGroup, mgr *Manager) {
	group.GET("/gatewayMeta", getGatewayMeta(mgr))
	group.GET("/gatewayCpu", getGatewayCpu(mgr))
	group.GET("/gatewayMem", getGatewayMem(mgr))
	group.GET("/gatewayDisk", getGatewayDisk(mgr))
}

func getGatewayMeta(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		g, _ := mgr.GetGatewayMeta()
		c.Header(apis.ETag, fmt.Sprintf("%s", g.GetVersion()))
		c.JSON(http.StatusOK, g)
	}
}

func getGatewayCpu(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		cpu, err := mgr.getGatewayCpu()
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, ResponseModel{Cpus: cpu})
	}
}

func getGatewayMem(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		mem, err := mgr.getGatewayMem()
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, ResponseModel{Mem: mem})
	}
}

func getGatewayDisk(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		disks, err := mgr.getGatewayDisk()
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, ResponseModel{Disks: disks})
	}
}
