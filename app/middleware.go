package app

import (
	"HappyOnlineJudge/common"
	"HappyOnlineJudge/dao"
	"github.com/gin-gonic/gin"
)

//中间件

//验证是否登陆
func AuthLogin(c *gin.Context) {
	if !isLogin(c) {
		setError(c, 401, "未登陆")
		c.Abort()
	}
}

//管理员验证
func AuthAdmin(c *gin.Context) {
	id := getUserID(c)
	if id != 0 {
		ud := &dao.UserDao{ID: id}
		if !dao.OneCol(ud, "is_admin").ToBool() {
			setError(c, 403, "没有权限")
			c.Abort()
		}
	} else {
		setError(c, 401, "未登陆")
		c.Abort()
	}
}

func AuthSuperAdmin(c *gin.Context) {
	id := getUserID(c)
	if id != 0 {
		ud := &dao.UserDao{ID: id}
		if !dao.OneCol(ud, "is_super_admin").ToBool() {
			setError(c, 403, "没有权限")
			c.Abort()
		}
	} else {
		setError(c, 401, "未登陆")
		c.Abort()
	}
}

//c中没有返回码, 默认为200,
func jsonResponse(c *gin.Context) {
	c.Next()
	statusCode := c.Writer.Status()
	if statusCode == 404 {
		c.JSON(404, gin.H{"errmsg": "Not Found"})
	} else if _, exist := c.Get("noPack"); !exist {
		if username := getUserName(c); username != "" {
			c.Set("message_count", dao.MessageCount(username))
		}
		delete(c.Keys, "github.com/gin-contrib/sessions")
		c.JSON(200, c.Keys)
	}
}
func IsSuperAdmin(c *gin.Context) bool {
	return getUserID(c) == 1
}

//管理员权限验证
func __ScanProblemSubmission(c *gin.Context) {
	ud := &dao.UserDao{ID: getUserID(c)}
	if ud.GetID() != 1 && (dao.OneCol(ud, "privilege").ToUint64()&ScanProblemSubmission == 0) {
		setError(c, 403, "没有权限")
		c.Abort()
	}
}

func __ScanPrivateProblem(c *gin.Context) {
	isOpen := common.StrToBool(c.DefaultQuery("is_open", "false"))
	ud := &dao.UserDao{ID: getUserID(c)}
	if !dao.OneCol(ud, "is_admin").ToBool() {
		setError(c, 403, "没有权限")
		c.Abort()
	}
	if !isOpen && ud.GetID() != 1 && (dao.OneCol(ud, "privilege").ToUint64()&ScanPrivateProblem == 0) {
		setError(c, 403, "没有权限")
		c.Abort()
	}
}

func __CreateProblem(c *gin.Context) {
	ud := &dao.UserDao{ID: getUserID(c)}
	if ud.GetID() != 1 && (dao.OneCol(ud, "privilege").ToUint64()&CreateProblem == 0) {
		setError(c, 403, "没有权限")
		c.Abort()
	}
}

func __UpdateProblem(c *gin.Context) {
	ud := &dao.UserDao{ID: getUserID(c)}
	if ud.GetID() != 1 && (dao.OneCol(ud, "privilege").ToUint64()&UpdateProblem == 0) {
		setError(c, 403, "没有权限")
		c.Abort()
	}
}

func __CopyProblem(c *gin.Context) {
	ud := &dao.UserDao{ID: getUserID(c)}
	if ud.GetID() != 1 && (dao.OneCol(ud, "privilege").ToUint64()&CopyProblem == 0) {
		setError(c, 403, "没有权限")
		c.Abort()
	}
}

func __DeleteProblem(c *gin.Context) {
	ud := &dao.UserDao{ID: getUserID(c)}
	if ud.GetID() != 1 && (dao.OneCol(ud, "privilege").ToUint64()&DeleteProblem == 0) {
		setError(c, 403, "没有权限")
		c.Abort()
	}
}

func __ScanTestdata(c *gin.Context) {
	ud := &dao.UserDao{ID: getUserID(c)}
	if ud.GetID() != 1 && (dao.OneCol(ud, "privilege").ToUint64()&ScanTestdata == 0) {
		setError(c, 403, "没有权限")
		c.Abort()
	}
}

func __UpdateTestdata(c *gin.Context) {
	ud := &dao.UserDao{ID: getUserID(c)}
	if ud.GetID() != 1 && (dao.OneCol(ud, "privilege").ToUint64()&UpdateTestdata == 0) {
		setError(c, 403, "没有权限")
		c.Abort()
	}
}

func __CreateContest(c *gin.Context) {
	ud := &dao.UserDao{ID: getUserID(c)}
	if ud.GetID() != 1 && (dao.OneCol(ud, "privilege").ToUint64()&CreateContest == 0) {
		setError(c, 403, "没有权限")
		c.Abort()
	}
}

func __UpdateContest(c *gin.Context) {
	ud := &dao.UserDao{ID: getUserID(c)}
	if ud.GetID() != 1 && (dao.OneCol(ud, "privilege").ToUint64()&UpdateContest == 0) {
		setError(c, 403, "没有权限")
		c.Abort()
	}
}

func __DeleteContest(c *gin.Context) {
	ud := &dao.UserDao{ID: getUserID(c)}
	if ud.GetID() != 1 && (dao.OneCol(ud, "privilege").ToUint64()&DeleteContest == 0) {
		setError(c, 403, "没有权限")
		c.Abort()
	}
}

func __ScanContestSubmission(c *gin.Context) {
	ud := &dao.UserDao{ID: getUserID(c)}
	if ud.GetID() != 1 && (dao.OneCol(ud, "privilege").ToUint64()&ScanContestSubmission == 0) {
		setError(c, 403, "没有权限")
		c.Abort()
	}
}

func __RejudgeContestProblem(c *gin.Context) {
	ud := &dao.UserDao{ID: getUserID(c)}
	if ud.GetID() != 1 && (dao.OneCol(ud, "privilege").ToUint64()&RejudgeContestProblem == 0) {
		setError(c, 403, "没有权限")
		c.Abort()
	}
}

func __ScanUserInfo(c *gin.Context) {
	ud := &dao.UserDao{ID: getUserID(c)}
	if ud.GetID() != 1 && (dao.OneCol(ud, "privilege").ToUint64()&ScanUserInfo == 0) {
		setError(c, 403, "没有权限")
		c.Abort()
	}
}

func __SetUserPrivilege(c *gin.Context) {
	ud := &dao.UserDao{ID: getUserID(c)}
	if ud.GetID() != 1 && (dao.OneCol(ud, "privilege").ToUint64()&SetUserPrivilege == 0) {
		setError(c, 403, "没有权限")
		c.Abort()
	}
}
