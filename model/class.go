package model

import "time"

type Class struct {
	ID       int64  `json:"id" xorm:"pk autoincr"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

//家庭作业
type Homework struct {
	ID      int64     `json:"id" xorm:"pk autoincr"`
	Begin   time.Time `json:"begin"`
	End     time.Time `json:"end"`
	Content string    `json:"content" xorm:"text"`
	ClassID int64     `json:"class_id" xorm:"index"`
}

type Student struct {
	ID      int64 `json:"id" xorm:"pk autoincr"`
	UserID  int64 `json:"user_id" xorm:"index"`
	ClassID int64 `json:"class_id" xorm:"index"`
}
