package dao

import (
	"HappyOnlineJudge/common"
	"HappyOnlineJudge/model"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"strconv"
	sync2 "sync"
	"time"
)

type (
	Contest     = model.Contest
	Team        = model.Team
	Cproblem    = model.Cproblem
	CSubmission = model.Csubmission
)

var (
	pending, running    map[int64]*time.Timer
	teamMutex           map[int64]*sync2.Mutex
	teamCreateMutex     *sync2.Mutex
	cproblemMutex       map[int64]*sync2.Mutex
	cproblemCreateMutex *sync2.Mutex
)

func contestInit() {
	pending, running = make(map[int64]*time.Timer), make(map[int64]*time.Timer)

	contest := make([]Contest, 0)
	engine.Find(&contest)
	for idx, _ := range contest {
		cd := &ContestDao{Contest: &contest[idx]}
		cd.PutContest()
	}
	teamMutex = make(map[int64]*sync2.Mutex)
	teamCreateMutex = new(sync2.Mutex)
	cproblemMutex = make(map[int64]*sync2.Mutex)
	cproblemCreateMutex = new(sync2.Mutex)
}

/*
	team_zset: 排名分数 , userid
	team_hash: userid : jsonInfo
	cpoblems_list: json,

	cid_submission_zset: score:runid, jsInfo
	cid_uid_submission_hash: label:runid_list
*/

const (
	CONTEST_REDIS_EXPIRE = time.Hour * 24
)

type ContestDao struct {
	ID      int64
	Contest *Contest
}

func (cd *ContestDao) GetTableName() string {
	return "contest"
}
func (cd *ContestDao) GetRedisExpire() time.Duration {
	return CONTEST_REDIS_EXPIRE
}
func (cd *ContestDao) GetSelf() interface{} {
	if cd.Contest == nil {
		cd.Contest = &Contest{}
	}
	return cd.Contest
}
func (cd *ContestDao) GetID() int64 {
	if cd.ID == 0 {
		if cd.Contest != nil {
			cd.ID = cd.Contest.ID
		}
	}
	return cd.ID
}
func (cd *ContestDao) GetRedisKey() string {
	return cd.GetTableName() + "_" + strconv.FormatInt(cd.GetID(), 10)
}
func (cd *ContestDao) BeforePutToRedis() error {
	cd.PutContest()
	return nil
}
func (cd *ContestDao) BeforeDelete() error {
	return nil
}
func (cd *ContestDao) Create() error {
	duration := cd.Contest.End.Sub(cd.Contest.Begin)
	cd.Contest.Length = uint(duration / time.Second)
	if err := Create(cd); err != nil {
		return err
	}
	cd.PutContest()
	return nil
}
func (cd *ContestDao) Delete() error {
	RemoveContest(cd.GetID())
	return Delete(cd)
}

func (cd *ContestDao) Update(mp common.H) error {
	if err := UpdateCols(cd, mp); err != nil {
		return err
	}
	_, ok1 := mp["begin"]
	_, ok2 := mp["end"]
	if ok1 || ok2 {
		GetSelfAll(cd)
		UpdateCols(cd, common.H{"length": uint(cd.Contest.End.Sub(cd.Contest.Begin) / time.Second)})
		cd.PutContest()
	}
	return nil
}

func removeContestMutex(id int64) {
	cps := GetCproblems(id)
	for _, cp := range cps {
		delete(cproblemMutex, cp.ID)
	}
	zkey := getTeamZsetKey(id)
	ids := rdb.ZRange(ctx, zkey, 0, -1).Val()
	for _, idStr := range ids {
		tid := (id << 32) | common.StrToInt64(idStr)
		delete(teamMutex, tid)
	}
}

func RemoveContest(id int64) {
	if _, ok := pending[id]; ok {
		pending[id].Stop()
		delete(pending, id)
	}
	if _, ok := running[id]; ok {
		running[id].Stop()
		delete(running, id)
	}
}

func (cd *ContestDao) PutContest() {
	c := cd.Contest
	RemoveContest(c.ID)
	now := time.Now()
	if now.After(c.End) {
		if c.Status != "Ended" {
			UpdateCols(cd, common.H{"status": "Ended"})
			time.AfterFunc(time.Hour*2, func() {
				removeContestMutex(c.ID)
			})
		}
		return
	}
	if now.Before(c.Begin) {
		if c.Status != "Pending" {
			UpdateCols(cd, common.H{"status": "Pending"})
		}
		pending[c.ID] = time.AfterFunc(c.Begin.Sub(now), func() {
			fmt.Println("比赛", c.ID, "已经开始")
			cd.PutContest()
		})
		return
	}
	if c.Status != "Running" {
		UpdateCols(cd, common.H{"status": "Running"})
	}
	running[c.ID] = time.AfterFunc(c.End.Sub(c.Begin), func() {
		fmt.Println("比赛", c.ID, "已经结束")
		cd.PutContest()
	})
}
func SearchContests(l, r int64, rules []string, values []interface{}) (int64, []Contest) {
	cs := make([]Contest, 0)
	var total int64
	if len(rules) > 0 {
		engine.Desc("id").Where(ToSqlConditions(rules), values...).Limit(int(r-l+1), int(l-1)).Find(&cs)
		total, _ = engine.Desc("id").Where(ToSqlConditions(rules), values...).Count(new(Contest))
	} else {
		engine.Desc("id").Limit(int(r-l+1), int(l-1)).Find(&cs)
		total, _ = engine.Desc("id").Count(new(Contest))
	}
	return total, cs
}

func getTeamZsetKey(cid int64) string {
	return "c_" + strconv.FormatInt(cid, 10) + "_team_zset"
}
func getTeamHashKey(cid int64) string {
	return "c_" + strconv.FormatInt(cid, 10) + "_team_hash"
}
func getCproblemListKey(cid int64) string {
	return "c_" + strconv.FormatInt(cid, 10) + "_cproblem_list"
}
func getCsubZsetKey(cid int64) string {
	return "c" + strconv.FormatInt(cid, 10) + "_sub_zset"
}
func getUserCsubHashKey(cid, uid int64) string {
	return "c" + strconv.FormatInt(cid, 10) + "_u" + strconv.FormatInt(uid, 10) + "_hash"
}
func getTeamRankScore(t *Team, format string) float64 {
	score := -float64(t.Scores)
	if format == "ACM" {
		score = -float64(t.Solved*100000) + float64(t.Penalty)
	}
	return score
}
func teamCache(cid int64) string {
	zkey := getTeamZsetKey(cid)
	hkey := getTeamHashKey(cid)
	if rdb.Exists(ctx, zkey, hkey).Val() < 2 {
		cd := &ContestDao{ID: cid}
		format := OneCol(cd, "format").ToString()
		x := make([]Team, 0)
		engine.Where("contest_id = ?", cid).Find(&x)
		for _, t := range x {
			rdb.ZAdd(ctx, zkey, &redis.Z{
				Score:  getTeamRankScore(&t, format),
				Member: t.UserID,
			})
			bt, _ := json.Marshal(t)
			rdb.HSet(ctx, hkey, t.UserID, bt)
		}
		rdb.Expire(ctx, zkey, CONTEST_REDIS_EXPIRE)
		rdb.Expire(ctx, hkey, CONTEST_REDIS_EXPIRE)
	}
	return hkey
}
func cproblemCache(cid int64) string {
	lkey := getCproblemListKey(cid)
	if rdb.Exists(ctx, lkey).Val() <= 0 {
		cp := make([]Cproblem, 0)
		engine.Where("contest_id = ?", cid).Find(&cp)
		js := make([]interface{}, len(cp))
		for i, item := range cp {
			js[i], _ = json.Marshal(item)
		}
		rdb.RPush(ctx, lkey, js...)
		rdb.Expire(ctx, lkey, CONTEST_REDIS_EXPIRE)
	}
	return lkey
}
func csubmissionCache(cid int64) string {
	zkey := getCsubZsetKey(cid)
	if rdb.Exists(ctx, zkey).Val() <= 0 {
		data := make([]CSubmission, 0)
		engine.Where("contest_id = ?", cid).Find(&data)
		zarr := make([]*redis.Z, len(data))
		for i, item := range data {
			js, _ := json.Marshal(item)
			zarr[i] = &redis.Z{
				Score:  float64(item.RunID),
				Member: js,
			}
		}
		rdb.ZAdd(ctx, zkey, zarr...)
		rdb.Expire(ctx, zkey, CONTEST_REDIS_EXPIRE)
	}
	return zkey
}
func csubmissionOfUserCache(cid, uid int64) string {
	csubmissionCache(cid)
	hkey := getUserCsubHashKey(cid, uid)
	if rdb.Exists(ctx, hkey).Val() <= 0 {
		data := make([]CSubmission, 0)
		engine.Where("contest_id = ? and user_id = ?", cid, uid).Find(&data)
		mp := make(common.H)
		for _, item := range data {
			if tmp, ok := mp[item.Label].([]uint); ok {
				mp[item.Label] = append(tmp, item.RunID)
			} else {
				mp[item.Label] = []uint{item.RunID}
			}
		}
		HMSetMap(hkey, mp, CONTEST_REDIS_EXPIRE)
	}
	return hkey
}
func GetCproblems(cid int64) []Cproblem {
	cproblemCache(cid)
	data := rdb.LRange(ctx, getCproblemListKey(cid), 0, -1).Val()
	cps := make([]Cproblem, len(data))
	for i, item := range data {
		json.Unmarshal([]byte(item), &cps[i])
	}
	return cps
}

func DeleteCproblems(cid int64) error {
	if _, err := engine.Exec("delete from cproblem where contest_id = ?", cid); err != nil {
		return err
	}
	rdb.Del(ctx, getCproblemListKey(cid))
	return nil
}

func AddCproblems(cps []Cproblem) error {
	if _, err := engine.Insert(cps); err != nil {
		return err
	}
	return nil
}

func CountTeams(cid int64) int64 {
	teamCache(cid)
	zkey := getTeamZsetKey(cid)
	if rdb.Exists(ctx, zkey).Val() > 0 {
		return rdb.ZCount(ctx, zkey, "-inf", "+inf").Val()
	}
	total, _ := engine.Where("contest_id = ?", cid).Count(&Team{})
	return total
}
func ExistTeam(cid int64, uid int64) bool {
	hkey := teamCache(cid)
	if rdb.Exists(ctx, hkey).Val() > 0 {
		return rdb.HExists(ctx, hkey, strconv.FormatInt(uid, 10)).Val()
	}
	exist, _ := engine.Where("contest_id = ? and user_id = ?", cid, uid).Exist(&Team{})
	return exist
}
func GetTeam(cid, uid int64) *Team {
	hkey := teamCache(cid)
	t := &Team{}
	json.Unmarshal([]byte(rdb.HGet(ctx, hkey, strconv.FormatInt(uid, 10)).Val()), t)
	return t
}
func NewTeam(cid, uid int64) error {
	teamCache(cid)
	t := &Team{ContestID: cid, UserID: uid}
	if num, err := engine.InsertOne(t); err != nil || num != 1 {
		return errors.New("操作失败")
	}
	rdb.ZAdd(ctx, getTeamZsetKey(cid), &redis.Z{
		Score:  0,
		Member: uid,
	})
	js, _ := json.Marshal(t)
	rdb.HSet(ctx, getTeamHashKey(cid), uid, js)
	return nil
}
func GetOneTeamStatus(cid, uid int64) map[string]map[string]uint {
	if uid == 0 {
		return make(map[string]map[string]uint)
	}
	teamCache(cid)
	key := getTeamHashKey(cid)
	t := &Team{}
	field := strconv.FormatInt(uid, 10)
	if rdb.HExists(ctx, key, field).Val() {
		json.Unmarshal([]byte(rdb.HGet(ctx, key, field).Val()), t)
	}
	if t.ProblemStatus == nil {
		return make(map[string]map[string]uint)
	}
	return t.ProblemStatus
}
func GetAllTeamRankData(cid int64) []common.H {
	hkey := teamCache(cid)
	zkey := getTeamZsetKey(cid)
	z := rdb.ZRangeWithScores(ctx, zkey, 0, -1).Val()
	rk := 0
	lastScore := int64(-1)
	data := make([]common.H, len(z))
	cp := make(map[string]*Cproblem)
	for i, item := range z {
		uid := common.StrToInt64(item.Member.(string))
		nowScore := int64(item.Score)
		if nowScore != lastScore {
			rk++
			lastScore = nowScore
		}
		team := &Team{}
		json.Unmarshal([]byte(rdb.HGet(ctx, hkey, strconv.FormatInt(uid, 10)).Val()), team)
		ud := &UserDao{ID: team.UserID}
		info := common.H{
			"rank":    rk,
			"team":    OneCol(ud, "username").ToString(),
			"score":   team.Scores,
			"penalty": team.Penalty,
			"solved":  team.Solved,
		}
		for k, v := range team.ProblemStatus {
			info[k] = v
			if v["score"] < 100 {
				info[k+"_"] = "WA"
			} else {
				info[k+"_"] = "AC"
				if cp[k] == nil {
					cp[k] = GetCproblem(cid, k)
				}
				if cp[k] != nil && cp[k].FirstSolveTime == v["minutes"] {
					info[k+"_"] = "FirstBlood"
				}
			}
		}
		data[i] = info
	}
	return data
}
func GetLabels(cid int64) []string {
	lkey := cproblemCache(cid)
	d := rdb.LLen(ctx, lkey).Val()
	labels := make([]string, d)
	for i := int64(0); i < d; i++ {
		labels[i] = string([]byte{byte(65 + i)})
	}
	return labels
}
func GetCproblemTitles(cid int64) ([]string, []string) {
	lkey := cproblemCache(cid)
	js := rdb.LRange(ctx, lkey, 0, -1).Val()
	d := len(js)
	titles := make([]string, d)
	labels := make([]string, d)
	for i, item := range js {
		cp := &Cproblem{}
		json.Unmarshal([]byte(item), cp)
		titles[i] = cp.Title
		labels[i] = cp.Label
	}
	return labels, titles
}

func GetCproblem(cid int64, label string) *Cproblem {
	lkey := cproblemCache(cid)
	p := int64(label[0]) - int64('A')
	ret := &Cproblem{}
	if err := json.Unmarshal([]byte(rdb.LIndex(ctx, lkey, p).Val()), ret); err != nil {
		return nil
	}
	return ret
}

func SearchCsubByRedis(cid, l, r int64) (int64, []CSubmission) {
	zkey := csubmissionCache(cid)
	arr := rdb.ZRange(ctx, zkey, -r, -l).Val()
	cs := make([]CSubmission, len(arr))
	d := len(arr)
	for i, item := range arr {
		json.Unmarshal([]byte(item), &cs[d-i-1])
	}
	return rdb.ZCard(ctx, zkey).Val(), cs
}

func SearchCsubmissions(l, r int64, rules []string, values []interface{}) (int64, []CSubmission) {
	if len(rules) == 1 && rules[0] == "contest_id" {
		return SearchCsubByRedis(values[0].(int64), l, r)
	}
	cs := make([]CSubmission, 0)
	var total int64
	engine.Desc("id").Where(ToSqlConditions(rules), values...).Omit("code", "compile_info").Limit(int(r-l+1), int(l-1)).Find(&cs)
	total, _ = engine.Desc("id").Where(ToSqlConditions(rules), values...).Count(new(CSubmission))
	return total, cs
}

func GetCsubmission(run_id, cid int64) *CSubmission {
	cs := new(CSubmission)
	zkey := csubmissionCache(cid)
	score := strconv.FormatInt(int64(run_id), 10)
	arr := rdb.ZRangeByScore(ctx, zkey, &redis.ZRangeBy{
		Min: score,
		Max: score,
	}).Val()
	if len(arr) == 0 {
		return nil
	}
	json.Unmarshal([]byte(arr[0]), cs)
	return cs
}
func GetUserOneProblemCsub(cid, uid int64, label string) []CSubmission {
	csubmissionOfUserCache(cid, uid)
	hkey := getUserCsubHashKey(cid, uid)
	if !rdb.HExists(ctx, hkey, label).Val() {
		return make([]CSubmission, 0)
	}
	ids := make([]uint, 0)
	json.Unmarshal([]byte(rdb.HGet(ctx, hkey, label).Val()), &ids)
	d := len(ids)
	data := make([]CSubmission, d)
	zkey := getCsubZsetKey(cid)
	for i, item := range ids {
		score := strconv.FormatInt(int64(item), 10)
		cs := rdb.ZRangeByScore(ctx, zkey, &redis.ZRangeBy{
			Min: score,
			Max: score,
		}).Val()[0]
		json.Unmarshal([]byte(cs), &data[d-i-1])
	}
	return data
}

func AddOneCsubmission(cid, uid int64, label, code, lang string) *CSubmission {
	hkey := csubmissionOfUserCache(cid, uid)
	zkey := getCsubZsetKey(cid)
	runID := uint(rdb.ZCard(ctx, zkey).Val() + 1)
	cs := &CSubmission{
		RunID:     runID,
		ContestID: cid,
		UserID:    uid,
		Code:      code,
		Label:     label,
		Length:    uint(len(code)),
		Status:    "Accepted",
		Lang:      lang,
	}
	if num, err := engine.InsertOne(cs); num != 1 || err != nil {
		return nil
	}
	js, _ := json.Marshal(cs)
	rdb.ZAdd(ctx, zkey, &redis.Z{
		Score:  float64(runID),
		Member: js,
	})
	arr := make([]uint, 0)
	if rdb.HExists(ctx, hkey, label).Val() {
		json.Unmarshal([]byte(rdb.HGet(ctx, hkey, label).Val()), &arr)
	}
	arr = append(arr, runID)
	js2, _ := json.Marshal(arr)
	rdb.HSet(ctx, hkey, label, js2)
	return cs
}

func UpdateCsub(cs *CSubmission, update common.H) { //cs自己已更新
	score := strconv.FormatInt(int64(cs.RunID), 10)
	zkey := csubmissionCache(cs.ContestID)
	rdb.ZRemRangeByScore(ctx, zkey, score, score)
	js, _ := json.Marshal(cs)
	rdb.ZAdd(ctx, zkey, &redis.Z{
		Score:  float64(cs.RunID),
		Member: js,
	})
	UpdateColsBySql(cs.GetTableName(), cs.ID, update)
}

func ContestJudger(cs *CSubmission) {
	chID := <-common.CH
	defer func() {
		common.CH <- chID
	}()

	//1:测试前更新状态
	cs.Status = "Running"
	UpdateCsub(cs, common.H{"status": cs.Status})

	//2测评
	cp := GetCproblem(cs.ContestID, cs.Label)
	pd := &ProblemDao{ID: cp.ProblemID}
	cd := &ContestDao{ID: cs.ContestID}
	cols := Cols(pd, "time_limit", "memory_limit", "index")
	update := common.ToJudge(common.H{
		"lang":         cs.Lang,
		"max_cpu_time": cols[0].ToUint(),
		"max_memory":   cols[1].ToUint() * 1024 * 1024,
		"test_case":    cols[2].ToString(),
		"src":          cs.Code,
	})

	//3更新测评数据
	if cs.UserID != 1 && !OneCol(cd, "end").ToTime().Before(cs.CreatedAt) { //不能是super_admin提交、比赛结束后提交

		dur := uint(cs.CreatedAt.Sub(OneCol(cd, "begin").ToTime()).Seconds()) //秒
		newScore := update["score"].(uint)

		//更新 cp
		cproblemCreateMutex.Lock()
		if _, ok := cproblemMutex[cp.ID]; !ok {
			cproblemMutex[cp.ID] = new(sync2.Mutex)
		}
		cproblemCreateMutex.Unlock()
		cproblemMutex[cp.ID].Lock()
		cp = GetCproblem(cs.ContestID, cs.Label)
		cp.All++
		if newScore == 100 {
			if cp.AC == 0 || dur/60 < cp.FirstSolveTime {
				cp.FirstSolveTime = dur / 60
			}
			cp.AC++
		}
		UpdateColsBySql(cp.GetTableName(), cp.ID, common.H{"first_solve_time": cp.FirstSolveTime, "ac": cp.AC, "all": cp.All})
		idx := int64(cs.Label[0]) - int64('A')
		js, _ := json.Marshal(cp)
		rdb.LSet(ctx, getCproblemListKey(cs.ContestID), idx, js)
		cproblemMutex[cp.ID].Unlock()

		//更新team
		tkey := (cs.ContestID << 32) | (cs.UserID)
		teamCreateMutex.Lock()
		if _, ok := teamMutex[tkey]; !ok {
			teamMutex[tkey] = new(sync2.Mutex)
		}
		teamCreateMutex.Unlock()
		teamMutex[tkey].Lock()
		team := GetTeam(cs.ContestID, cs.UserID)
		if team.ProblemStatus == nil {
			team.ProblemStatus = make(map[string]map[string]uint)
		}
		if _, ok := team.ProblemStatus[cs.Label]; !ok {
			team.ProblemStatus[cs.Label] = map[string]uint{
				"fail_times": 0,
				"minutes":    0,
				"score":      0,
				"last":       0,
			}
		}
		mp := team.ProblemStatus[cs.Label]
		orgScore := mp["score"]
		last := mp["last"]
		teamNeedUpdated := true
		if newScore < 100 && orgScore < 100 {
			mp["fail_times"]++
			if last == 0 || last < dur {
				mp["last"] = dur
			}
			if orgScore < newScore {
				team.Scores += newScore - orgScore
				mp["score"] = newScore
			}
		} else if newScore < 100 {
			if dur < last {
				mp["fail_times"]++
				team.Penalty += 20
			} else {
				teamNeedUpdated = false
			}
		} else if dur >= last { //新分数100
			if orgScore < 100 {
				team.Scores += 100 - orgScore
				team.Solved++
				mp["score"] = 100
				mp["minutes"] = dur / 60
				mp["last"] = dur
				team.Penalty += mp["fail_times"]*20 + dur/60
			} else {
				teamNeedUpdated = false
			}
		} else {
			css := GetUserOneProblemCsub(cs.ContestID, cs.UserID, cs.Label)
			cost := make([]uint, len(css))
			for i, item := range css {
				cost[i] = uint(item.CreatedAt.Sub(OneCol(cd, "begin").ToTime()).Seconds())
				if item.Score == 100 && dur > cost[i] {
					dur = cost[i]
					mp["last"] = dur
				}
			}
			fails := uint(0)
			for _, item := range cost {
				if item < dur {
					fails++
				}
			}
			if orgScore < 100 {
				team.Scores += 100 - orgScore
				team.Solved++
				mp["score"] = 100
			} else {
				team.Penalty -= mp["fail_times"]*20 + mp["minutes"]
			}
			mp["fail_times"] = fails
			mp["minutes"] = dur / 60
			team.Penalty += mp["fail_times"]*20 + dur/60
		}
		if teamNeedUpdated {
			UpdateColsBySql(team.GetTableName(), team.ID, common.H{"solved": team.Solved, "penalty": team.Penalty, "scores": team.Scores, "problem_status": team.ProblemStatus})
			rdb.ZAdd(ctx, getTeamZsetKey(cs.ContestID), &redis.Z{
				Score:  getTeamRankScore(team, OneCol(cd, "format").ToString()),
				Member: team.UserID,
			})
			js, _ := json.Marshal(team)
			rdb.HSet(ctx, getTeamHashKey(cs.ContestID), team.UserID, js)
		}
		teamMutex[tkey].Unlock()
	}

	//更新cs
	cs.Score = update["score"].(uint)
	cs.Time = update["time"].(uint)
	cs.Memory = update["memory"].(uint)
	cs.CompileInfo = update["compile_info"].(string)
	cs.Status = update["status"].(string)
	UpdateCsub(cs, update)
}

func ReJudge(cid int64, labels []string) {
	cd := &ContestDao{ID: cid}
	format := OneCol(cd, "format").ToString()
	sql := "contest_id = ? and ("
	args := make([]interface{}, 1)
	args[0] = cid
	//更新problem
	for cnt, item := range labels {
		cp := GetCproblem(cid, item)
		if cp == nil {
			continue
		}
		mut := cproblemMutex[cp.ID]
		if mut != nil {
			mut.Lock()
		}

		if cnt == 0 {
			sql += "label = ?"
		} else {
			sql += " or label = ?"
		}
		args = append(args, item)
		UpdateColsBySql("cproblem", cp.ID, common.H{
			"first_solve_time": 0,
			"ac":               0,
			"all":              0,
		})
		cp.FirstSolveTime, cp.All, cp.AC = 0, 0, 0
		idx := int64(item[0]) - int64('A')
		js, _ := json.Marshal(cp)
		rdb.LSet(ctx, getCproblemListKey(cid), idx, js)
		if mut != nil {
			mut.Unlock()
		}
	}
	sql += ")"

	//更新team
	hkey := teamCache(cid)
	zkey := getTeamZsetKey(cid)
	ids := rdb.ZRange(ctx, zkey, 0, -1).Val()
	for _, uid := range ids {

		tMutKey := (cid << 32) | common.StrToInt64(uid)
		mut := teamMutex[tMutKey]
		if mut != nil {
			mut.Lock()
		}
		t := &Team{}
		if err := json.Unmarshal([]byte(rdb.HGet(ctx, hkey, uid).Val()), t); err == nil {
			if t.ProblemStatus != nil {
				had := false
				for _, label := range labels {
					if mp, ok := t.ProblemStatus[label]; ok {
						if mp["score"] == 100 {
							t.Solved--
							t.Penalty -= mp["fail_times"]*20 + mp["minutes"]
						}
						t.Scores -= mp["score"]
						delete(t.ProblemStatus, label)
						had = true
					}
				}
				if had {
					UpdateColsBySql(t.GetTableName(), t.ID, common.H{"solved": t.Solved, "penalty": t.Penalty, "scores": t.Scores, "problem_status": t.ProblemStatus})
					rdb.ZAdd(ctx, zkey, &redis.Z{
						Score:  getTeamRankScore(t, format),
						Member: t.UserID,
					})
					js, _ := json.Marshal(t)
					rdb.HSet(ctx, hkey, t.UserID, js)
				}
			}
		}

		if mut != nil {
			mut.Unlock()
		}
	}

	csubs := make([]CSubmission, 0)
	engine.Where(sql, args...).Find(&csubs)
	for i, _ := range csubs {
		csubs[i].Status = "Queueing"
		UpdateCsub(&csubs[i], common.H{"status": "Queueing"})
		go ContestJudger(&csubs[i])
	}
	if OneCol(cd, "end").ToTime().Before(time.Now()) {
		time.AfterFunc(time.Hour, func() {
			removeContestMutex(cid)
		})
	}
}
