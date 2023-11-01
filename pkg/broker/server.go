package broker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"harnsgateway/pkg/apis"
	"harnsgateway/pkg/apis/response"
	"harnsgateway/pkg/generic"
	"harnsgateway/pkg/runtime"
	"io"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"strconv"
)

func InstallHandler(group *gin.RouterGroup, mgr *Manager) {
	group.POST("/devices", createDevice(mgr))
	group.DELETE("/devices/:id", deleteDevice(mgr))
	group.GET("/devices", listDevices(mgr))
	group.GET("/devices/:id", getDeviceById(mgr))
	group.PUT("/devices/:id/control", controlDeviceById(mgr))
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

func listDevices(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Request.URL.Query()
		exploded := false
		filter := runtime.DeviceFilter{}
		if len(query) > 0 {
			v := query.Get(apis.Filter)
			if len(v) > 0 {
				if err := json.Unmarshal([]byte(v), &filter); err != nil {
					c.JSON(http.StatusBadRequest, response.NewMultiError(response.ErrMalformedJSON))
					return
				}
			}
			exploded, _ = strconv.ParseBool(query.Get("exploded"))
		}
		rds, _ := mgr.listDevices(&filter, exploded)

		c.JSON(http.StatusOK, &runtime.ResponseModel{Devices: rds})
	}
}

func getDeviceById(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		query := c.Request.URL.Query()
		exploded := false
		if len(query) > 0 {
			exploded, _ = strconv.ParseBool(query.Get("exploded"))
		}
		rd, err := mgr.GetDeviceById(id, exploded)
		if err != nil {
			if os.IsNotExist(err) {
				c.Status(http.StatusNotFound)
			} else {
				c.Status(http.StatusInternalServerError)
			}
			return
		}

		c.Header(apis.ETag, fmt.Sprintf("%s", rd.GetVersion()))
		c.JSON(http.StatusOK, rd)
	}
}

func controlDeviceById(mgr *Manager) gin.HandlerFunc {
	return nil
}
