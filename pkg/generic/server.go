package generic

import "github.com/gin-gonic/gin"

type Server struct {
	Router  *gin.Engine
	Port    string
	Methods []string
}
