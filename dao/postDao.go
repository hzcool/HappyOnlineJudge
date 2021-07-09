package dao

import (
	"HappyOnlineJudge/common"
	"HappyOnlineJudge/model"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"strconv"
	"strings"
	"time"
)

type (
	Post         = model.Post
	Comment      = model.Comment
	Attitude     = model.Attitude
	Reply        = model.Reply
	Announcement = model.Announcement
)

/*
redis 结构
个人
blog_list_of_user:id => postid
puzzle_list_of_user:id => postid

solution_list_of_user:id => postid
solution_zset:problemid => postid:postid

puzzle_zset:problremid => postid:postid

attitude(uid and postid) => Hash
reply_zset_of_post:id =>  reply_id:reply_id
comment_list_of_(reply|postkind:id) = listHash
*/

const (
	POST_REDIS_EXPIRE     = time.Hour * 24
	USER_POST_EXPIRE      = time.Hour * 3
	ATTITUDE_REDIS_EXPIRE = time.Hour * 3
	REPLY_REDIS_EXPIRE    = time.Hour * 3
	COMMENT_REDIS_EXPIRE  = time.Minute * 30
	WRONG_KIND            = 10
)

func Kind(kd string) uint {
	kd = strings.ToLower(kd)
	if kd == "blog" || kd == strconv.Itoa(model.Blog) {
		return model.Blog
	} else if kd == "puzzle" || kd == strconv.Itoa(model.Puzzle) {
		return model.Puzzle
	} else if kd == "solution" || kd == strconv.Itoa(model.Solution) {
		return model.Solution
	}
	return WRONG_KIND
}
func KindName(kd uint) string {
	if kd == model.Blog {
		return "blog"
	} else if kd == model.Puzzle {
		return "puzzle"
	} else if kd == model.Solution {
		return "solution"
	}
	return "solution"
}

type PostDao struct {
	ID   int64
	Post *Post
}

func (pd *PostDao) GetTableName() string {
	return "post"
}
func (pd *PostDao) GetRedisExpire() time.Duration {
	return POST_REDIS_EXPIRE
}
func (pd *PostDao) GetSelf() interface{} {
	if pd.Post == nil {
		pd.Post = &Post{}
	}
	return pd.Post
}
func (pd *PostDao) GetID() int64 {
	if pd.ID == 0 {
		if pd.Post != nil {
			pd.ID = pd.Post.ID
		}
	}
	return pd.ID
}
func (pd *PostDao) GetRedisKey() string {
	return pd.GetTableName() + "_" + strconv.FormatInt(pd.GetID(), 10)
}
func (pd *PostDao) BeforePutToRedis() error {
	lkey := getPrivatePostKey(pd.Post.Kind, pd.Post.AuthorID)
	if rdb.Exists(ctx, lkey).Val() <= 0 {
		rdb.RPush(ctx, lkey, pd.GetID())
	}
	if pd.Post.ProblemID != 0 {
		zkey := getProblemPostZsetKey(KindName(pd.Post.Kind), pd.Post.ProblemID)
		if rdb.Exists(ctx, zkey).Val() > 0 {
			rdb.ZAdd(ctx, zkey, &redis.Z{Score: float64(pd.Post.ID), Member: pd.Post.ID})
			rdb.Expire(ctx, zkey, POST_REDIS_EXPIRE)
		}
	}
	return nil
}
func (pd *PostDao) BeforeDelete() error {
	rdb.Del(ctx, getPrivatePostKey(model.Blog, pd.GetID()))
	rdb.Del(ctx, getPrivatePostKey(model.Puzzle, pd.GetID()))
	rdb.Del(ctx, getPrivatePostKey(model.Solution, pd.GetID()))
	return nil
}

func getPrivatePostKey(kind uint, uid int64) string {
	return KindName(kind) + "_list_of_user:" + strconv.FormatInt(uid, 10)
}

func cacheUserPostList(kind uint, uid int64) string {
	lkey := getPrivatePostKey(kind, uid)
	if rdb.Exists(ctx, lkey).Val() <= 0 {
		ids := make([]int64, 0)
		engine.Table("post").Where("`author_id`=? and `kind`=?", uid, kind).Cols("id").Find(&ids)
		data := make([]interface{}, len(ids))
		for i, item := range ids {
			data[i] = item
		}
		rdb.RPush(ctx, lkey, data...)
		rdb.Expire(ctx, lkey, USER_POST_EXPIRE)
	}
	return lkey
}

func CountMyPosts(uid int64) common.H {
	blogKey := cacheUserPostList(model.Blog, uid)
	puzzleKey := cacheUserPostList(model.Puzzle, uid)
	solutionKey := cacheUserPostList(model.Solution, uid)
	return common.H{
		"blog_count":     rdb.LLen(ctx, blogKey).Val(),
		"puzzle_count":   rdb.LLen(ctx, puzzleKey).Val(),
		"solution_count": rdb.LLen(ctx, solutionKey).Val(),
	}
}

func getProblemPostZsetKey(kind string, problemID int64) string {
	return "problem_" + kind + "_zset" + strconv.FormatInt(problemID, 10)
}
func cahceProblemPostList(kind uint, problemID int64) string {
	zkey := getProblemPostZsetKey(KindName(kind), problemID)
	if rdb.Exists(ctx, zkey).Val() <= 0 {
		ids := make([]int64, 0)
		engine.Table("post").Where("`kind` = ? and `problem_id` = ? and `is_open` = ? ", kind, problemID, 1).Cols("id").Find(&ids)
		if len(ids) > 0 {
			data := make([]*redis.Z, len(ids))
			for i, item := range ids {
				data[i] = &redis.Z{
					Score:  float64(item),
					Member: item,
				}
			}
			rdb.ZAdd(ctx, zkey, data...)
			rdb.Expire(ctx, zkey, POST_REDIS_EXPIRE)
		}
	}
	return zkey
}

func GetProblemPostList(problemID, l, r, uid int64, kind uint) (int64, []common.H) {
	zkey := cahceProblemPostList(kind, problemID)
	ids := rdb.ZRange(ctx, zkey, -r, -l).Val()
	d := len(ids)
	data := make([]common.H, d)
	for i, pidStr := range ids {
		pid := common.StrToInt64(pidStr)
		postDao := &PostDao{ID: pid}
		GetSelfAll(postDao)
		item := postDao.Post
		ud := &UserDao{ID: item.AuthorID}
		data[i] = common.H{
			"post_id":      item.ID,
			"created_at":   item.CreatedAt.Format(common.TIME_FORMAT),
			"author":       ud.GetName(),
			"content":      item.Content,
			"content_html": item.HtmlContent,
			"good_count":   item.GoodCount,
			"bad_count":    item.BadCount,
			"reply_count":  item.ReplyCount,
			"tags":         item.Tags,
			"avatar":       OneCol(ud, "avatar").ToString(),
			"head":         item.Head,
		}
		if item.UpdatedAt.After(item.CreatedAt) {
			data[i]["updated_at"] = item.UpdatedAt.Format(common.TIME_FORMAT)
		}
		pd := &ProblemDao{ID: item.ProblemID}
		data[i]["index"] = pd.GetIndex()
		data[i]["title"] = OneCol(pd, "title")
		if uid != 0 {
			if att := GetAttitude(uid, pid); att != nil {
				data[i]["for_post"] = att.ForPost
			} else {
				data[i]["for_post"] = 0
			}
		}
	}
	return rdb.ZCard(ctx, zkey).Val(), data
}

func GetPostList(kind uint, problemID, l, r int64) []common.H { //blog, puzzle
	posts := make([]Post, 0)
	engine.Desc("id").Where("`kind` = ? and `problem_id` = ? and `is_open` = ? ", kind, problemID, 1).Limit(int(r-l+1), int(l-1)).Omit("content", "html_content").Find(&posts)
	d := len(posts)
	data := make([]common.H, d)
	for i, item := range posts {
		ud := &UserDao{ID: item.AuthorID}
		data[i] = common.H{
			"post_id":     item.ID,
			"created_at":  item.CreatedAt.Format(common.TIME_FORMAT),
			"head":        item.Head,
			"author":      ud.GetName(),
			"good_count":  item.GoodCount,
			"bad_count":   item.BadCount,
			"reply_count": item.ReplyCount,
			"tags":        item.Tags,
			"avatar":      OneCol(ud, "avatar").ToString(),
		}
		if item.ProblemID != 0 {
			pd := &ProblemDao{ID: item.ProblemID}
			data[i]["index"] = pd.GetIndex()
		}
	}
	return data
}

func GetUserPostList(uid int64, kind uint) []common.H {
	key := cacheUserPostList(kind, uid)
	ids := rdb.LRange(ctx, key, 0, -1).Val()
	d := len(ids)
	data := make([]common.H, d)
	wants := []string{"id", "created_at", "head", "author_id", "problem_id", "good_count", "bad_count", "reply_count", "tags"}
	for i, item := range ids {
		pd := &PostDao{ID: common.StrToInt64(item)}
		cols := Cols(pd, wants...)
		ud := &UserDao{ID: cols[3].ToInt64()}
		data[d-i-1] = common.H{
			"kind":        KindName(kind),
			"post_id":     cols[0].ToInt64(),
			"created_at":  cols[1].ToString(),
			"head":        cols[2].ToString(),
			"author":      ud.GetName(),
			"good_count":  cols[5].ToUint(),
			"bad_count":   cols[6].ToUint(),
			"reply_count": cols[7].ToUint(),
			"tags":        cols[8].ToString(),
			"avatar":      OneCol(ud, "avatar").ToString(),
		}
		pid := cols[4].ToUint64()
		if pid != 0 {
			problemDao := &ProblemDao{ID: cols[4].ToInt64()}
			data[d-i-1]["index"] = problemDao.GetIndex()
			data[d-i-1]["title"] = OneCol(problemDao, "title").ToString()
		}
	}
	return data
}

func getAttitudekey(uid, pid int64) string {
	return "attitude:" + strconv.FormatInt((uid<<32)|pid, 10)
}
func GetAttitude(uid, pid int64) *Attitude {
	key := getAttitudekey(uid, pid)
	attitude := &Attitude{}
	if rdb.Exists(ctx, key).Val() > 0 {
		json.Unmarshal([]byte(rdb.Get(ctx, key).Val()), attitude)
	} else {
		if exist, err := engine.Where("`index`=?", (uid<<32)|pid).Get(attitude); !exist || err != nil {
			return nil
		}
		js, _ := json.Marshal(attitude)
		rdb.Set(ctx, key, js, ATTITUDE_REDIS_EXPIRE)
	}
	return attitude
}
func UpdatePostAttitude(uid, pid int64, att int) (uint, uint) {
	attitude := GetAttitude(uid, pid)
	pd := &PostDao{ID: pid}
	cols := Cols(pd, "good_count", "bad_count")
	good := cols[0].ToUint()
	bad := cols[1].ToUint()
	if attitude == nil {

		attitude = &Attitude{
			Index:   (uid << 32) | pid,
			ForPost: att,
		}
		if num, err := engine.InsertOne(attitude); num != 1 || err != nil {
			return good, bad
		}
		if att == 1 {
			good++
		} else if att == 2 {
			bad++
		}
	} else if att != attitude.ForPost {
		switch (att << 2) + attitude.ForPost {
		case 1:
			good--
		case 2:
			bad--
		case 4:
			good++
		case 6:
			bad--
			good++
		case 8:
			bad++
		case 9:
			good--
			bad++
		}
		attitude.ForPost = att
		UpdateColsBySql("attitude", attitude.ID, common.H{"for_post": att})
	}
	js, _ := json.Marshal(attitude)
	rdb.Set(ctx, getAttitudekey(uid, pid), js, ATTITUDE_REDIS_EXPIRE)
	UpdateCols(pd, common.H{"good_count": good, "bad_count": bad})
	return good, bad
}

type ReplyDao struct {
	ID    int64
	reply *Reply
}

func (rd *ReplyDao) GetTableName() string {
	return "reply"
}
func (rd *ReplyDao) GetRedisExpire() time.Duration {
	return REPLY_REDIS_EXPIRE
}
func (rd *ReplyDao) GetSelf() interface{} {
	if rd.reply == nil {
		rd.reply = &Reply{}
	}
	return rd.reply
}
func (rd *ReplyDao) GetID() int64 {
	if rd.ID == 0 {
		if rd.reply != nil {
			rd.ID = rd.reply.ID
		}
	}
	return rd.ID
}
func (rd *ReplyDao) GetRedisKey() string {
	return rd.GetTableName() + "_" + strconv.FormatInt(rd.GetID(), 10)
}
func (rd *ReplyDao) BeforePutToRedis() error {
	zkey := getReplyZsetKey(rd.reply.PostID)
	if rdb.Exists(ctx, zkey).Val() >= 1 {
		rdb.ZAdd(ctx, zkey, &redis.Z{
			Score:  float64(rd.GetID()),
			Member: rd.GetID(),
		})
		rdb.Expire(ctx, zkey, REPLY_REDIS_EXPIRE)
	}
	return nil
}
func (rd *ReplyDao) BeforeDelete() error {
	rdb.ZRem(ctx, getReplyZsetKey(rd.GetID()), strconv.FormatInt(rd.ID, 10))
	return nil
}

func getReplyZsetKey(pid int64) string {
	return "Reply_zset_of_post:" + strconv.FormatInt(pid, 10)
}
func cacheReplyZset(pid int64) string {
	zkey := getReplyZsetKey(pid)
	if rdb.Exists(ctx, zkey).Val() <= 0 {
		ids := make([]int64, 0)
		engine.Table("reply").Where("`post_id` = ?", pid).Cols("id").Find(&ids)
		data := make([]*redis.Z, len(ids))
		for i, item := range ids {
			data[i] = &redis.Z{
				Score:  float64(item),
				Member: item,
			}
		}
		rdb.ZAdd(ctx, zkey, data...)
		rdb.Expire(ctx, zkey, REPLY_REDIS_EXPIRE)
	}
	return zkey
}
func GetReplies(pid, l, r int64) (int64, []common.H) {
	zkey := cacheReplyZset(pid)
	x := rdb.ZRange(ctx, zkey, l-1, r-1).Val()
	data := make([]common.H, len(x))
	for i, item := range x {
		rd := &ReplyDao{ID: common.StrToInt64(item)}
		GetSelfAll(rd)
		re := rd.reply
		ud := &UserDao{ID: re.AuthorID}
		data[i] = common.H{
			"reply_id":      re.ID,
			"created_at":    re.CreatedAt.Format(common.TIME_FORMAT),
			"html_content":  re.HtmlContent,
			"author":        ud.GetName(),
			"good_count":    re.GoodCount,
			"bad_count":     re.BadCount,
			"comment_count": re.CommentCount,
			"avatar":        OneCol(ud, "avatar").ToString(),
		}
	}
	return rdb.ZCard(ctx, zkey).Val(), data
}
func NewReply(pid, uid int64, htmlContent string) *Reply {
	re := &Reply{
		PostID:      pid,
		AuthorID:    uid,
		HtmlContent: htmlContent,
	}
	rd := &ReplyDao{reply: re}
	if err := Create(rd); err != nil {
		return nil
	}
	pd := &PostDao{ID: pid}
	num := OneCol(pd, "reply_count").ToUint()
	UpdateCols(pd, common.H{"reply_count": num + 1})
	return re
}

func getCommentListKey(pid, rid int64) string {
	return "comment_list:" + strconv.FormatInt((pid<<32)|rid, 10)
}

func cacheCommentList(pid, rid int64) string {
	lkey := getCommentListKey(pid, rid)
	if rdb.Exists(ctx, lkey).Val() <= 0 {
		comments := make([]Comment, 0)
		if rid != 0 {
			engine.Where("`reply_id` = ?", rid).Find(&comments)
		} else {
			engine.Where("`post_id` = ?", pid).Find(&comments)
		}
		data := make([]interface{}, len(comments))
		for i, item := range comments {
			data[i], _ = json.Marshal(item)
		}
		rdb.RPush(ctx, lkey, data...)
		rdb.Expire(ctx, lkey, COMMENT_REDIS_EXPIRE)
	}
	return lkey
}

func GetComments(pid, rid, l, r int64) (int64, []common.H) {
	lkey := cacheCommentList(pid, rid)
	x := rdb.LRange(ctx, lkey, l-1, r-1).Val()
	data := make([]common.H, len(x))
	for i, item := range x {
		c := &Comment{}
		json.Unmarshal([]byte(item), c)
		ud := &UserDao{ID: c.AuthorID}
		to := ""
		if c.To != 0 {
			ud2 := &UserDao{ID: c.To}
			to = ud2.GetName()
		}
		data[i] = common.H{
			"created_at": c.CreatedAt.Format(common.TIME_FORMAT),
			"content":    c.Content,
			"author":     ud.GetName(),
			"to":         to,
			"avatar":     OneCol(ud, "avatar").ToString(),
		}
	}
	return rdb.LLen(ctx, lkey).Val(), data
}

func NewComment(pid, uid, rid, to int64, content string) *Comment {
	comment := &Comment{
		PostID:   pid,
		AuthorID: uid,
		Content:  content,
		ReplyID:  rid,
		To:       to,
	}
	lkey := cacheCommentList(pid, rid)
	if num, err := engine.InsertOne(comment); num != 1 && err != nil {
		return nil
	}
	js, _ := json.Marshal(comment)
	rdb.RPush(ctx, lkey, js)
	rdb.Expire(ctx, lkey, COMMENT_REDIS_EXPIRE)
	if rid != 0 {
		rd := &ReplyDao{ID: rid}
		num := OneCol(rd, "comment_count").ToUint()
		UpdateCols(rd, common.H{"comment_count": num + 1})
	} else {
		pd := &PostDao{ID: pid}
		num := OneCol(pd, "reply_count").ToUint()
		UpdateCols(pd, common.H{"reply_count": num + 1})
	}
	return comment
}

//func getHomeworkKey() string {
//	return "homework_zset"
//}
//func cacheHomework() string {
//	zkey := getHomeworkKey()
//	if rdb.Exists(ctx, zkey).Val() <= 0 {
//		hs := make([]Homework, 0)
//		engine.Find(&hs)
//		data := make([]*redis.Z, len(hs))
//		for i, item := range hs {
//			js, _ := json.Marshal(item)
//			data[i] = &redis.Z{
//				Score:  -float64(item.Grade),
//				Member: js,
//			}
//		}
//		rdb.ZAdd(ctx, zkey, data...)
//	}
//	return zkey
//}
//func NewHomeWork(uid int64, begin, end time.Time, head, content string, grade int) *Homework {
//	h := &Homework{
//		Begin:    begin,
//		End:      end,
//		Head:     head,
//		Content:  content,
//		AuthorID: uid,
//		Grade:    grade,
//	}
//	if num, err := engine.InsertOne(h); num != 1 && err != nil {
//		return nil
//	}
//	key := cacheHomework()
//	js, _ := json.Marshal(h)
//	rdb.ZAdd(ctx, key, &redis.Z{
//		Score:  -float64(h.Grade),
//		Member: js,
//	})
//	return h
//}
//
//func DeleteHomeWork(hid int64) error {
//	h := &Homework{}
//	engine.Where("id = ?", hid).Get(h)
//	js, _ := json.Marshal(h)
//	rdb.ZRem(ctx, getHomeworkKey(), js)
//	engine.ID(h.ID).Delete(&Homework{})
//	return nil
//}
//
//func getAnnouncementKey() string {
//	return "announcement_zset"
//}
//func cacheAnnouncement() string {
//	zkey := getAnnouncementKey()
//	if rdb.Exists(ctx, zkey).Val() <= 0 {
//		as := make([]Announcement, 0)
//		engine.Find(&as)
//		data := make([]*redis.Z, len(as))
//		for i, item := range as {
//			js, _ := json.Marshal(item)
//			data[i] = &redis.Z{
//				Score:  -float64(item.Grade),
//				Member: js,
//			}
//		}
//		rdb.ZAdd(ctx, zkey, data...)
//	}
//	return zkey
//}
//func NewAnnouncement(uid int64, head, content string, grade int) *Announcement {
//	h := &Announcement{
//		Head:     head,
//		Content:  content,
//		AuthorID: uid,
//		Grade:    grade,
//	}
//	if num, err := engine.InsertOne(h); num != 1 && err != nil {
//		return nil
//	}
//	js, _ := json.Marshal(h)
//	rdb.ZAdd(ctx, getHomeworkKey(), &redis.Z{
//		Score:  -float64(h.Grade),
//		Member: js,
//	})
//	return h
//}
//
//func DeleteAnnouncement(aid int64) error {
//	h := &Announcement{}
//	engine.Where("id = ?", aid).Get(h)
//	js, _ := json.Marshal(h)
//	rdb.ZRem(ctx, getHomeworkKey(), js)
//	engine.ID(h.ID).Delete(&Announcement{})
//	return nil
//}
