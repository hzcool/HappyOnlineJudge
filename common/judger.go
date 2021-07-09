package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

type ErrType int

const (
	OK                         ErrType = 0 //没有错误
	INVALID_HEADER             ErrType = 1 // 请求头错误
	INVALID_JSON_REQUEST       ErrType = 2 // 请求的json数据不对
	STORE_SRC_FAILED           ErrType = 3 // 系统保存上传的测试代码出错
	COMPILE_ERROR              ErrType = 4 // 编译错误
	INVALID_CONFIG             ErrType = 5 // 时间、内存的设置不在合理范围内
	INVAILD_TESTCASE_INFO_FILE ErrType = 6 //测试数据的info描述文件不对
	SYSTEM_EXCEPTION           ErrType = 7 //系统异常
)

const (
	ACCEPTED              int = 0
	WRONG_ANSWER          int = 1
	TIME_LIMIT_EXCEEDED   int = 2
	MEMORY_LIMIT_EXCEEDED int = 3
	OUTPUT_LIMIT_EXCEEDED int = 4
	RUNTIME_ERROR         int = 5
	SYSTEM_ERROR          int = 6
)

var resultMap = map[int]string{
	0: "Accepted",
	1: "WrongAnswer",
	2: "TimeLimitExceeded",
	3: "MemoryLimitExceeded",
	4: "OutputLimitExceeded",
	5: "RuntimeError",
	6: "SystemError",
}

var (
	TEST_CASE_DIR = ""
	MAX_TASKS     = 0
	SERVICE_URL   = ""
	ACCESS_TOKEN  = ""
	client        = &http.Client{}
	CH            chan int //控制最大数量的同时处理
)

func initJudger(judge_cfg map[string]interface{}) {
	TEST_CASE_DIR, _ = judge_cfg["TEST_CASE_DIR"].(string)
	SERVICE_URL = judge_cfg["SERVICE_URL"].(string)
	ACCESS_TOKEN = judge_cfg["ACCESS_TOKEN"].(string)
	tmp, _ := judge_cfg["MAX_TASKS"].(string)
	MAX_TASKS, _ = strconv.Atoi(tmp)

	CH = make(chan int, MAX_TASKS)
	for i := 1; i <= MAX_TASKS; i++ {
		CH <- i
	}
	Ping()
}

func Ping() {
	url := SERVICE_URL + "/ping"
	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Access-Token", ACCESS_TOKEN)

	res, _ := client.Do(req)
	defer res.Body.Close()

	var mp H
	json.NewDecoder(res.Body).Decode(&mp)
	fmt.Println("ping result : ")
	fmt.Println(mp["info"].(string))
}

func ToJudge(task H) H {
	post, _ := json.Marshal(task)
	url := SERVICE_URL + "/judge"
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(post))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Access-Token", ACCESS_TOKEN)
	res, _ := client.Do(req)
	defer res.Body.Close()

	//结果处理
	resBody, _ := ioutil.ReadAll(res.Body)
	var result H
	_ = json.Unmarshal(resBody, &result)
	var update = H{
		"status":       "",
		"time":         uint(0),
		"memory":       uint(0),
		"score":        uint(0),
		"compile_info": "",
	}
	err := ErrType(result["err"].(float64))

	if err == OK {
		update = H{
			"status":       resultMap[int(result["result"].(float64))],
			"time":         uint(result["cpu_time"].(float64)),
			"memory":       uint(result["memory"].(float64)) / 1024 / 1024,
			"score":        uint(result["pass"].(float64) / result["total"].(float64) * 100),
			"compile_info": result["compile_info"].(string),
		}
	} else if err == COMPILE_ERROR {
		update["status"] = "CompileError"
		update["compile_info"] = result["info"].(string)
	} else {
		update["status"] = "SystemError"
	}
	return update
}

//func Judge(s *model.Submission, problem *model.Problem) {
//	chID := <-CH
//	defer func() {
//		CH <- chID
//	}()
//	//测评
//
//	s.UpdateMap(H{"status": "Running"})
//	update := ToJudge(H{
//		"lang":         s.Lang,
//		"max_cpu_time": problem.TimeLimit,
//		"max_memory":   problem.MemoryLimit * 1024 * 1024,
//		"test_case":    problem.Index,
//		"src":          s.Code,
//	})
//	s.UpdateMap(update)
//	if s.Status == "Accepted" {
//		problem.UpdateMap(H{"ac": problem.AC + 1, "all": problem.All + 1})
//	} else {
//		problem.UpdateMap(H{"all": problem.All + 1})
//	}
//}
