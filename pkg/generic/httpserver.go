package generic

import "github.com/gin-gonic/gin"

func Default() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(logger(), gin.Recovery())
	return engine
}

func logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		// start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Stop timer
		// latency := time.Now().Sub(start)
		if raw != "" {
			path = path + "?" + raw
		}

		// klog.V(4).InfoS("Received HTTP request",
		// 	"verb", c.Request.Method,
		// 	"URI", path,
		// 	"status", c.Writer.Status(),
		// 	// "latency", latency,
		// )
	}
}
