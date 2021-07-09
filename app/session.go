package app

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	USERNAME_KEY   = "who"
	ID_KEY         = "which"
	SESSION_EXPIRE = 3600 * 24 * 3 //
)

//根据session 验证是否登陆
func isLogin(c *gin.Context) bool {
	return getUserID(c) != 0
}

func getUserID(c *gin.Context) int64 {
	session := sessions.Default(c)
	id := session.Get(ID_KEY)
	if id == nil {
		return 0
	}
	return id.(int64)
}

//根据session获取用户名
func getUserName(c *gin.Context) string {
	session := sessions.Default(c)
	username := session.Get(USERNAME_KEY)
	if username == nil {
		return ""
	}
	return username.(string)
}

//设置session
func setSession(c *gin.Context, val string, id int64) {
	session := sessions.Default(c)
	session.Set(USERNAME_KEY, val)
	session.Set(ID_KEY, id)
	session.Save()
}

//删除session
func deleteSession(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete(USERNAME_KEY)
	session.Delete(ID_KEY)
	session.Save()
}
