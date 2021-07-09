package dao

import (
	"HappyOnlineJudge/common"
	"HappyOnlineJudge/model"
	"archive/zip"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	sync2 "sync"
	"time"
)

const (
	PROBLEM_REDIS_EXPIRE   = 0
	PROBLEM_HASH_KEY       = "problem_hash(index:id)"
	OPEN_PROBLEM_ZSET_KEY  = "open_problem_zset"  //id作为分数
	CLOSE_PROBLEM_ZSET_KEY = "close_problem_zset" //id作为分数
)

var (
	problem_update_mutex sync2.Mutex
)

type (
	Problem = model.Problem
)

type ProblemDao struct {
	ID      int64
	Index   string
	Problem *Problem
}

func problemInitRedis() {
	problems := make([]Problem, 0)
	engine.Find(&problems)
	for _, item := range problems {
		ud := &ProblemDao{Problem: &item}
		PutToRedis(ud)
	}
}

func (pd *ProblemDao) GetTableName() string {
	return "problem"
}
func (pd *ProblemDao) GetRedisExpire() time.Duration {
	return PROBLEM_REDIS_EXPIRE
}
func (pd *ProblemDao) GetSelf() interface{} {
	if pd.Problem == nil {
		pd.Problem = &Problem{}
	}
	return pd.Problem
}
func (pd *ProblemDao) GetID() int64 {
	if pd.ID == 0 {
		if pd.Problem != nil && pd.Problem.ID != 0 {
			pd.ID = pd.Problem.ID
		} else {
			index := pd.Index
			if index == "" && pd.Problem != nil {
				index = pd.Problem.Index
			}
			if index != "" {
				if rdb.HExists(ctx, PROBLEM_HASH_KEY, index).Val() {
					pd.ID = common.StrToInt64(rdb.HGet(ctx, PROBLEM_HASH_KEY, index).Val())
				} else {
					x := new(Col)
					if ok, err := engine.SQL("select id from problem where index = ?", index).Get(&x.data); err == nil && ok {
						pd.ID = x.ToInt64()
					}
				}
			}
		}
	}
	return pd.ID
}
func (pd *ProblemDao) GetRedisKey() string {
	return pd.GetTableName() + "_" + strconv.FormatInt(pd.GetID(), 10)
}

func (pd *ProblemDao) GetIndex() string {
	if pd.Index == "" {
		if pd.Problem != nil && pd.Problem.Index != "" {
			pd.Index = pd.Problem.Index
		} else if pd.ID != 0 || (pd.Problem != nil && pd.Problem.ID != 0) {
			pd.Index = OneCol(pd, "index").ToString()
		}
	}
	return pd.Index
}

func (pd *ProblemDao) Created() error {
	return Create(pd)
}
func (pd *ProblemDao) BeforePutToRedis() error {
	rdb.HSet(ctx, PROBLEM_HASH_KEY, pd.GetIndex(), pd.GetID())
	rdb.ZAdd(ctx, GetProblemZSetKey(pd.Problem.IsOpen), &redis.Z{
		Score:  float64(pd.GetID()),
		Member: pd.Index,
	})
	return nil
}
func (pd *ProblemDao) BeforeDelete() error {
	isOpen := OneCol(pd, "is_open").ToBool()
	rdb.HDel(ctx, PROBLEM_HASH_KEY, pd.GetIndex())
	rdb.ZRem(ctx, GetProblemZSetKey(isOpen), pd.Index)
	return nil
}
func (pd *ProblemDao) Delete() error {
	return Delete(pd)
}
func nextIndex(index string) string {
	s := common.StrToInt64(index[1:]) + 1
	return index[0:1] + strconv.FormatInt(s, 10)
}
func (pd *ProblemDao) SetIndex() {
	problem := &Problem{}
	if has, _ := engine.Where("is_open = ?", pd.Problem.IsOpen).Desc("id").Cols("index").Get(problem); has {
		pd.Problem.Index = nextIndex(problem.Index)
	} else {
		if pd.Problem.IsOpen {
			pd.Problem.Index = "P1000"
		} else {
			pd.Problem.Index = "U1000"
		}
	}
	pd.Index = pd.Problem.Index
}

func (pd *ProblemDao) HandleZipData(dir string, zipPath string) error {
	zipReader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer zipReader.Close()
	input := make(map[int]*zip.File)
	output := make(map[int]*zip.File)
	for _, f := range zipReader.File {
		var id, base = 0, 1
		for i := strings.LastIndex(f.Name, ".") - 1; i >= 0; i-- {
			ch := int(f.Name[i]) - int('0')
			if ch < 0 || ch > 9 {
				break
			}
			id += base * ch
			base *= 10
		}
		if path.Ext(f.Name) == ".in" {
			input[id] = f
		} else if path.Ext(f.Name) == ".out" || path.Ext(f.Name) == ".ans" {
			output[id] = f
		}
	}
	testdatas, tmp := make(common.H), make(common.H)
	isSpj := OneCol(pd, "is_spj").ToBool()
	testdatas["spj"] = isSpj
	if isSpj {
		testdatas["spj_lang"] = OneCol(pd, "spj_type").ToString()
	}
	for k, in := range input {
		if out, ok := output[k]; ok {
			inPath := path.Join(dir, in.Name)
			outPath := path.Join(dir, out.Name)
			if err := common.StoreZipFile(in, inPath); err != nil {
				return err
			}
			if err := common.StoreZipFile(out, outPath); err != nil {
				return err
			}
			md5, err := common.MD5(outPath)
			if err != nil {
				return err
			}
			tmp[strconv.Itoa(k)] = common.H{
				"input_name":          in.Name,
				"output_name":         out.Name,
				"input_size":          common.FileSize(inPath),
				"output_size":         common.FileSize(outPath),
				"stripped_output_md5": md5,
			}
		}
	}
	testdatas["test_cases"] = tmp
	js, _ := json.Marshal(testdatas)
	if err := UpdateCols(pd, common.H{"testdatas": js}); err != nil {
		return err
	}
	if err := ioutil.WriteFile(path.Join(dir, "info"), js, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func GetProblemZSetKey(isOpen bool) string {
	if isOpen {
		return OPEN_PROBLEM_ZSET_KEY
	}
	return CLOSE_PROBLEM_ZSET_KEY
}
func ProblemCount(isOpen bool) int64 {
	return rdb.ZCount(ctx, GetProblemZSetKey(isOpen), "-inf", "+inf").Val()
}

func GetProblemsBaseInfo(l int64, r int64, isOpen bool) []common.H {
	ps := rdb.ZRange(ctx, GetProblemZSetKey(isOpen), l-1, r-1).Val()
	mps := make([]common.H, len(ps))
	for idx, item := range ps {
		pd := &ProblemDao{Index: item}
		cols := Cols(pd, "author", "source", "title", "accepted_count", "all_count", "tags")
		mps[idx] = common.H{
			"index":  item,
			"author": cols[0].ToString(),
			"source": cols[1].ToString(),
			"title":  cols[2].ToString(),
			"ac":     cols[3].ToUint(),
			"all":    cols[4].ToUint(),
			"tags":   cols[5].ToString(),
		}
	}
	return mps
}

func GetOneProblemInfoAsWants(index string, wants []string) common.H {
	pd := &ProblemDao{Index: index}
	if !Exists(pd) {
		return nil
	}
	cols := Cols(pd, wants...)
	mp := make(common.H, len(cols))
	for idx, item := range cols {
		mp[wants[idx]] = item.data
	}
	return mp
}
func GetOneProblemInfoAsWantsByID(id int64, wants []string) common.H {
	pd := &ProblemDao{ID: id}
	if !Exists(pd) {
		return nil
	}
	cols := Cols(pd, wants...)
	mp := make(common.H, len(cols))
	for idx, item := range cols {
		mp[wants[idx]] = item.data
	}
	return mp
}
func GetProblemsAsWants(l, r int64, wants []string, isOpen bool) []common.H {
	ps := rdb.ZRange(ctx, GetProblemZSetKey(isOpen), l-1, r-1).Val()
	mps := make([]common.H, len(ps))
	for idx, item := range ps {
		mps[idx] = GetOneProblemInfoAsWants(item, wants)
	}
	return mps
}

func GetProblemIDsByTitle(isOpen bool, title string) []int64 {
	ps := rdb.ZRange(ctx, GetProblemZSetKey(isOpen), 0, -1).Val()
	ids := make([]int64, 0)
	for _, index := range ps {
		pd := &ProblemDao{Index: index}
		t := OneCol(pd, "title").ToString()
		if strings.Contains(t, title) {
			ids = append(ids, pd.GetID())
		}
	}
	return ids
}
