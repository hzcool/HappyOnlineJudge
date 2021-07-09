package dao

import (
	"HappyOnlineJudge/common"
	"HappyOnlineJudge/model"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"
)

const (
	SUBMISSION_REDIS_EXPIRE = time.Hour * 5
)

type (
	Submission = model.Submission
)

type SubmissionDao struct {
	ID         int64
	Submission *Submission
}

func (sd *SubmissionDao) GetTableName() string {
	return "submission"
}

func (sd *SubmissionDao) GetRedisExpire() time.Duration {
	return SUBMISSION_REDIS_EXPIRE
}

func (sd *SubmissionDao) GetSelf() interface{} {
	if sd.Submission == nil {
		sd.Submission = &Submission{}
	}
	return sd.Submission
}

func (sd *SubmissionDao) GetID() int64 {
	if sd.ID == 0 {
		if sd.Submission != nil {
			sd.ID = sd.Submission.ID
		}
	}
	return sd.ID
}
func (sd *SubmissionDao) GetRedisKey() string {
	return sd.GetTableName() + "_" + strconv.FormatInt(sd.GetID(), 10)
}

func PutZSetTORedis(aid int64, pid int64) {
	key := getSubZSetKey(aid, pid)
	if rdb.Exists(ctx, key).Val() <= 0 {
		ids := make([]int64, 0)
		sd := &SubmissionDao{}
		if err := engine.Table(sd.GetTableName()).Where("author_id = ? and problem_id = ?", aid, pid).Cols("id").Find(&ids); err != nil {
			return
		}
		zs := make([]*redis.Z, len(ids))
		for i, id := range ids {
			zs[i] = &redis.Z{Score: float64(-id), Member: id}
		}
		rdb.ZAdd(ctx, key, zs...)
	}
}
func GetSubZSet(aid int64, pid int64) []int64 {
	key := getSubZSetKey(aid, pid)
	if rdb.Exists(ctx, key).Val() <= 0 {
		ids := make([]int64, 0)
		sd := &SubmissionDao{}
		if err := engine.Table(sd.GetTableName()).Where("author_id = ? and problem_id = ?", aid, pid).Cols("id").Desc("id").Find(&ids); err != nil {
			return nil
		}
		zs := make([]*redis.Z, len(ids))
		for i, id := range ids {
			zs[i] = &redis.Z{Score: float64(-id), Member: id}
		}
		rdb.ZAdd(ctx, key, zs...)
		return ids
	}
	idsStr := rdb.ZRange(ctx, key, 0, -1).Val()
	rdb.Expire(ctx, key, SUBMISSION_REDIS_EXPIRE)
	ids := make([]int64, len(idsStr))
	for i, _ := range ids {
		ids[i] = common.StrToInt64(idsStr[i])
	}
	return ids
}
func (sd *SubmissionDao) BeforePutToRedis() error {
	key := sd.getSubZSetKey()
	if rdb.Exists(ctx, key).Val() > 0 {
		rdb.ZAdd(ctx, key, &redis.Z{
			Score:  float64(-sd.GetID()),
			Member: sd.GetID(),
		})
	} else {
		ids := make([]int64, 0)
		if err := engine.Table(sd.GetTableName()).Where("author_id = ? and problem_id = ?", sd.Submission.AuthorID, sd.Submission.ProblemID).Cols("id").Find(&ids); err != nil {
			return err
		}
		zs := make([]*redis.Z, len(ids))
		for i, id := range ids {
			zs[i] = &redis.Z{Score: float64(-id), Member: id}
		}
		rdb.ZAdd(ctx, key, zs...)
	}
	rdb.Expire(ctx, key, SUBMISSION_REDIS_EXPIRE)
	return nil
}
func (sd *SubmissionDao) BeforeDelete() error {
	return nil
}

func (sd *SubmissionDao) getSubZSetKey() string { //某个人关于某道题的提交zset
	return "u:" + strconv.FormatInt(sd.Submission.AuthorID, 10) + "_p:" + strconv.FormatInt(sd.Submission.ProblemID, 10) + "_sub_zset"
}
func getSubZSetKey(aid int64, pid int64) string {
	return "u:" + strconv.FormatInt(aid, 10) + "_p:" + strconv.FormatInt(pid, 10) + "_sub_zset"
}

func HandleJudgeResult(isAC bool, ud *UserDao, pd *ProblemDao) {
	cols := Cols(ud, "passed_problems", "failed_problems", "passed_count", "passed_sub_count", "all_sub_count")
	passed := cols[0].ToStringMapInt64()
	failed := cols[1].ToStringMapInt64()
	passed_count := cols[2].ToUint()
	passed_sub_count := cols[3].ToUint()
	all_sub_count := cols[4].ToUint() + 1
	index := pd.GetIndex()

	pcols := Cols(pd, "accepted_count", "all_count")
	accepted_count := pcols[0].ToUint()
	all_count := pcols[1].ToUint() + 1

	if isAC {
		accepted_count++
		passed_sub_count++
		if _, ok := passed[index]; !ok {
			passed_count++
			passed[index] = pd.GetID()
			delete(failed, index)
			UpdateCols(ud, common.H{"passed_problems": passed, "failed_problems": failed, "passed_count": passed_count, "passed_sub_count": passed_sub_count, "all_sub_count": all_sub_count})
		} else {
			UpdateCols(ud, common.H{"passed_sub_count": passed_sub_count, "all_sub_count": all_sub_count})
		}
	} else {
		_, ok1 := passed[index]
		_, ok2 := failed[index]
		if !ok1 && !ok2 {
			failed[index] = pd.GetID()
			UpdateCols(ud, common.H{"failed_problems": failed, "all_sub_count": all_sub_count})
		} else {
			UpdateCols(ud, common.H{"all_sub_count": all_sub_count})
		}
	}

	rdb.ZAdd(ctx, USER_ZSET_KEY, &redis.Z{
		Score:  -(float64(passed_count)*1000000000 + float64(passed_sub_count)/float64(all_sub_count)),
		Member: ud.GetID(),
	})

	//更新问题过题数
	problem_update_mutex.Lock()
	UpdateCols(pd, common.H{"accepted_count": accepted_count, "all_count": all_count})
	problem_update_mutex.Unlock()
}

func (sd *SubmissionDao) Judge(pd *ProblemDao) {
	chID := <-common.CH
	defer func() {
		common.CH <- chID
	}()
	UpdateCols(sd, common.H{"status": "Running"})
	cols := Cols(pd, "time_limit", "memory_limit", "index", "accepted_count", "all_count")
	update := common.ToJudge(common.H{
		"lang":         sd.Submission.Lang,
		"max_cpu_time": cols[0].ToUint(),
		"max_memory":   cols[1].ToUint() * 1024 * 1024,
		"test_case":    cols[2].ToString(),
		"src":          sd.Submission.Code,
	})
	UpdateCols(sd, update)
	ud := &UserDao{ID: sd.Submission.AuthorID}
	HandleJudgeResult(update["status"].(string) == "Accepted", ud, pd)
}

func SearchSubmissions(l, r int64, rules []string, values []interface{}) []Submission {
	ss := make([]Submission, 0)
	if len(rules) > 0 {
		engine.Desc("id").Where(ToSqlConditions(rules), values...).Limit(int(r-l+1), int(l-1)).Omit("code", "compile_info").Find(&ss)
	} else {
		engine.Desc("id").Limit(int(r-l+1), int(l-1)).Omit("code", "compile_info").Find(&ss)
	}
	return ss
}

func GetUserSubCondition(authorID int64) map[string]uint {
	status := make([]string, 0)
	data := map[string]uint{
		"Accepted":            0,
		"WrongAnswer":         0,
		"TimeLimitExceeded":   0,
		"MemoryLimitExceeded": 0,
		"OutputLimitExceeded": 0,
		"RuntimeError":        0,
		"SystemError":         0,
	}
	engine.Table("submission").Where("author_id = ?", authorID).Cols("status").Find(&status)
	for _, item := range status {
		data[item]++
	}
	return data
}

func AuthScanProblemSubmission(ud *UserDao, sd *SubmissionDao) bool {
	if ud.GetID() == sd.Submission.AuthorID || OneCol(ud, "is_super_admin").ToBool() {
		return true
	}
	return false
}
