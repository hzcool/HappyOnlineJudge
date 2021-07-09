package model

import "time"

//私信
type Message struct {
	ID             int64     `json:"-" xorm:"pk autoincr"`
	CreatedAt      time.Time `json:"created_at" xorm:"created"`
	Content        string    `json:"content" xorm:"text"`
	From           int64     `json:"from"`
	To             int64     `json:"to"`
	ConversationID int64     `json:"-"  xorm:"index"` //两个人会话的唯一标识物
}

