package model

import "time"

type Submission struct {
	ID          int64     `json:"id" xorm:"pk autoincr"`
	CreatedAt   time.Time `json:"created_at" xorm:"created"`
	Code        string    `json:"code" xorm:"text notnull"`
	Time        uint      `json:"time"`
	Memory      uint      `json:"memory"`
	Length      uint      `json:"length"`
	Status      string    `json:"status" xorm:"varchar(20) default 'Queueing'"`
	CompileInfo string    `json:"compile_info" xorm:"text"`
	Lang        string    `json:"lang" xorm:"varchar(10)"`
	Score       uint      `json:"score"`
	ProblemID   int64     `json:"problem_id" xorm:"index"`
	AuthorID    int64     `json:"author_id" xorm:"index"`
}
