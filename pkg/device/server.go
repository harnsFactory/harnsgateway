package device

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/gin-gonic/gin"
	"harnsgateway/pkg/apis"
	"harnsgateway/pkg/apis/response"
	"harnsgateway/pkg/generic"
	"harnsgateway/pkg/runtime"
	"io"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func InstallHandler(group *gin.RouterGroup, mgr *Manager) {
	group.POST("/devices", createDevice(mgr))
	group.DELETE("/devices/:id", deleteDevice(mgr))
	group.PATCH("/devices/:id", patchDeviceById(mgr))
	group.PUT("/devices/:id", updateDeviceById(mgr))
	group.GET("/devices", listDevices(mgr))
	group.GET("/devices/:id", getDeviceById(mgr))
	group.PUT("/devices/:id/:status", switchDeviceStatusById(mgr))
	group.PUT("/devices/:id/action", controlDeviceById(mgr))

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
	return func(c *gin.Context) {
		id := c.Param("id")
		eTag := c.GetHeader(apis.IfMatch)
		if len(eTag) == 0 {
			c.Status(http.StatusPreconditionRequired)
			return
		}
		device, err := mgr.DeleteDevice(id, eTag)
		if err != nil {
			if os.IsNotExist(err) {
				c.Status(http.StatusNotFound)
			} else if errors.Is(err, apis.ErrMismatch) {
				c.Status(http.StatusPreconditionFailed)
			} else {
				c.JSON(http.StatusBadRequest, response.NewMultiError(err))
			}
			return
		}
		c.JSON(http.StatusOK, device)
	}
}

func patchDeviceById(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer c.Request.Body.Close()

		contentType := c.GetHeader("Content-Type")
		// Remove "; charset=" if included in header.
		if idx := strings.Index(contentType, ";"); idx > 0 {
			contentType = contentType[:idx]
		}

		if !patchTypes.Has(contentType) {
			c.Status(http.StatusUnsupportedMediaType)
			return
		}

		eTag := c.GetHeader(apis.IfMatch)
		if len(eTag) == 0 {
			c.Status(http.StatusPreconditionRequired)
			return
		}

		pathBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			klog.V(3).InfoS("Failed to read", "err", err)
			c.Status(http.StatusInternalServerError)
			return
		}

		id := c.Param("id")
		old, err := mgr.GetDeviceById(id, true)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		versionedJS, err := json.Marshal(old)
		if err != nil {
			klog.V(3).InfoS("Failed to marshal", "err", err)
			c.Status(http.StatusInternalServerError)
			return
		}

		patchedJS, err := applyJSPatch(types.PatchType(contentType), pathBytes, versionedJS)
		if err != nil {
			c.JSONP(http.StatusBadRequest, response.NewMultiError(err))
			return
		}

		newObj := generic.DeviceTypeMap[old.GetDeviceType()]()
		if err := json.NewDecoder(bytes.NewBuffer(patchedJS)).Decode(newObj); err != nil {
			klog.V(3).InfoS("Failed to decode", "err", err)
			c.JSON(http.StatusBadRequest, response.NewMultiError(response.ErrMalformedJSON))
			return
		}

		updated, err := mgr.UpdateDeviceById(id, eTag, newObj)
		if err != nil {
			switch {
			case os.IsNotExist(err):
				c.Status(http.StatusNotFound)
			case errors.Is(err, apis.ErrMismatch):
				c.Status(http.StatusPreconditionFailed)
			default:
				if response.IsResponseError(err) {
					c.JSON(http.StatusBadRequest, response.NewMultiError(err))
				} else {
					c.Status(http.StatusInternalServerError)
				}
			}
			return
		}

		c.Header(apis.ETag, updated.GetVersion())
		c.JSON(http.StatusOK, updated)
	}
}

func updateDeviceById(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer c.Request.Body.Close()

		eTag := c.GetHeader(apis.IfMatch)
		if len(eTag) == 0 {
			c.Status(http.StatusPreconditionRequired)
			return
		}

		id := c.Param("id")
		old, err := mgr.GetDeviceById(id, true)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		newObj := generic.DeviceTypeMap[old.GetDeviceType()]()
		if err := json.NewDecoder(c.Request.Body).Decode(newObj); err != nil {
			klog.V(3).InfoS("Failed to decode", "err", err)
			c.JSON(http.StatusBadRequest, response.NewMultiError(response.ErrMalformedJSON))
			return
		}

		updated, err := mgr.UpdateDeviceById(id, eTag, newObj)
		if err != nil {
			switch {
			case os.IsNotExist(err):
				c.Status(http.StatusNotFound)
			case errors.Is(err, apis.ErrMismatch):
				c.Status(http.StatusPreconditionFailed)
			default:
				if response.IsResponseError(err) {
					c.JSON(http.StatusBadRequest, response.NewMultiError(err))
				} else {
					c.Status(http.StatusInternalServerError)
				}
			}
			return
		}

		if updated != nil {
			c.Header(apis.ETag, updated.GetVersion())
		}
		c.JSON(http.StatusOK, updated)
	}
}

func listDevices(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer c.Request.Body.Close()

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
		rds, _ := mgr.ListDevices(&filter, exploded)

		c.JSON(http.StatusOK, &runtime.ResponseModel{Devices: rds})
	}
}

func getDeviceById(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer c.Request.Body.Close()

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

func switchDeviceStatusById(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer c.Request.Body.Close()

		id := c.Param("id")
		status := c.Param("status")
		if err := mgr.SwitchDeviceStatus(id, status); err != nil {
			if os.IsNotExist(err) {
				c.Status(http.StatusNotFound)
			} else {
				c.JSON(http.StatusBadRequest, response.NewMultiError(err))
			}
		}
		c.Status(http.StatusAccepted)
	}
}

func controlDeviceById(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer c.Request.Body.Close()

		id := c.Param("id")
		var actions []map[string]interface{}
		if err := json.NewDecoder(c.Request.Body).Decode(&actions); err != nil {
			klog.V(3).InfoS("Failed to parse action", "err", err)
			c.JSON(http.StatusBadRequest, response.NewMultiError(response.ErrMalformedJSON))
			return
		}

		err := mgr.DeliverAction(id, actions)

		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}

		c.Status(http.StatusAccepted)
	}
}

func applyJSPatch(patchType types.PatchType, patchBytes, versionedJS []byte) (patchedJS []byte, err error) {
	switch patchType {
	case types.JSONPatchType:
		patchObj, err := jsonpatch.DecodePatch(patchBytes)
		if err != nil {
			return nil, response.ErrMalformedJSON
		}
		if len(patchObj) > maxJSONPatchOperations {
			klog.V(3).InfoS("Too many json patch operations", "count", len(patchObj))
			return nil, response.ErrTooManyJsonPatchOperations(maxJSONPatchOperations)
		}
		patchedJS, err := patchObj.Apply(versionedJS)
		if err != nil {
			klog.V(3).InfoS("Failed to apply json patch", "err", err)
			return nil, response.ErrMalformedJSON
		}
		return patchedJS, nil
	case types.MergePatchType:
		patchedJS, err = jsonpatch.MergePatch(versionedJS, patchBytes)
		if err != nil {
			klog.V(3).InfoS("Failed to apply json merge patch", "err", err)
			return nil, response.ErrMalformedJSON
		}
		return patchedJS, err
	default:
		// only here as a safety net - gin filters content-type
		return nil, fmt.Errorf("unknown Content-Type header for patch: %v", patchType)
	}
}
