package app

import (
	"HappyOnlineJudge/common"
	"HappyOnlineJudge/dao"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

func newPost(c *gin.Context) {
	kind := dao.Kind(c.PostForm("kind"))
	if kind == dao.WRONG_KIND {
		setError(c, 403, "文章类型未设置或错误")
		return
	}
	problemDao := &dao.ProblemDao{Index: c.PostForm("problem")}
	if problemDao.Index != "" && problemDao.GetID() == 0 {
		setError(c, 403, "找不到题目")
		return
	}
	userID := getUserID(c)
	pd := &dao.PostDao{
		Post: &dao.Post{
			Head:        c.PostForm("head"),
			Content:     c.PostForm("content"),
			HtmlContent: c.PostForm("html_content"),
			Kind:        kind,
			AuthorID:    userID,
			ProblemID:   problemDao.GetID(),
			Tags:        c.DefaultPostForm("tags", ""),
			CanReply:    common.StrToBool(c.DefaultPostForm("can_reply", "1")),
			IsOpen:      common.StrToBool(c.DefaultPostForm("is_open", "1")),
		},
	}
	if err := dao.Create(pd); err != nil {
		setError(c, 500, err.Error())
		return
	}
	c.Set("post_id", pd.GetID())
	c.Set("created_at", pd.Post.CreatedAt.Format(common.TIME_FORMAT))
}

func deletePost(c *gin.Context) {
	pd := &dao.PostDao{ID: common.StrToInt64(c.DefaultQuery("post_id", "0"))}
	if !dao.Exists(pd) {
		setError(c, 403, "找不到该post")
		return
	}
	if dao.OneCol(pd, "author_id").ToInt64() != getUserID(c) {
		setError(c, 403, "没有权限")
		return
	}
	dao.Delete(pd)
	c.Set("result", "ok")
}

func onePost(c *gin.Context) {
	id := getUserID(c)
	pd := &dao.PostDao{ID: common.StrToInt64(c.DefaultQuery("post_id", "0"))}
	if !dao.Exists(pd) {
		setError(c, 403, "找不到该post")
		return
	}
	wants := []string{"author_id", "problem_id", "head", "content", "tags", "can_reply", "is_open", "kind"}
	cols := dao.Cols(pd, wants...)
	if cols[0].ToInt64() != id {
		setError(c, 403, "没有权限")
		return
	}
	index := ""
	problemDao := &dao.ProblemDao{ID: cols[1].ToInt64()}
	if problemDao.ID != 0 {
		index = problemDao.GetIndex()
	}
	post := common.H{
		"head":      cols[2].ToString(),
		"content":   cols[3].ToString(),
		"tags":      cols[4].ToString(),
		"can_reply": cols[5].ToBool(),
		"is_open":   cols[6].ToBool(),
		"kind":      dao.KindName(cols[7].ToUint()),
		"index":     index,
	}
	if index != "" {
		post["problem"] = index
	}
	c.Set("post", post)
}

func updatePost(c *gin.Context) {
	id := getUserID(c)
	pd := &dao.PostDao{ID: common.StrToInt64(c.DefaultPostForm("post_id", "0"))}
	if !dao.Exists(pd) {
		setError(c, 403, "找不到该post")
		return
	}
	if dao.OneCol(pd, "author_id").ToInt64() != id {
		setError(c, 403, "权限不够")
		return
	}
	problemDao := &dao.ProblemDao{Index: c.PostForm("problem")}
	if problemDao.Index != "" && problemDao.GetID() == 0 {
		setError(c, 403, "找不到题目")
		return
	}
	mp := common.H{
		"head":         c.PostForm("head"),
		"content":      c.PostForm("content"),
		"updated_at":   time.Now(),
		"html_content": c.PostForm("html_content"),
		"tags":         c.PostForm("tags"),
		"can_reply":    common.StrToBool(c.DefaultPostForm("can_reply", "1")),
		"is_open":      common.StrToBool(c.DefaultPostForm("is_open", "1")),
	}
	if problemDao.ID != 0 {
		mp["problem_id"] = problemDao.ID
	}
	if err := dao.UpdateCols(pd, mp); err != nil {
		setError(c, 500, err.Error())
		return
	}
}

func countMyPost(c *gin.Context) {
	c.Set("post_count", dao.CountMyPosts(getUserID(c)))
}

func getPostList(c *gin.Context) {
	kind := c.Query("kind")
	l, err1 := strconv.ParseInt(c.DefaultQuery("l", "1"), 10, 64)
	r, err2 := strconv.ParseInt(c.DefaultQuery("r", "10"), 10, 64)
	pd := &dao.ProblemDao{Index: c.DefaultQuery("index", "")}
	problemID := int64(0)
	if pd.Index != "" {
		problemID = pd.GetID()
		if problemID == 0 {
			setError(c, 403, "参数错误")
			return
		}
	}
	if err1 != nil || err2 != nil || kind == "" || (kind == "solution" && problemID == 0) {
		setError(c, 403, "参数错误")
		return
	}
	if problemID != 0 {
		total, data := dao.GetProblemPostList(problemID, l, r, getUserID(c), dao.Kind(kind))
		c.Set("total", total)
		c.Set("data", data)
		return
	}
	posts := dao.GetPostList(dao.Kind(kind), problemID, l, r)
	c.Set("data", posts)
}

func getUserPostList(c *gin.Context) {
	ud := &dao.UserDao{ID: getUserID(c)}
	kind := c.DefaultQuery("kind", "blog")
	posts := dao.GetUserPostList(ud.GetID(), dao.Kind(kind))
	c.Set("data", posts)
}

func getPost(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Query("post_id"), 10, 64)
	if err != nil {
		setError(c, 403, "参数错误")
		return
	}
	pd := &dao.PostDao{ID: postID}
	if !dao.Exists(pd) {
		setError(c, 403, "Not Found")
		return
	}
	uid := getUserID(c)
	dao.GetSelfAll(pd)
	if !pd.Post.IsOpen {
		if uid != pd.Post.AuthorID {
			setError(c, 403, "没有权限查看")
			return
		}
	}
	forPost := 0
	if uid != 0 {
		atti := dao.GetAttitude(uid, postID)
		if atti != nil {
			forPost = atti.ForPost
		}
	}
	ud := &dao.UserDao{ID: pd.Post.AuthorID}
	data := common.H{
		"created_at":   pd.Post.CreatedAt.Format(common.TIME_FORMAT),
		"head":         pd.Post.Head,
		"html_content": pd.Post.HtmlContent,
		"author":       ud.GetName(),
		"avatar":       dao.OneCol(ud, "avatar").ToString(),
		"good_count":   pd.Post.GoodCount,
		"bad_count":    pd.Post.BadCount,
		"reply_count":  pd.Post.ReplyCount,
		"tags":         pd.Post.Tags,
		"can_reply":    pd.Post.CanReply,
		"is_open":      pd.Post.IsOpen,
		"content":      pd.Post.Content,
		"for_post":     forPost,
		"kind":         dao.KindName(pd.Post.Kind),
	}
	if pd.Post.ProblemID != 0 {
		problemDao := &dao.ProblemDao{ID: pd.Post.ProblemID}
		cols := dao.Cols(problemDao, "index", "title")
		data["index"] = cols[0].ToString()
		data["title"] = cols[1].ToString()
	}
	if pd.Post.UpdatedAt.Unix() > pd.Post.CreatedAt.Unix() {
		data["updated_at"] = pd.Post.UpdatedAt.Format(common.TIME_FORMAT)
	}
	c.Set("post", data)
}

func getReplies(c *gin.Context) {
	pid, err := strconv.ParseInt(c.DefaultQuery("post_id", "0"), 10, 64)
	l, err2 := strconv.ParseInt(c.DefaultQuery("l", "0"), 10, 64)
	r, err3 := strconv.ParseInt(c.DefaultQuery("r", "0"), 10, 64)
	if l == 0 || r == 0 || pid == 0 || err != nil || err2 != nil || err3 != nil {
		setError(c, 403, "参数错误")
		return
	}
	total, replies := dao.GetReplies(pid, l, r)
	c.Set("replies", replies)
	c.Set("total", total)
}
func sendOneReply(c *gin.Context) {
	pid, err := strconv.ParseInt(c.DefaultPostForm("post_id", "0"), 10, 64)
	if pid == 0 || err != nil {
		setError(c, 403, "参数错误")
		return
	}
	pd := &dao.PostDao{ID: pid}
	if !dao.Exists(pd) {
		setError(c, 403, "Not Find")
		return
	}
	if !dao.OneCol(pd, "can_reply").ToBool() {
		setError(c, 403, "作者设置禁止评论")
		return
	}
	reply := dao.NewReply(pid, getUserID(c), c.PostForm("editor"))
	if reply == nil {
		setError(c, 500, "操作失败")
		return
	}
	c.Set("result", "ok")
}

func addPostAttitude(c *gin.Context) {
	pid, err := strconv.ParseInt(c.DefaultQuery("post_id", "0"), 10, 64)
	if err != nil || pid == 0 {
		setError(c, 403, "参数错误")
		return
	}
	att, err2 := strconv.ParseInt(c.DefaultQuery("attitude", "-1"), 10, 64)
	if err2 != nil || att == -1 {
		setError(c, 403, "参数错误")
		return
	}
	uid := getUserID(c)
	goodCount, badCount := dao.UpdatePostAttitude(uid, pid, int(att))
	c.Set("good_count", goodCount)
	c.Set("bad_count", badCount)
}

func getComments(c *gin.Context) {
	pid := common.StrToInt64(c.Query("post_id"))
	rid := common.StrToInt64(c.DefaultQuery("reply_id", "0"))
	l := common.StrToInt64(c.DefaultQuery("l", "1"))
	r := common.StrToInt64(c.DefaultQuery("r", "1000"))
	total, comments := dao.GetComments(pid, rid, l, r)
	c.Set("comments", comments)
	c.Set("total_comments", total)
}

func sendOneComment(c *gin.Context) {
	pid, err := strconv.ParseInt(c.DefaultPostForm("post_id", "0"), 10, 64)
	if pid == 0 || err != nil {
		setError(c, 403, "参数错误")
		return
	}
	pd := &dao.PostDao{ID: pid}
	if !dao.Exists(pd) {
		setError(c, 403, "Not Find")
		return
	}
	if !dao.OneCol(pd, "can_reply").ToBool() {
		setError(c, 403, "作者设置禁止评论")
		return
	}

	uid := getUserID(c)
	rid := common.StrToInt64(c.DefaultPostForm("reply_id", "0"))
	toDao := &dao.UserDao{Username: c.DefaultPostForm("to", "")}
	toID := int64(0)
	if toDao.Username != "" {
		toID = toDao.GetID()
	}
	comment := dao.NewComment(pid, uid, rid, toID, c.PostForm("content"))
	if comment == nil {
		setError(c, 403, "插入失败")
		return
	}
	ud := &dao.UserDao{ID: uid}
	c.Set("comment", common.H{
		"created_at": comment.CreatedAt.Format(common.TIME_FORMAT),
		"content":    comment.Content,
		"author":     getUserName(c),
		"to":         toDao.Username,
		"avatar":     dao.OneCol(ud, "avatar").ToString(),
	})
}

func getContestDiscussList(c *gin.Context) {
	cid, err := strconv.ParseInt(c.Query("id"), 10, 64)
	if err != nil {
		setError(c, 403, "参数错误")
		return
	}
	uid := getUserID(c)
	if uid != 1 && !dao.ExistTeam(cid, uid) {
		setError(c, 403, "Not Found")
		return
	}
	getPostList(c)
}
