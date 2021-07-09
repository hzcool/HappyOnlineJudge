package dao

import (
	"HappyOnlineJudge/model"
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	"strconv"
)

type (
	Class    = model.Class
	Homework = model.Homework
	Student  = model.Student
)

/*
	zset 班级
*/

func getClassKey() string {
	return "class_zset"
}
func NewClass(name, password string) *Class {
	c := &Class{
		Name:     name,
		Password: password,
	}
	if num, err := engine.InsertOne(c); num != 1 || err != nil {
		return nil
	}
	js, _ := json.Marshal(c)
	rdb.ZAdd(ctx, getClassKey(), &redis.Z{Member: js, Score: float64(c.ID)})
	return c
}
func cacheClass() string {
	zkey := getClassKey()
	if rdb.Exists(ctx, zkey).Val() <= 0 {
		classes := make([]Class, 0)
		engine.Find(&classes)
		data := make([]*redis.Z, len(classes))
		for i, item := range classes {
			js, _ := json.Marshal(item)
			data[i] = &redis.Z{
				Score:  float64(item.ID),
				Member: js,
			}
		}
		rdb.ZAdd(ctx, zkey, data...)
	}
	return zkey
}
func GetClasses() []Class {
	zkey := cacheClass()
	x := rdb.ZRange(ctx, zkey, 0, -1).Val()
	data := make([]Class, len(x))
	for i, item := range x {
		json.Unmarshal([]byte(item), &data[i])
	}
	return data
}

func UpdateClassInfo(id int64, name, password string) {
	class := &Class{
		ID:       id,
		Name:     name,
		Password: password,
	}
	engine.ID(class.ID).Update(class)
	js, _ := json.Marshal(class)
	zkey := cacheClass()
	idStr := strconv.FormatInt(id, 10)
	rdb.ZRemRangeByScore(ctx, zkey, idStr, idStr)
	rdb.ZAdd(ctx, zkey, &redis.Z{Member: js, Score: float64(class.ID)})
}

func AddStudent(uid, cid int64) error {
	student := &Student{
		UserID:  uid,
		ClassID: cid,
	}
	if num, err := engine.InsertOne(student); num != 1 || err != nil {
		return errors.New("操作失败")
	}
	return nil
}
