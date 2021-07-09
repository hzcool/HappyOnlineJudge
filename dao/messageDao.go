package dao

import (
	"HappyOnlineJudge/common"
	"HappyOnlineJudge/model"
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"
)

const (
	MESSAGE_REDIS_EXPIRE = time.Hour * 5
	NEWMSG_COUNT_KEY     = "new_msg_count"
)

type (
	Message = model.Message
)

/*
redis 结构说明
Conversation 会话: 只缓存opened的会话
zset 有新消息的数量排序 member:to
key: opened_con_of_user:id

Message: 信息
List: 最近的信息靠前
key: message_list_of_con:id
*/

func getConversationRedisKey(userID int64) string {
	return "opened_con_of_user:" + strconv.FormatInt(userID, 10)
}

func ChangeToOpened(from int64, to int64, num int64) {
	key := getConversationRedisKey(from)
	member := strconv.FormatInt(to, 10)
	if num > 0 {
		ud := &UserDao{ID: from}
		rdb.HIncrBy(ctx, ud.GetRedisKey(), NEWMSG_COUNT_KEY, num)
	}
	rdb.ZIncrBy(ctx, key, float64(num), member)
}

func ChangeToClosed(from int64, to int64) {
	key := getConversationRedisKey(from)
	member := strconv.FormatInt(to, 10)
	if num := int64(rdb.ZScore(ctx, key, member).Val()); num > 0 {
		ud := &UserDao{ID: from}
		rdb.HIncrBy(ctx, ud.GetRedisKey(), NEWMSG_COUNT_KEY, -num)
	}
	rdb.ZRem(ctx, key, member)
}

func GetContacts(from int64) []redis.Z {
	key := getConversationRedisKey(from)
	return rdb.ZRevRangeWithScores(ctx, key, 0, -1).Val()
}

func getConversationID(a int64, b int64) int64 {
	if a > b {
		a, b = b, a
	}
	return (a << 32) + b
}

func getMessageRedisKey(cid int64) string {
	return "message_list_of_con:" + strconv.FormatInt(cid, 10)
}

func SendOneMessage(from int64, to int64, content string) (error, time.Time) {
	cid := getConversationID(from, to)
	msg := &Message{ConversationID: cid, From: from, To: to, Content: content}
	if num, err := engine.Insert(msg); err != nil || num != 1 {
		return errors.New("数据库插入消息失败"), time.Now()
	}
	ChangeToOpened(from, to, 0)
	ChangeToOpened(to, from, 1)
	key := getMessageRedisKey(cid)
	js, _ := json.Marshal(msg)
	rdb.ZAdd(ctx, key, &redis.Z{
		Score:  float64(msg.ID),
		Member: js,
	})
	rdb.Expire(ctx, key, MESSAGE_REDIS_EXPIRE)
	return nil, msg.CreatedAt
}

func GetMessages(a int64, b int64, l int64, r int64) (error, []Message) { //a获得关于b的messages
	cid := getConversationID(a, b)
	key := getMessageRedisKey(cid)
	messages := make([]Message, 0)
	exist := rdb.Exists(ctx, key).Val() > 0
	tot := int64(0)
	if exist {
		tot = rdb.ZCount(ctx, key, "-inf", "+inf").Val()
	}
	if tot < r {
		need := r - tot
		if need < 50 && exist {
			need = 50
		}
		id := int64(100000000000000)
		if exist {
			id = int64(rdb.ZRangeWithScores(ctx, key, 0, 0).Val()[0].Score)
		}
		if err := engine.Where("conversation_id = ? and id < ?", cid, id).Desc("id").Limit(int(need)).Find(&messages); err != nil {
			return err, nil
		}
		zarr := make([]*redis.Z, len(messages))
		for index, item := range messages {
			zarr[index] = &redis.Z{Score: float64(item.ID)}
			zarr[index].Member, _ = json.Marshal(item)
		}
		rdb.ZAdd(ctx, key, zarr...)
		rdb.Expire(ctx, key, MESSAGE_REDIS_EXPIRE)
	}
	s := rdb.ZRange(ctx, key, -r, -l).Val()
	messages = make([]Message, len(s))
	for idx, item := range s {
		json.Unmarshal([]byte(item), &messages[idx])
	}
	return nil, messages
}

func ClearOnePersonUnread(from int64, to int64) {
	key := getConversationRedisKey(from)
	member := strconv.FormatInt(to, 10)
	if num := int64(rdb.ZScore(ctx, key, member).Val()); num > 0 {
		ud := &UserDao{ID: from}
		rdb.HIncrBy(ctx, ud.GetRedisKey(), NEWMSG_COUNT_KEY, -num)
		rdb.ZIncrBy(ctx, key, float64(-num), member)
	}
}

func MessageCount(username string) int {
	ud := &UserDao{Username: username}
	s := rdb.HGet(ctx, ud.GetRedisKey(), NEWMSG_COUNT_KEY).Val()
	if s == "" {
		return 0
	}
	return common.StrToInt(s)
}
