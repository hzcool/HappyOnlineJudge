package app

import (
	"HappyOnlineJudge/common"
	"HappyOnlineJudge/dao"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"math/rand"
	"os"
	"path"
	"sort"
	"strconv"
	"time"
)

func hello(c *gin.Context) {
	c.Set("msg", "hello")
}

func ping(c *gin.Context) {
	c.Set("ping", "pong")
}

func autologin(c *gin.Context) {
	id := getUserID(c)
	if id != 0 {
		ud := &dao.UserDao{ID: id}
		cols := dao.Cols(ud, "avatar", "is_admin", "is_super_admin", "privilege")
		c.Set("username", getUserName(c))
		c.Set("avatar", cols[0].ToString())
		c.Set("is_admin", cols[1].ToBool())
		c.Set("is_super_admin", cols[2].ToBool())
		c.Set("privilege", cols[3].ToUint64())
		return
	}
	setError(c, 401, "未登录")
}

//登陆请求
func login(c *gin.Context) {
	if isLogin(c) {
		deleteSession(c)
	}
	form := new(loginValidtor)
	if err := c.ShouldBind(form); err != nil {
		setError(c, 403, err.Error())
		return
	}
	form.Password = common.PassWordHandle(form.Password)
	ud := &dao.UserDao{Username: form.Username}
	id := ud.GetID()
	if id <= 0 {
		setError(c, 403, "用户名不存在")
		return
	}
	if pwd := dao.OneCol(ud, "password").ToString(); pwd != form.Password {
		setError(c, 403, "密码错误")
		return
	}
	if !dao.IsInRedis(ud) {
		dao.GetSelfAll(ud)
		dao.PutToRedis(ud)
	}
	setSession(c, ud.Username, ud.GetID())
	autologin(c)
}

func logout(c *gin.Context) {
	deleteSession(c)
	c.Set("msg", "退出成功")
}

//注册请求
func register(c *gin.Context) {
	if isLogin(c) {
		setError(c, 403, "请先退出当前账户")
		return
	}
	form := new(registerValidtor)
	if err := c.ShouldBind(form); err != nil {
		setError(c, 403, err.Error())
		return
	}
	form.Password = string(common.RSADecrypt(form.Password))
	if ok, errInfo := form.isOk(); !ok {
		setError(c, 403, errInfo)
		return
	}
	if dao.Count(new(dao.UsersData), []string{"username"}, []interface{}{form.Username}) > 0 {
		setError(c, 403, "用户名已存在")
		return
	}
	if dao.Count(new(dao.UsersData), []string{"email"}, []interface{}{form.Email}) > 0 {
		setError(c, 403, "邮箱已被注册")
		return
	}
	form.Password = common.GetMD5Password(form.Password)
	ud := &dao.UserDao{
		User: &dao.User{
			Username: form.Username,
			Password: form.Password,
			School:   form.School,
			Email:    form.Email,
			Avatar:   common.Avatars[rand.Intn(len(common.Avatars))],
		},
	}
	if err := ud.Create(); err != nil {
		setError(c, 500, err.Error())
		return
	}
	setSession(c, form.Username, ud.GetID())
	autologin(c)
}

//更新用户的信息
func update(c *gin.Context) {
	form := new(updateValidtor)
	if err := c.ShouldBind(form); err != nil {
		setError(c, 403, err.Error())
		return
	}
	if form.NewPassword != "" {
		form.NewPassword = string(common.RSADecrypt(form.NewPassword))
	}
	if form.OldPassword != "" {
		form.OldPassword = string(common.RSADecrypt(form.OldPassword))
	}
	if ok, errInfo := form.isOk(); !ok {
		setError(c, 403, errInfo)
		return
	}
	if form.NewPassword != "" {
		form.NewPassword = common.GetMD5Password(form.NewPassword)
	}
	if form.OldPassword != "" {
		form.OldPassword = common.GetMD5Password(form.OldPassword)
	}
	name := getUserName(c)
	mp := make(map[string]interface{}) //要修改的内容
	ud := &dao.UserDao{ID: getUserID(c)}
	if form.Username != "" && form.Username != name {
		if dao.Count(new(dao.UsersData), []string{"username"}, []interface{}{form.Username}) > 0 {
			setError(c, 403, "用户名已存在")
			return
		}
		mp["username"] = form.Username
	}
	cols := dao.Cols(ud, "password", "email")
	if form.NewPassword != "" {
		if form.OldPassword != cols[0].ToString() {
			setError(c, 403, "密码错误")
			return
		}
		mp["password"] = form.NewPassword
	}

	if form.Email != "" && form.Email != cols[1].ToString() {
		if dao.Count(new(dao.UsersData), []string{"email"}, []interface{}{form.Email}) > 0 {
			setError(c, 403, "邮箱已被注册")
			return
		}
		mp["email"] = form.Email
	}

	if form.School != "" {
		mp["school"] = form.School
	}
	if form.Desc != "" {
		mp["description"] = form.Desc
	}
	if len(mp) > 0 {
		if err := ud.Update(mp); err != nil {
			setError(c, 500, err.Error())
			return
		}
	}
	if _, ok := mp["username"]; ok {
		setSession(c, mp["username"].(string), ud.GetID())
	}
	c.Set("msg", "修改成功")
}

func addImg(c *gin.Context) {
	file, _ := c.FormFile("file")
	fileName := file.Filename
	dir := path.Join("./statics", strconv.FormatInt(getUserID(c), 10))
	os.MkdirAll(dir, os.ModePerm)
	url := common.WebHttp + "/statics/" + strconv.FormatInt(getUserID(c), 10) + "/" + fileName
	if err := c.SaveUploadedFile(file, path.Join(dir, fileName)); err != nil {
		setError(c, 403, "上传图片失败")
		return
	}
	c.Set("url", url)
}

func showAvatars(c *gin.Context) {
	c.Set("avatars", common.Avatars)
}
func changeAvatar(c *gin.Context) {
	if avatar := c.DefaultQuery("avatar", ""); avatar == "" {
		setError(c, 403, "参数错误")
		return
	} else {
		ud := &dao.UserDao{ID: getUserID(c)}
		dao.UpdateCols(ud, common.H{"avatar": avatar})
	}
	c.Set("result", "ok")
}
func getUserInfo(c *gin.Context) {
	pd := &dao.UserDao{Username: c.Query("username")}
	if !dao.Exists(pd) {
		setError(c, 403, "没有该用户")
		return
	}
	cols := dao.Cols(pd, "username", "created_at", "school", "email", "description", "avatar", "is_admin", "passed_problems", "failed_problems")
	passed := cols[7].ToStringMapInt64()
	failed := cols[8].ToStringMapInt64()
	passedArr := make([]string, len(passed))
	failedArr := make([]string, len(failed))
	i, j := 0, 0
	for k, _ := range passed {
		passedArr[i] = k
		i++
	}
	for k, _ := range failed {
		failedArr[j] = k
		j++
	}
	sort.Strings(passedArr)
	sort.Strings(failedArr)
	info := common.H{
		"username":    cols[0].ToString(),
		"created_at":  cols[1].ToTime().Format(common.TIME_FORMAT),
		"school":      cols[2].ToString(),
		"email":       cols[3].ToString(),
		"description": cols[4].ToString(),
		"avatar":      cols[5].ToString(),
		"is_admin":    cols[6].ToBool(),
		"passed":      passedArr,
		"failed":      failedArr,
	}
	c.Set("info", info)
	c.Set("status", dao.GetUserSubCondition(pd.GetID()))
}

func searchUsers(c *gin.Context) {
	data := make([]common.H, 0)
	wants := []string{"username", "school", "passed_count", "passed_sub_count", "all_sub_count"}
	if name := c.DefaultQuery("username", ""); name != "" {
		ud := &dao.UserDao{Username: name}
		if dao.Exists(ud) {
			cols := dao.Cols(ud, wants...)
			ratio := float64(1)
			if cols[4].ToUint() != 0 {
				ratio = float64(cols[3].ToUint()) / float64(cols[4].ToUint())
			}
			data = append(data, common.H{
				"username":         name,
				"school":           cols[1].ToString(),
				"passed_count":     cols[2].ToUint(),
				"passed_sub_count": cols[3].ToUint(),
				"all_sub_count":    cols[4].ToUint(),
				"ratio":            ratio,
			})
		}
		c.Set("data", data)
		c.Set("total", 1)
		return
	}
	l := common.StrToInt64(c.DefaultQuery("l", "1"))
	r := common.StrToInt64(c.DefaultQuery("r", "50"))
	ids := dao.GetUsers(l, r)
	data = make([]common.H, len(ids))
	for i, id := range ids {
		ud := &dao.UserDao{ID: id}
		cols := dao.Cols(ud, wants...)
		ratio := float64(1)
		if cols[4].ToUint() != 0 {
			ratio = float64(cols[3].ToUint()) / float64(cols[4].ToUint())
		}
		data[i] = common.H{
			"username":         cols[0].ToString(),
			"school":           cols[1].ToString(),
			"passed_count":     cols[2].ToUint(),
			"passed_sub_count": cols[3].ToUint(),
			"all_sub_count":    cols[4].ToUint(),
			"ratio":            ratio,
		}
	}
	c.Set("data", data)
	c.Set("total", dao.CountUsers())
}

func getProblemset(c *gin.Context) {
	l := common.StrToInt64(c.DefaultPostForm("l", "1"))
	r := common.StrToInt64(c.DefaultPostForm("r", "50"))
	ruleStr := c.DefaultPostForm("rule", "")
	data := make([]common.H, 0)
	total := int64(0)
	wants := []string{"index", "title", "accepted_count", "all_count"}
	if ruleStr == "" {
		data = dao.GetProblemsAsWants(l, r, wants, true)
		total = dao.ProblemCount(true)
	} else {
		rule := make(common.H)
		if err := json.Unmarshal([]byte(ruleStr), &rule); err != nil {
			setError(c, 403, err.Error())
			return
		}
		if item, ok := rule["index"]; ok {
			if x := dao.GetOneProblemInfoAsWants(item.(string), wants); x != nil {
				data = append(data, x)
				total = 1
			}
		} else if item, ok := rule["title"]; ok {
			ids := dao.GetProblemIDsByTitle(true, item.(string))
			total = int64(len(ids))
			for i := l - 1; i < r && i < total; i++ {
				data = append(data, dao.GetOneProblemInfoAsWantsByID(ids[i], wants))
			}
		}
	}
	if uid := getUserID(c); uid != 0 {
		ud := &dao.UserDao{ID: uid}
		passed := dao.OneCol(ud, "passed_problems").ToStringMapInt64()
		failed := dao.OneCol(ud, "failed_problems").ToStringMapInt64()
		for i := len(data) - 1; i >= 0; i-- {
			index := data[i]["index"].(string)
			if _, ok := passed[index]; ok {
				data[i]["solved"] = 1
			} else if _, ok2 := failed[index]; ok2 {
				data[i]["solved"] = 2
			}
		}
	}
	c.Set("total", total)
	c.Set("data", data)
}

func getOneProblem(c *gin.Context) {
	index := c.DefaultQuery("index", "")
	if index == "" {
		setError(c, 403, "参数错误")
		return
	}
	pd := &dao.ProblemDao{Index: index}
	if err := dao.GetSelfAll(pd); err != nil {
		setError(c, 403, err.Error())
		return
	}
	c.Set("info", pd)
}
func getProblemTitle(c *gin.Context) {
	index := c.Query("index")
	pd := &dao.ProblemDao{Index: index}
	if index == "" || index[0] != 'P' || !dao.Exists(pd) {
		setError(c, 403, "参数错误")
		return
	}
	c.Set("title", dao.OneCol(pd, "title").ToString())
}
func getProblemAsWants(c *gin.Context) {
	index := c.DefaultPostForm("index", "")
	js := c.DefaultPostForm("wants", "")
	if index == "" || js == "" {
		setError(c, 403, "参数错误")
		return
	}
	if index[0] == 'U' && !IsSuperAdmin(c) {
		setError(c, 403, "无权查看")
		return
	}
	wants := make([]string, 0)
	if err := json.Unmarshal([]byte(js), &wants); err != nil {
		setError(c, 403, err.Error())
		return
	}
	data := dao.GetOneProblemInfoAsWants(index, wants)
	c.Set("data", data)
}

func submitCode(c *gin.Context) {
	index := c.DefaultPostForm("index", "")
	code := c.DefaultPostForm("code", "")
	lang := c.DefaultPostForm("lang", "")
	if index == "" || code == "" || lang == "" {
		setError(c, 403, "参数错误")
		return
	}
	pd := &dao.ProblemDao{Index: index}
	sd := &dao.SubmissionDao{
		Submission: &dao.Submission{
			Code:      code,
			Lang:      lang,
			ProblemID: pd.GetID(),
			AuthorID:  getUserID(c),
			Length:    uint(len(code)),
		},
	}
	if err := dao.Create(sd); err != nil {
		setError(c, 403, "提交失败")
		return
	}
	go sd.Judge(pd)
	c.Set("id", sd.GetID())
}

func getOneSubmissionBase(c *gin.Context) {
	sid := common.StrToInt64(c.Query("id"))
	sd := &dao.SubmissionDao{ID: sid}
	cols := dao.Cols(sd, "created_at", "status")
	c.Set("data", common.H{
		"created_at": cols[0].ToString(),
		"status":     cols[1].ToString(),
		"id":         sid,
	})
}

func getSubmissionHistory(c *gin.Context) {
	id := getUserID(c)
	history := make([]common.H, 0)
	code := ""
	lang := "C++11"
	if id > 0 {
		index := c.Query("index")
		pd := &dao.ProblemDao{Index: index}
		ids := dao.GetSubZSet(id, pd.GetID())
		history = make([]common.H, len(ids))
		for i, _ := range ids {
			sd := &dao.SubmissionDao{ID: ids[i]}
			cols := dao.Cols(sd, "status", "created_at")
			history[i] = common.H{
				"id":         ids[i],
				"status":     cols[0].ToString(),
				"created_at": cols[1].ToString(),
			}
			if i == 0 {
				x := dao.Cols(sd, "code", "lang")
				code = x[0].ToString()
				lang = x[1].ToString()
			}
		}
	}
	c.Set("history", history)
	c.Set("code", code)
	c.Set("lang", lang)
}

func getSubmission(c *gin.Context) {
	sid, err := strconv.ParseInt(c.DefaultQuery("id", "0"), 10, 64)
	if sid == 0 || err != nil {
		setError(c, 403, "参数错误")
		return
	}
	sd := &dao.SubmissionDao{ID: sid}
	if err := dao.GetSelfAll(sd); err != nil {
		setError(c, 403, err.Error())
		return
	}

	if !dao.AuthScanProblemSubmission(&dao.UserDao{ID: getUserID(c)}, sd) {
		setError(c, 403, "无权查看")
		return
	}

	pd := &dao.ProblemDao{ID: sd.Submission.ProblemID}
	ud := &dao.UserDao{ID: sd.Submission.AuthorID}

	info := common.H{
		"index":        pd.GetIndex(),
		"author":       dao.OneCol(ud, "username").ToString(),
		"lang":         sd.Submission.Lang,
		"status":       sd.Submission.Status,
		"time":         sd.Submission.Time,
		"memory":       sd.Submission.Memory,
		"length":       sd.Submission.Length,
		"created_at":   sd.Submission.CreatedAt.Format(common.TIME_FORMAT),
		"compile_info": sd.Submission.CompileInfo,
		"code":         sd.Submission.Code,
	}
	c.Set("info", info)
}

func searchSubmissions(c *gin.Context) {
	l := common.StrToInt64(c.DefaultPostForm("l", "1"))
	r := common.StrToInt64(c.DefaultPostForm("r", "50"))
	mp := make(map[string]interface{})
	if err := json.Unmarshal([]byte(c.DefaultPostForm("rules", "{}")), &mp); err != nil {
		setError(c, 403, err.Error())
		return
	}
	rules := make([]string, 0)
	values := make([]interface{}, 0)

	if item, ok := mp["author"]; ok {
		ud := &dao.UserDao{Username: item.(string)}
		uid := ud.GetID()
		if uid == 0 {
			c.Set("data", make([]common.H, 0))
			return
		}
		rules = append(rules, "author_id")
		values = append(values, uid)
	}
	if item, ok := mp["index"]; ok {
		pd := &dao.ProblemDao{Index: item.(string)}
		pid := pd.GetID()
		if pid == 0 {
			c.Set("data", make([]common.H, 0))
			return
		}
		rules = append(rules, "problem_id")
		values = append(values, pid)
	}
	for k, v := range mp {
		if k != "author" && k != "index" {
			rules = append(rules, k)
			values = append(values, v)
		}
	}
	res := dao.SearchSubmissions(l, r, rules, values)
	data := make([]common.H, len(res))
	for idx, item := range res {
		pd := &dao.ProblemDao{ID: item.ProblemID}
		ud := &dao.UserDao{ID: item.AuthorID}
		data[idx] = common.H{
			"id":         item.ID,
			"created_at": item.CreatedAt.Format(common.TIME_FORMAT),
			"time":       item.Time,
			"memory":     item.Memory,
			"length":     item.Length,
			"status":     item.Status,
			"lang":       item.Lang,
			"score":      item.Score,
			"index":      pd.GetIndex(),
			"author":     ud.GetName(),
		}
	}
	c.Set("data", data)
}

func getContests(c *gin.Context) {
	l := common.StrToInt64(c.DefaultPostForm("l", "1"))
	r := common.StrToInt64(c.DefaultPostForm("r", "20"))
	mp := make(map[string]interface{})
	if err := json.Unmarshal([]byte(c.DefaultPostForm("rules", "{}")), &mp); err != nil {
		setError(c, 403, err.Error())
		return
	}
	rules := make([]string, 0)
	values := make([]interface{}, 0)
	for k, v := range mp {
		rules = append(rules, k)
		values = append(values, v)
	}
	total, res := dao.SearchContests(l, r, rules, values)
	data := make([]common.H, len(res))
	for idx, item := range res {
		typ := "public"
		if !item.IsPublic {
			typ = "private"
		}
		data[idx] = common.H{
			"id":     item.ID,
			"begin":  item.Begin.Format(common.TIME_FORMAT),
			"title":  item.Title,
			"length": float64(item.Length) / 3600,
			"status": item.Status,
			"format": item.Format,
			"type":   typ,
			"num":    dao.CountTeams(item.ID),
		}
	}
	c.Set("total", total)
	c.Set("data", data)
}

func enterContest(c *gin.Context) {
	id := common.StrToInt64(c.DefaultQuery("id", "0"))
	if id == 0 {
		setError(c, 403, "参数错误")
		return
	}
	cd := &dao.ContestDao{ID: id}
	if !dao.Exists(cd) {
		setError(c, 403, "不存在该比赛")
		return
	}
	if dao.OneCol(cd, "is_public").ToBool() {
		c.Set("ok", true)
		return
	}
	uid := getUserID(c)
	if id == 0 {
		setError(c, 403, "未登陆")
		return
	}
	if dao.ExistTeam(id, uid) || uid == 1 {
		c.Set("ok", true)
		return
	}

	c.Set("ok", false)
}

func checkContestPassword(c *gin.Context) {
	cid := common.StrToInt64(c.DefaultQuery("id", "0"))
	if cid == 0 {
		setError(c, 403, "参数错误")
		return
	}
	pwd := c.Query("password")
	cd := &dao.ContestDao{ID: cid}
	if dao.OneCol(cd, "password").ToString() != pwd {
		setError(c, 403, "密码错误")
		return
	}
	if err := dao.NewTeam(cid, getUserID(c)); err != nil {
		setError(c, 403, err.Error())
		return
	}
	c.Set("ok", true)
}

func getContestContent(c *gin.Context) {
	cid := common.StrToInt64(c.DefaultQuery("id", "0"))
	if cid == 0 {
		setError(c, 403, "参数错误")
		return
	}
	cd := &dao.ContestDao{ID: cid}
	if err := dao.GetSelfAll(cd); err != nil {
		setError(c, 403, err.Error())
		return
	}
	contest := cd.Contest
	typ := "public"
	if !contest.IsPublic {
		typ = "private"
	}
	c.Set("contest", common.H{
		"id":     contest.ID,
		"title":  contest.Title,
		"begin":  contest.Begin.Format(common.TIME_FORMAT),
		"end":    contest.End.Format(common.TIME_FORMAT),
		"length": float64(contest.Length) / 3600,
		"desc":   contest.Desc,
		"author": contest.Author,
		"type":   typ,
		"format": contest.Format,
		"num":    dao.CountTeams(cid),
		"now":    time.Now().Format(common.TIME_FORMAT),
	})
}

func loadCproblems(c *gin.Context) {
	cid := common.StrToInt64(c.DefaultQuery("id", "0"))
	cd := &dao.ContestDao{ID: cid}
	if !dao.Exists(cd) {
		setError(c, 403, "not find")
		return
	}
	begin := dao.OneCol(cd, "begin").ToTime()
	if !IsSuperAdmin(c) && time.Now().Before(begin) {
		setError(c, 403, "比赛还未开始")
		return
	}
	cps := dao.GetCproblems(cid)
	mp := dao.GetOneTeamStatus(cid, getUserID(c))
	data := make([]common.H, len(cps))
	for i, item := range cps {
		solved := 0
		if _, ok := mp[item.Label]; ok {
			tmp := mp[item.Label]
			if tmp["score"] == 100 {
				solved = 1
			} else if tmp["fail_times"] > 0 {
				solved = 2
			}
		}
		ratio := float64(1)
		if item.All > 0 {
			ratio = float64(item.AC) / float64(item.All)
		}
		data[i] = common.H{
			"title":  item.Title,
			"label":  item.Label,
			"solved": solved,
			"ratio":  ratio,
			"ac":     item.AC,
			"all":    item.All,
		}
	}
	c.Set("data", data)
}

func searchCsubmissions(c *gin.Context) {
	cid := common.StrToInt64(c.DefaultPostForm("id", "0"))
	l := common.StrToInt64(c.DefaultPostForm("l", "1"))
	r := common.StrToInt64(c.DefaultPostForm("r", "50"))
	mp := make(map[string]interface{})
	if err := json.Unmarshal([]byte(c.DefaultPostForm("rules", "{}")), &mp); err != nil {
		setError(c, 403, err.Error())
		return
	}
	rules := make([]string, len(mp)+1)
	values := make([]interface{}, len(mp)+1)
	rules[0] = "contest_id"
	values[0] = cid
	i := 1
	for k, v := range mp {
		if k == "author" {
			ud := &dao.UserDao{Username: v.(string)}
			rules[i] = "user_id"
			values[i] = ud.GetID()
		} else {
			rules[i] = k
			values[i] = v
		}
		i++
	}
	total, cs := dao.SearchCsubmissions(l, r, rules, values)
	data := make([]common.H, len(cs))
	for i, item := range cs {
		ud := &dao.UserDao{ID: item.UserID}
		data[i] = common.H{
			"run_id":     item.RunID,
			"author":     ud.GetName(),
			"status":     item.Status,
			"score":      item.Score,
			"label":      item.Label,
			"time":       item.Time,
			"memory":     item.Memory,
			"length":     item.Length,
			"lang":       item.Lang,
			"submitTime": item.CreatedAt.Format(common.TIME_FORMAT),
		}
		c.Set("total", total)
		c.Set("data", data)
	}
}

func getOneCproblem(c *gin.Context) {
	cid := common.StrToInt64(c.DefaultPostForm("id", "0"))
	label := c.PostForm("label")
	js := c.PostForm("wants")
	if cid == 0 || label == "" || js == "" {
		setError(c, 403, "参数错误")
		return
	}
	wants := make([]string, 0)
	if err := json.Unmarshal([]byte(js), &wants); err != nil {
		setError(c, 403, err.Error())
		return
	}
	cp := dao.GetCproblem(cid, label)
	if cp == nil {
		setError(c, 403, "Not found")
		return
	}
	pd := dao.ProblemDao{ID: cp.ProblemID}
	data := dao.GetOneProblemInfoAsWants(pd.GetIndex(), wants)
	data["tags"] = cp.Tags
	data["ac"] = cp.AC
	data["all"] = cp.All
	c.Set("data", data)
}

func getCSubmissionHistory(c *gin.Context) {
	id := getUserID(c)
	history := make([]common.H, 0)
	code := ""
	lang := "C++11"
	if id > 0 {
		cid := common.StrToInt64(c.DefaultQuery("id", "0"))
		label := c.Query("label")
		csubs := dao.GetUserOneProblemCsub(cid, id, label)
		history = make([]common.H, len(csubs))
		for i, item := range csubs {
			history[i] = common.H{
				"run_id":     item.RunID,
				"status":     item.Status,
				"created_at": item.CreatedAt.Format(common.TIME_FORMAT),
			}
			if i == 0 {
				code = item.Code
				lang = item.Lang
			}
		}
	}
	c.Set("history", history)
	c.Set("code", code)
	c.Set("lang", lang)
}

func submitContestCode(c *gin.Context) {
	cid := common.StrToInt64(c.PostForm("id"))
	cd := &dao.ContestDao{ID: cid}
	if !dao.Exists(cd) {
		setError(c, 403, "Not Found")
		return
	}
	label := c.PostForm("label")
	code := c.PostForm("code")
	lang := c.PostForm("lang")
	uid := getUserID(c)
	if uid != 1 {
		if dao.OneCol(cd, "begin").ToTime().After(time.Now()) {
			setError(c, 403, "比赛未开始,无法提交")
			return
		}
		if !dao.ExistTeam(cid, uid) {
			if err := dao.NewTeam(cid, uid); err != nil {
				setError(c, 401, err.Error())
				return
			}
		}
	}
	cs := dao.AddOneCsubmission(cid, uid, label, code, lang)
	dao.ContestJudger(cs)
	c.Set("run_id", cs.RunID)
}

func getOneCSubmissionBase(c *gin.Context) {
	run_id := common.StrToInt64(c.DefaultQuery("run_id", "0"))
	cid := common.StrToInt64(c.DefaultQuery("id", "0"))
	cs := dao.GetCsubmission(run_id, cid)
	c.Set("data", common.H{
		"run_id":     run_id,
		"created_at": cs.CreatedAt.Format(common.TIME_FORMAT),
		"status":     cs.Status,
	})
}

func showCsubmission(c *gin.Context) {
	run_id := common.StrToInt64(c.DefaultQuery("run_id", "0"))
	cid := common.StrToInt64(c.DefaultQuery("id", "0"))
	if run_id == 0 || cid == 0 {
		setError(c, 403, "参数错误")
		return
	}
	uid := getUserID(c)
	cs := dao.GetCsubmission(run_id, cid)
	if cs == nil {
		setError(c, 403, "not found")
		return
	}
	if uid != 1 && cs.UserID != uid {
		setError(c, 403, "无权查看")
		return
	}

	ud := &dao.UserDao{ID: cs.UserID}

	c.Set("info", common.H{
		"label":        cs.Label,
		"author":       dao.OneCol(ud, "username").ToString(),
		"lang":         cs.Lang,
		"status":       cs.Status,
		"time":         cs.Time,
		"memory":       cs.Memory / 1024 / 1024,
		"length":       cs.Length,
		"created_at":   cs.CreatedAt.Format(common.TIME_FORMAT),
		"compile_info": cs.CompileInfo,
		"code":         cs.Code,
	})
}

func getCproblemLabels(c *gin.Context) {
	cd := &dao.ContestDao{ID: common.StrToInt64(c.DefaultQuery("id", "0"))}
	if !dao.Exists(cd) {
		setError(c, 403, "Not Found")
		return
	}
	c.Set("labels", dao.GetLabels(cd.ID))
}
func getRankList(c *gin.Context) {
	cd := &dao.ContestDao{ID: common.StrToInt64(c.DefaultQuery("id", "0"))}
	if !dao.Exists(cd) {
		setError(c, 403, "Not Found")
		return
	}
	c.Set("data", dao.GetAllTeamRankData(cd.ID))
}

func getClarification(c *gin.Context) {
	cd := &dao.ContestDao{ID: common.StrToInt64(c.Query("id"))}
	clarifications := dao.OneCol(cd, "clarification").ToStringSlice()
	c.Set("data", clarifications)
}

func getCproblemIndexAndTitle(c *gin.Context) {
	cid, err := strconv.ParseInt(c.Query("id"), 10, 64)
	if err != nil {
		setError(c, 403, "参数错误")
		return
	}
	cd := &dao.ContestDao{ID: cid}
	uid := getUserID(c)
	if !dao.Exists(cd) || (uid != 1 && !dao.ExistTeam(cid, uid)) {
		setError(c, 403, "Not Found")
		return
	}
	cp := dao.GetCproblem(cid, c.Query("label"))
	pd := &dao.ProblemDao{ID: cp.ProblemID}
	if !dao.Exists(pd) {
		setError(c, 403, "Not Found")
		return
	}
	c.Set("index", dao.OneCol(pd, "index").ToString())
	c.Set("title", cp.Title)
}

func checkExistTeam(c *gin.Context) {
	cid, err := strconv.ParseInt(c.Query("id"), 10, 64)
	if err != nil {
		setError(c, 403, "参数错误")
		return
	}
	cd := &dao.ContestDao{ID: cid}
	uid := getUserID(c)
	if !dao.Exists(cd) || (uid != 1 && !dao.ExistTeam(cid, uid)) {
		setError(c, 403, "Not Found")
		return
	}
	c.Set("result", "ok")
}
