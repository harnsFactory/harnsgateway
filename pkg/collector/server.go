package collector

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"harnsgateway/pkg/apis"
	"harnsgateway/pkg/apis/response"
	"harnsgateway/pkg/generic"
	"io"
	"k8s.io/klog/v2"
	"net/http"
	"os"
)

func InstallHandler(group *gin.RouterGroup, mgr *Manager) {
	group.POST("/devices", createDevice(mgr))
	group.DELETE("/devices/:id", deleteDevice(mgr))
}

func createDevice(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			klog.V(2).InfoS("Failed to get request body", "err", err)
			c.JSON(http.StatusBadRequest, response.NewMultiError(response.ErrMalformedJSON))
		}

		var target struct {
			DeviceType string `json:"deviceType"`
		}
		err = json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&target)
		if err != nil {
			klog.V(2).InfoS("Failed to parse device type", "err", err)
			c.JSON(http.StatusBadRequest, response.NewMultiError(response.ErrRequestBody))
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		object := generic.DeviceTypeMap[target.DeviceType]()
		if err := c.ShouldBindJSON(object); err != nil {
			klog.V(2).InfoS("Failed to parse Device", "err", err)
			c.JSON(http.StatusBadRequest, response.NewMultiError(response.ErrMalformedJSON))
			return
		}
		d, err := mgr.CreateDevice(object)

		if err != nil {
			c.JSON(http.StatusBadRequest, response.NewMultiError(err))
			return
		}

		// TODO use different scheme
		c.Header(apis.ETag, fmt.Sprintf("%s", d.GetVersion()))
		c.Header(apis.Location, fmt.Sprintf("https://%s%s/%s", c.Request.Host, c.Request.RequestURI, d.GetID()))
		c.JSON(http.StatusCreated, d)
	}
}

func deleteDevice(mgr *Manager) gin.HandlerFunc {
	return func(context *gin.Context) {
		id := context.Param("id")
		eTag := context.GetHeader(apis.IfMatch)
		if len(eTag) == 0 {
			context.Status(http.StatusPreconditionRequired)
			return
		}
		device, err := mgr.deleteDevice(id, eTag)
		if err != nil {
			if os.IsNotExist(err) {
				context.Status(http.StatusNotFound)
			} else if errors.Is(err, apis.ErrMismatch) {
				context.Status(http.StatusPreconditionFailed)
			} else {
				context.JSON(http.StatusBadRequest, response.NewMultiError(err))
			}
			return
		}
		context.JSON(http.StatusOK, device)
	}
}
