package model

import "time"

type Contest struct {
	ID            int64     `json:"id" xorm:"pk autoincr"`
	Title         string    `json:"title" xorm:"varchar(32) notnull"`
	Begin         time.Time `json:"begin"`
	End           time.Time `json:"end"`
	Length        uint      `json:"length"` //单位秒
	Desc          string    `json:"desc"`
	Author        string    `json:"author"`
	IsPublic      bool      `json:"is_public"`
	Password      string    `json:"password" xorm:"VARBINARY(16)"`
	Format        string    `json:"format" xorm:"varchar(10) default 'OI'"` //ACM 或 OI
	Status        string    `json:"status" xorm:"default 'Pending'"`        //Pending,Running,Ended
	Clarification []string  `json:"clarification"`
}

func (c *Contest) GetTableName() string {
	return "contest"
}

type Team struct {
	ID            int64                      `json:"id" xorm:"pk autoincr"`
	ContestID     int64                      `json:"contest_id" xorm:"index"`
	UserID        int64                      `json:"user_id" xorm:"index"`
	Solved        uint                       `json:"solved"`
	Penalty       uint                       `json:"penalty"`
	Scores        uint                       `json:"scores"`
	ProblemStatus map[string]map[string]uint `json:"problem_status"`
}

func (t *Team) GetTableName() string {
	return "team"
}

//ProblemStatus  {'A':{'fail_times':uint,'minutes':ac时间, 'score':分数, 'last': 上一次提交时间(s)}}
type Cproblem struct {
	ID             int64  `json:"id" xorm:"pk autoincr"`
	ProblemID      int64  `json:"problem_id" xorm:"index"`
	ContestID      int64  `json:"contest_id" xorm:"index"`
	Label          string `json:"label"`
	FirstSolveTime uint   `json:"first_solve_time"`
	Tags           string `json:"tags"`
	Title          string `json:"title"`
	AC             uint   `json:"ac"`
	All            uint   `json:"all"`
}

func (cp *Cproblem) GetTableName() string {
	return "cproblem"
}

type Csubmission struct {
	ID          int64     `json:"id" xorm:"pk autoincr"`
	RunID       uint      `json:"run_id" xorm:"index"`
	ContestID   int64     `json:"contest_id" xorm:"index"`
	UserID      int64     `json:"user_id" xorm:"index"`
	CreatedAt   time.Time `json:"created_at" xorm:"created"`
	Code        string    `json:"code" xorm:"text"`
	Time        uint      `json:"time"`
	Memory      uint      `json:"memory"`
	Length      uint      `json:"length"`
	Status      string    `json:"status" xorm:"varchar(20) default 'Queueing'"`
	CompileInfo string    `json:"compile_info" xorm:"text"`
	Lang        string    `json:"lang" xorm:"varchar(20)"`
	Score       uint      `json:"score"`
	Label       string    `json:"label"`
}

func (cs *Csubmission) GetTableName() string {
	return "csubmission"
}
func (cs *Csubmission) JudgeID() int64 {
	return (cs.ContestID << 32) | (int64(cs.Label[0]) - int64('A'))
}

//type Cquesttion struct { //提问
//	ID        int64  `json:"id" xorm:"pk autoincr"`
//	Content   string `json:"content" xorm:"text"`
//	AuthorID  int64  `json:"author_id"`
//	ContestID int64  `json:"contest_id" xorm:"index"`
//	Label     string `json:"label"`
//}
//
//type Canswer struct {
//	ID         int64  `json:"id" xorm:"pk autoincr"`
//	Content    string `json:"content" xorm:"text"`
//	AuthorID   int64  `json:"author_id"`
//	Cquesttion int64  `json:"cquesttion" xorm:"index"`
//}
