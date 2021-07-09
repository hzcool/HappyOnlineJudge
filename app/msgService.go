package app

import (
	"HappyOnlineJudge/common"
	"HappyOnlineJudge/dao"
	"github.com/gin-gonic/gin"
)

func findContactPerson(c *gin.Context) {
	username := getUserName(c)
	person := c.DefaultQuery("person", "")
	if username == "" || person == "" || person == username {
		setError(c, 403, "参数错误")
		return
	}
	ud := &dao.UserDao{Username: person}
	if ud.GetID() == 0 {
		setError(c, 403, "无此用户")
		return
	}
	c.Set("chat", map[string]interface{}{
		"unread_count": 0,
		"person":       person,
		"avatar":       dao.OneCol(ud, "avatar").ToString(),
	})
}

func getMessages(c *gin.Context) {
	username := getUserName(c)
	person := c.DefaultQuery("person", "")
	l := common.StrToInt64(c.DefaultQuery("l", "1"))
	r := common.StrToInt64(c.DefaultQuery("r", "10"))
	if username == "" || person == "" || person == username {
		setError(c, 403, "参数错误")
		return
	}
	ud1 := &dao.UserDao{Username: username}
	ud2 := &dao.UserDao{Username: person}
	err, messages := dao.GetMessages(ud1.GetID(), ud2.GetID(), l, r)
	if err != nil {
		setError(c, 500, "查询错误")
		return
	}
	if c.DefaultQuery("unread", "") != "" {
		dao.ClearOnePersonUnread(ud1.GetID(), ud2.GetID())
	}
	ret := make([]map[string]interface{}, len(messages))
	len := len(messages)
	for idx := 0; idx < len; idx++ {
		item := &messages[idx]
		if item.From == ud1.GetID() {
			ret[len-idx-1] = map[string]interface{}{
				"from":       username,
				"to":         person,
				"created_at": item.CreatedAt.Format(common.TIME_FORMAT),
				"content":    item.Content,
			}
		} else {
			ret[len-idx-1] = map[string]interface{}{
				"from":       person,
				"to":         username,
				"created_at": item.CreatedAt.Format(common.TIME_FORMAT),
				"content":    item.Content,
			}
		}
	}
	c.Set("messages", ret)
}

func sendOneMessage(c *gin.Context) {
	username := getUserName(c)
	person := c.PostForm("person")
	content := c.PostForm("content")
	if username == "" || person == "" || content == "" || person == username {
		setError(c, 403, "参数错误")
		return
	}
	ud1 := &dao.UserDao{Username: username}
	ud2 := &dao.UserDao{Username: person}
	err, t := dao.SendOneMessage(ud1.GetID(), ud2.GetID(), content)
	if err != nil {
		setError(c, 403, err.Error())
		return
	}
	c.Set("message", map[string]interface{}{
		"from":       username,
		"to":         person,
		"content":    content,
		"created_at": t.Format(common.TIME_FORMAT),
	})
}

func getContacts(c *gin.Context) {
	username := getUserName(c)
	ud := &dao.UserDao{Username: username}
	conversations := dao.GetContacts(ud.GetID())
	contacts := make([]map[string]interface{}, len(conversations))
	for idx, z := range conversations {
		ud2 := &dao.UserDao{ID: common.StrToInt64(z.Member.(string))}
		contacts[idx] = map[string]interface{}{
			"person":       dao.OneCol(ud2, "username").ToString(),
			"avatar":       dao.OneCol(ud2, "avatar").ToString(),
			"unread_count": int64(z.Score),
		}
	}
	c.Set("contacts", contacts)
}

func delOneContact(c *gin.Context) {
	username := getUserName(c)
	person := c.DefaultQuery("person", "")
	if username == "" || person == "" || person == username {
		setError(c, 403, "参数错误")
		return
	}
	ud1 := &dao.UserDao{Username: username}
	ud2 := &dao.UserDao{Username: person}
	dao.ChangeToClosed(ud1.GetID(), ud2.GetID())
}
