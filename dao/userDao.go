package dao

import (
	"HappyOnlineJudge/common"
	"HappyOnlineJudge/model"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"
)

const (
	USER_REDIS_EXPIRE = 0 //用户在redis的超时时间,不过期
	USER_ZSET_KEY     = "user_zset(id)"
	USER_HASH_KEY     = "user_hash(name:id)"
)

/*
	user_zset: id 按过题数量
	user_hash: username:id
*/

type User = model.User

type UserDao struct {
	ID       int64
	Username string
	User     *User
}

func userInitRedis() {
	users := make([]User, 0)
	engine.Find(&users)
	for _, item := range users {
		ud := &UserDao{User: &item}
		PutToRedis(ud)
	}
}

func (ud *UserDao) GetRedisExpire() time.Duration {
	return USER_REDIS_EXPIRE
}
func (ud *UserDao) GetTableName() string {
	return "user"
}
func (ud *UserDao) GetSelf() interface{} {
	if ud.User == nil {
		ud.User = &User{}
	}
	return ud.User
}
func (ud *UserDao) GetName() string {
	if ud.Username == "" {
		if ud.User != nil && ud.User.Username != "" {
			ud.Username = ud.User.Username
		} else {
			ud.Username = OneCol(ud, "username").ToString()
		}
	}
	return ud.Username
}

//func (ud *UserDao) GetNameKey() string {
//	return ud.GetTableName() + "_" + ud.GetName()
//}
func (ud *UserDao) GetRedisKey() string { //必须有id
	return ud.GetTableName() + "_" + strconv.FormatInt(ud.GetID(), 10)
}
func (ud *UserDao) GetID() int64 {
	if ud.ID == 0 {
		if ud.User != nil && ud.User.ID != 0 {
			ud.ID = ud.User.ID
		} else {
			name := ud.Username
			if name == "" && ud.User != nil {
				name = ud.User.Username
			}
			if name != "" {
				if rdb.HExists(ctx, USER_HASH_KEY, name).Val() {
					ud.ID = common.StrToInt64(rdb.HGet(ctx, USER_HASH_KEY, name).Val())
				} else {
					x := new(Col)
					if ok, err := engine.SQL("select id from user where username = ?", name).Get(&x.data); err == nil && ok {
						ud.ID = x.ToInt64()
					}
				}
			}
		}
	}
	return ud.ID
}

func (ud *UserDao) BeforePutToRedis() error {
	score := float64(1)
	if ud.User.AllSubCount != 0 {
		score = float64(ud.User.PassedCount)*1000000000 + float64(ud.User.PassedSubCount)/float64(ud.User.AllSubCount)
	}
	rdb.ZAdd(ctx, USER_ZSET_KEY, &redis.Z{
		Score:  -score,
		Member: ud.GetID(),
	})
	rdb.HSet(ctx, USER_HASH_KEY, ud.GetName(), ud.GetID())
	return nil
}

func (ud *UserDao) BeforeDelete() error {
	rdb.ZRem(ctx, USER_ZSET_KEY, ud.GetID())
	rdb.HDel(ctx, USER_HASH_KEY, ud.GetName())
	return nil
}

func (ud *UserDao) Create() error {
	return Create(ud)
}
func (ud *UserDao) Delete()  {
	
}

func (ud *UserDao) Update(mp map[string]interface{}) error {
	if err := UpdateCols(ud, mp); err != nil {
		return err
	}
	if newName, ok := mp["username"]; ok {
		ud.Username = newName.(string)
		ud.BeforePutToRedis()
	}
	return nil
}
func CountUsers() int64 {
	return rdb.ZCount(ctx, USER_ZSET_KEY, "-inf", "+inf").Val()
}

func GetUsers(l, r int64) []int64 {
	ids := rdb.ZRange(ctx, USER_ZSET_KEY, l-1, r-1).Val()
	ret := make([]int64, len(ids))
	for i, id := range ids {
		ret[i] = common.StrToInt64(id)
	}
	return ret
}
func GetUsersByCreatedTime(l, r int64, cols []string) []User { //顺序获取
	ret := make([]User, 0)
	engine.Table("user").Limit(int(r-l+1), int(l-1)).Cols(cols...).Find(&ret)
	return ret
}

type UsersData struct {
	IDs   []int64
	Datas [][]Col
}

func GetTableName() string {
	return "user"
}
func (us *UsersData) GetIDs(cols []string, values []interface{}, a ...int) []int64 { //len(a)=0或2
	if len(a) == 0 {
		engine.Table("user").Where(ToSqlConditions(cols), values...).Cols("id").Find(&us.IDs)
	} else {
		engine.Table("user").Where(ToSqlConditions(cols), values...).Cols("id").Limit(a[0], a[1]).Find(&us.IDs)
	}
	return us.IDs
}
func (us *UsersData) GetColsByIDs(wants []string) [][]Col {
	for _, id := range us.IDs {
		us.Datas = append(us.Datas, Cols(&UserDao{ID: id}, wants...))
	}
	return us.Datas
}
