package model

import "time"

const (
	Blog   = iota //博客
	Puzzle        //提问(讨论)
	Solution
)

// 帖子
type Post struct {
	//基础信息
	ID        int64     `json:"id" xorm:"pk autoincr"`
	CreatedAt time.Time `json:"created_at" xorm:"created"`
	UpdatedAt time.Time `json:"updated_at"`

	Head        string `json:"head"`                     //题解没有head                //标题
	Content     string `json:"content" xorm:"text"`      //原生markdown内容
	HtmlContent string `json:"html_content" xorm:"text"` //转换为html的内容
	Kind        uint   `json:"kind" xorm:"index"`        //帖子类型(blog,puzzle,announce)

	AuthorID  int64 `json:"author_id" xorm:"index"`  //作者ID
	ProblemID int64 `json:"problem_id" xorm:"index"` //题解相关的某道题, 求助时可能相关的某道题, 该id也可以为0, 与题目无关的帖子

	GoodCount  uint `json:"good_count"`  //点赞的数量
	BadCount   uint `json:"bad_count"`   //反对
	ReplyCount uint `json:"reply_count"` //回复的数量

	//
	Tags string `json:"tags"` //帖子的标签

	//权限
	CanReply bool `json:"can_reply"` //允许回复
	IsOpen   bool `json:"is_open"`   //
}

type Attitude struct { //个人的态度
	ID       int64           `json:"id" xorm:"pk autoincr"`
	Index    int64           `json:"index" xorm:"index"` //联合索引,(uid<<32)|pid
	ForPost  int             `json:"for_post"`           // 0 无态度,1赞成,2反对
	ForReply map[int64]int64 `json:"for_reply"`          //id
}

type Reply struct { //回复post
	ID           int64     `json:"id" xorm:"pk autoincr"`
	PostID       int64     `json:"post_id" xorm:"index"`
	CreatedAt    time.Time `json:"created_at" xorm:"created"`
	AuthorID     int64     `json:"author_id" xorm:"index"` //用户ID
	HtmlContent  string    `json:"html_content" xorm:"text"`
	CommentCount uint      `json:"comment_count"`
	GoodCount    uint      `json:"good_count"` //点赞的数量
	BadCount     uint      `json:"bad_count"`  //反对
}

type Comment struct { //对reply进行评论
	//基础信息
	ID        int64     `json:"id" xorm:"pk autoincr"`
	CreatedAt time.Time `json:"created_at" xorm:"created"`
	Content   string    `json:"content" xorm:"text"`
	//索引
	AuthorID int64 `json:"author_id" xorm:"index"` //用户ID
	ReplyID  int64 `json:"reply_id" xorm:"index"`  //帖子ID
	PostID   int64 `json:"post_id" xorm:"index"`   //题解的评论
	To       int64 `json:"to"`                     //对于某条
}



//公告
type Announcement struct {
	ID        int64     `json:"id" xorm:"pk autoincr"`
	CreatedAt time.Time `json:"created_at" xorm:"created"`
	Head      string    `json:"head"`
	Content   string    `json:"content" xorm:"text"`
	AuthorID  int64     `json:"author_id" xorm:"index"`
	Grade     int       `json:"grade"`
}


