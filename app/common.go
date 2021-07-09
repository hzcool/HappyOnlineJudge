package app

import (
	"github.com/gin-gonic/gin"
)

func setError(c *gin.Context, errno int, errmsg string) {
	c.Set("errno", errno)
	c.Set("errmsg", errmsg)
}

func setMap(c *gin.Context, mp map[string]interface{}) {
	for k, v := range mp {
		c.Set(k, v)
	}
}


