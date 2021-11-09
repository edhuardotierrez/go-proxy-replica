package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

func main() {

	bindFlag := flag.String("bind", "", "bind address (ex: -bind 127.0.0.1:8080")
	flag.Parse()

	var bindAddress = "127.0.0.1:8080"

	if len(*bindFlag) > 0 {
		bindAddress = strings.TrimSpace(*bindFlag)
	}

	route := gin.Default()
	route.NoRoute(func(c *gin.Context) {
		log.Println(c.Request.Method, c.Request.URL)
		c.String(http.StatusOK, "ok")
	})

	if err := route.Run(bindAddress); err != nil {
		log.Errorf("Failed to run server: %v", err)
	}

}
