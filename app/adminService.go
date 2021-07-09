package app

import (
	"HappyOnlineJudge/common"
	"HappyOnlineJudge/dao"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"mime/multipart"
	"os"
	"path"
)

func forAdminPage(c *gin.Context) {
	c.Set("result", "ok")
}

func newProblem(c *gin.Context) {
	pJson := c.DefaultPostForm("problem", "")
	if pJson == "" {
		setError(c, 403, "参数错误")
		return
	}
	pd := &dao.ProblemDao{Problem: &dao.Problem{}}
	if err := json.Unmarshal([]byte(pJson), pd.Problem); err != nil {
		setError(c, 403, "参数错误")
		return
	}
	pd.SetIndex()
	dir := path.Join(TEST_CASE_DIR, pd.GetIndex())
	os.MkdirAll(dir, os.ModePerm)
	if pd.Problem.IsSpj {
		if spjFile, err := c.FormFile("spj_file"); err == nil {
			if err := c.SaveUploadedFile(spjFile, path.Join(dir, spjFile.Filename)); err != nil {
				setError(c, 403, err.Error())
				return
			}
		} else {
			setError(c, 403, "未上传特判文件")
			return
		}
	}
	if err := pd.Created(); err != nil {
		os.RemoveAll(dir)
		setError(c, 500, err.Error())
		return
	}
	c.Set("index", pd.GetIndex())
}

func getProblems(c *gin.Context) {
	l := common.StrToInt64(c.DefaultQuery("l", "1"))
	r := common.StrToInt64(c.DefaultQuery("r", "50"))
	isOpen := common.StrToBool(c.DefaultQuery("is_open", "true"))
	l, r = -r+1, -l+1
	c.Set("data", dao.GetProblemsBaseInfo(l, r, isOpen))
	c.Set("total", dao.ProblemCount(isOpen))
}

func getProblemAllInfo(c *gin.Context) {
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
	p := pd.Problem
	c.Set("problem", common.H{
		"author":         p.Author,
		"source":         p.Source,
		"title":          p.Title,
		"background":     p.Background,
		"statement":      p.Statement,
		"input":          p.Input,
		"output":         p.Output,
		"examples_in":    p.ExamplesIn,
		"examples_out":   p.ExamplesOut,
		"hint":           p.Hint,
		"time_limit":     p.TimeLimit,
		"memory_limit":   p.MemoryLimit,
		"is_open":        p.IsOpen,
		"is_spj":         p.IsSpj,
		"spj_type":       p.SpjType,
		"accepted_count": p.AcceptedCount,
		"AllCount":       p.AllCount,
	})
}
func getOneProblemInfoByJson(c *gin.Context) {
	js := c.DefaultPostForm("json", "")
	index := c.DefaultPostForm("index", "")
	if js == "" || index == "" {
		setError(c, 403, "参数错误")
		return
	}
	wants := make([]string, 0)
	if err := json.Unmarshal([]byte(js), &wants); err != nil {
		setError(c, 403, err.Error())
		return
	}
	mp := dao.GetOneProblemInfoAsWants(index, wants)
	if mp == nil {
		setError(c, 403, "参数错误")
		return
	}
	c.Set("data", mp)
}

func updateProblemByJson(c *gin.Context) {
	js := c.DefaultPostForm("json", "")
	index := c.DefaultPostForm("index", "")
	if js == "" || index == "" {
		setError(c, 403, "参数错误")
		return
	}
	mp := make(common.H)
	if err := json.Unmarshal([]byte(js), &mp); err != nil {
		setError(c, 403, err.Error())
		return
	}
	pd := &dao.ProblemDao{Index: index}
	//if !dao.Exists(pd) {
	//	setError(c, 403, "不存在该题目")
	//	return
	//}
	if err := dao.UpdateCols(pd, mp); err != nil {
		setError(c, 500, err.Error())
		return
	}
	c.Set("result", "ok")
}

func addSpjFile(c *gin.Context) {
	spjType := c.DefaultPostForm("spj_type", "")
	index := c.DefaultPostForm("index", "")
	if spjType == "" || index == "" {
		setError(c, 403, "参数错误")
		return
	}
	pd := &dao.ProblemDao{Index: index}
	//if !dao.Exists(pd) {
	//	setError(c, 403, "不存在该题目")
	//	return
	//}
	dir := path.Join(TEST_CASE_DIR, pd.GetIndex())
	spjFile, err := c.FormFile("spj_file")
	if err == nil {
		if err := c.SaveUploadedFile(spjFile, path.Join(dir, spjFile.Filename)); err != nil {
			setError(c, 403, err.Error())
			return
		}
	} else {
		setError(c, 403, "未上传特判文件")
		return
	}
	if err := dao.UpdateCols(pd, common.H{"is_spj": true, "spj_type": spjType}); err != nil {
		os.Remove(path.Join(dir, spjFile.Filename))
		setError(c, 403, err.Error())
		return
	}
	c.Set("result", "ok")
}

func uploadTestdatas(c *gin.Context) {
	index := c.DefaultPostForm("index", "")
	pd := &dao.ProblemDao{Index: index}
	var zipFile *multipart.FileHeader
	var err error
	if zipFile, err = c.FormFile("zip"); err != nil {
		c.String(403, err.Error())
		return
	}
	dir := path.Join(TEST_CASE_DIR, index)
	common.RemoveTestDatas(dir)
	zipPath := path.Join(dir, zipFile.Filename)
	c.SaveUploadedFile(zipFile, zipPath)
	fmt.Println(dir)
	if err := pd.HandleZipData(dir, zipPath); err != nil {
		setError(c, 403, err.Error())
		return
	}
	fmt.Println(dir)
	os.Remove(zipPath)
	c.Set("result", "ok")
}
func removeTestdatas(c *gin.Context) {
	index := c.DefaultQuery("index", "")
	pd := &dao.ProblemDao{Index: index}
	dir := path.Join(TEST_CASE_DIR, index)
	common.RemoveTestDatas(dir)
	os.Remove(path.Join(dir, "info"))
	if err := dao.UpdateCols(pd, common.H{"testdatas": "null"}); err != nil {
		setError(c, 403, err.Error())
		return
	}
	c.Set("result", "ok")
}

func getTestdatas(c *gin.Context) {
	index := c.DefaultQuery("index", "")
	pd := &dao.ProblemDao{Index: index}
	mp := dao.OneCol(pd, "testdatas").ToStringMapAny()
	data := make([]common.H, 0)
	if len(mp) > 0 {
		mp = mp["test_cases"].(common.H)
		for k, v := range mp {
			tmp := v.(common.H)
			tmp["case_id"] = k
			data = append(data, tmp)
		}
	}
	c.Set("data", data)
}

func showOneTestdata(c *gin.Context) {
	index := c.DefaultQuery("index", "")
	caseID := c.DefaultQuery("case_id", "")
	pd := &dao.ProblemDao{Index: index}
	dir := path.Join(TEST_CASE_DIR, index)
	mp := dao.OneCol(pd, "testdatas").ToStringMapAny()
	mp = mp["test_cases"].(common.H)
	item := mp[caseID].(common.H)
	input, err1 := common.GetContent(path.Join(dir, item["input_name"].(string)))
	output, err2 := common.GetContent(path.Join(dir, item["output_name"].(string)))
	if err1 != nil || err2 != nil {
		setError(c, 403, "获取失败")
		return
	}
	c.Set("input", input)
	c.Set("output", output)
}

func copyProblem(c *gin.Context) {
	index := c.DefaultQuery("index", "")
	pd := &dao.ProblemDao{Index: index}
	dir := path.Join(TEST_CASE_DIR, index)
	dao.GetSelfAll(pd)
	pd.ID, pd.Problem.ID, pd.Problem.AllCount, pd.Problem.AcceptedCount, pd.Problem.IsOpen = 0, 0, 0, 0, !pd.Problem.IsOpen
	pd.SetIndex()
	if err := common.CopyDir(dir, path.Join(TEST_CASE_DIR, pd.GetIndex())); err != nil {
		setError(c, 403, err.Error())
		return
	}
	if err := pd.Created(); err != nil {
		setError(c, 403, err.Error())
		return
	}
	c.Set("result", "ok")
}
func delProblem(c *gin.Context) {
	index := c.DefaultQuery("index", "")
	pd := &dao.ProblemDao{Index: index}
	dir := path.Join(TEST_CASE_DIR, index)
	os.RemoveAll(dir)
	if err := pd.Delete(); err != nil {
		setError(c, 403, err.Error())
		return
	}
	c.Set("result", "ok")
}
func updateProblemTags(c *gin.Context) {
	tags := c.DefaultPostForm("tags", "")
	pd := &dao.ProblemDao{Index: c.DefaultPostForm("index", "")}
	if err := dao.UpdateCols(pd, common.H{"tags": tags}); err != nil {
		setError(c, 403, "操作失败")
		return
	}
	c.Set("result", "ok")
}

func downloadDatas(c *gin.Context) {
	index := c.DefaultPostForm("index", "")
	if index == "" {
		setError(c, 403, "参数错误")
		return
	}
	dir := path.Join(TEST_CASE_DIR, index)
	dest := path.Join(dir, "data.zip")
	//time.AfterFunc(time.Minute, func() { os.Remove(dest) })
	defer os.Remove(dest)
	if err := common.CompressToZip(common.GetFilesOfSomeExts(dir, []string{".in", ".out", ".ans"}), dest); err != nil {
		setError(c, 500, err.Error())
		return
	}
	c.Writer.Header().Add("Content-Disposition", "attachment;filename=data.zip")
	c.Writer.Header().Set("Content-Type", "application/zip")
	c.Set("noPack", true)
	c.File(dest)
}

func newContest(c *gin.Context) {
	cd := &dao.ContestDao{
		Contest: &dao.Contest{},
	}
	if err := json.Unmarshal([]byte(c.PostForm("contest")), &cd.Contest); err != nil {
		c.String(403, err.Error())
		return
	}
	if err := cd.Create(); err != nil {
		c.String(403, err.Error())
		return
	}
	c.Set("id", cd.GetID())
}

func searchContests(c *gin.Context) {
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

func getContestInfo(c *gin.Context) {
	id := common.StrToInt64(c.Query("id"))
	cd := &dao.ContestDao{ID: id}
	if x := dao.GetSelfAll(cd); x != nil {
		setError(c, 403, "Not find")
		return
	}
	c.Set("contest", cd.Contest)
}
func updateContest(c *gin.Context) {
	update := c.PostForm("update")
	id := c.PostForm("id")
	mp := make(common.H)
	if err := json.Unmarshal([]byte(update), &mp); err != nil || id == "" {
		setError(c, 403, "参数错误")
		return
	}
	cd := &dao.ContestDao{ID: common.StrToInt64(id)}
	if err := cd.Update(mp); err != nil {
		setError(c, 403, err.Error())
		return
	}
	c.Set("result", "ok")
}

func getCproblems(c *gin.Context) {
	id := common.StrToInt64(c.DefaultQuery("id", "0"))
	if id == 0 {
		setError(c, 403, "参数错误")
		return
	}
	cps := dao.GetCproblems(id)
	ps, titles := make([]string, len(cps)), make([]string, len(cps))
	for i, item := range cps {
		pd := &dao.ProblemDao{ID: item.ProblemID}
		ps[i] = pd.GetIndex()
		titles[i] = dao.OneCol(pd, "title").ToString()

	}
	c.Set("num", len(cps))
	c.Set("problems", ps)
	c.Set("titles", titles)
}

func searchProbelmTitle(c *gin.Context) {
	index := c.DefaultQuery("index", "")
	pd := &dao.ProblemDao{Index: index}
	if !dao.Exists(pd) {
		c.Set("existteam", false)
	} else {
		c.Set("exist", true)
		c.Set("title", dao.OneCol(pd, "title").ToString())
	}
}

func addCproblems(c *gin.Context) {
	id := common.StrToInt64(c.DefaultPostForm("id", "0"))
	problems := make([]string, 0)
	labels := make([]string, 0)
	if err := json.Unmarshal([]byte(c.PostForm("problems")), &problems); err != nil {
		setError(c, 403, err.Error())
		return
	}
	if err := json.Unmarshal([]byte(c.PostForm("labels")), &labels); err != nil {
		setError(c, 403, err.Error())
		return
	}
	if err := dao.DeleteCproblems(id); err != nil {
		setError(c, 403, err.Error())
		return
	}
	cps := make([]dao.Cproblem, len(problems))
	for idx, item := range problems {
		pd := &dao.ProblemDao{Index: item}
		cols := dao.Cols(pd, "title", "tags")
		cps[idx] = dao.Cproblem{
			ProblemID: pd.GetID(),
			ContestID: id,
			Label:     labels[idx],
			Title:     cols[0].ToString(),
			Tags:      cols[1].ToString(),
		}
	}
	dao.AddCproblems(cps)
	c.Set("result", "ok")
}

func deleteContest(c *gin.Context) {
	id := common.StrToInt64(c.DefaultQuery("id", "0"))
	if id == 0 {
		setError(c, 403, "参数错误")
		return
	}
	cd := &dao.ContestDao{ID: id}
	if err := cd.Delete(); err != nil {
		setError(c, 403, err.Error())
		return
	}
	c.Set("result", "ok")
}

func forCproblemLabels(c *gin.Context) {
	cd := &dao.ContestDao{ID: common.StrToInt64(c.DefaultQuery("id", "0"))}
	if !dao.Exists(cd) {
		setError(c, 403, "Not Found")
		return
	}
	c.Set("labels", dao.GetLabels(cd.ID))
}
func forCproblemTitles(c *gin.Context) {
	cd := &dao.ContestDao{ID: common.StrToInt64(c.DefaultQuery("id", "0"))}
	if !dao.Exists(cd) {
		setError(c, 403, "Not Found")
		return
	}
	labels, titles := dao.GetCproblemTitles(cd.ID)
	c.Set("labels", labels)
	c.Set("titles", titles)
}
func forRankList(c *gin.Context) {
	cd := &dao.ContestDao{ID: common.StrToInt64(c.DefaultQuery("id", "0"))}
	if !dao.Exists(cd) {
		setError(c, 403, "Not Found")
		return
	}
	c.Set("data", dao.GetAllTeamRankData(cd.ID))
}

func forCsubmissions(c *gin.Context) {
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

func rejudge(c *gin.Context) {
	cid := common.StrToInt64(c.DefaultPostForm("id", "0"))
	js := c.PostForm("labels")
	labels := make([]string, 0)
	err := json.Unmarshal([]byte(js), &labels)
	if cid == 0 || err != nil {
		setError(c, 403, "参数错误")
		return
	}
	if len(labels) > 0 {
		go dao.ReJudge(cid, labels)
	}
	c.Set("result", "ok")
}

func getUsers(c *gin.Context) {
	username := c.DefaultQuery("username", "")
	wants := []string{"id", "username", "created_at", "school", "email", "is_admin"}

	if username != "" {
		ud := &dao.UserDao{Username: username}
		if !dao.Exists(ud) {
			c.Set("total", 0)
			c.Set("data", []common.H{})
			return
		}
		cols := dao.Cols(ud, wants...)
		status := "普通用户"
		if cols[0].ToInt64() == 1 {
			status = "超级管理员"
		} else if cols[5].ToBool() {
			status = "管理员"
		}
		c.Set("data", []common.H{common.H{
			"id":         cols[0].ToInt64(),
			"username":   cols[1].ToString(),
			"created_at": cols[2].ToTime(),
			"school":     cols[3].ToString(),
			"email":      cols[4].ToString(),
			"status":     status,
		}})
		c.Set("total", 1)
		return
	}
	l := common.StrToInt64(c.DefaultQuery("l", "1"))
	r := common.StrToInt64(c.DefaultQuery("r", "50"))
	users := dao.GetUsersByCreatedTime(l, r, wants)
	data := make([]common.H, len(users))
	for i, item := range users {
		status := "普通用户"
		if item.ID == 1 {
			status = "超级管理员"
		} else if item.IsAdmin {
			status = "管理员"
		}
		data[i] = common.H{
			"id":         item.ID,
			"username":   item.Username,
			"created_at": item.CreatedAt.Format(common.TIME_FORMAT),
			"school":     item.School,
			"email":      item.Email,
			"status":     status,
		}
	}
	c.Set("data", data)
	c.Set("total", dao.CountUsers())
}

func getPrivileges(c *gin.Context) {
	ud := &dao.UserDao{ID: common.StrToInt64(c.DefaultQuery("id", "0"))}
	if !dao.Exists(ud) {
		setError(c, 403, "参数错误")
		return
	}
	pri := dao.OneCol(ud, "privilege").ToUint64()
	hadPris := make([]uint64, 0)

	d := uint64(len(PrivilegeDesc))
	priList := make([]uint64, d)
	for i := uint64(0); i < d; i++ {
		if (pri & (1 << i)) > 0 {
			hadPris = append(hadPris, 1<<i)
		}
		priList[i] = 1 << i
	}
	c.Set("pri_value_list", priList)
	c.Set("pri_desc_list", PrivilegeDesc)
	c.Set("had_pri_list", hadPris)
}
func updatePrivileges(c *gin.Context) {
	ud := &dao.UserDao{ID: common.StrToInt64(c.DefaultPostForm("id", "0"))}
	if !dao.Exists(ud) {
		setError(c, 403, "参数错误")
		return
	}
	if ud.GetID() == 1 {
		setError(c, 403, "超级管理员拥有完全的权限,无需设置")
		return
	}
	privileges := make([]uint64, 0)
	if err := json.Unmarshal([]byte(c.PostForm("privileges")), &privileges); err != nil {
		setError(c, 403, "参数错误")
		return
	}
	pri := uint64(0)
	for _, item := range privileges {
		pri |= item
	}
	mp := common.H{"privilege": pri}
	status := "管理员"
	if pri == 0 {
		status = "普通用户"
	} else {
		mp["is_admin"] = true
	}
	if err := dao.UpdateCols(ud, mp); err != nil {
		setError(c, 403, "操作失败")
		return
	}
	c.Set("status", status)
}

func updateClarification(c *gin.Context) {
	cd := &dao.ContestDao{ID: common.StrToInt64(c.PostForm("id"))}
	clarifications := dao.OneCol(cd, "clarification").ToStringSlice()
	index := common.StrToInt(c.DefaultPostForm("index", "-1"))
	kind := c.PostForm("action")
	if index == -1 {
		clarifications = append([]string{c.PostForm("clarification")}, clarifications...)
	} else if kind == "del" {
		clarifications = append(clarifications[:index], clarifications[index+1:]...)
	} else {
		clarifications[index] = c.PostForm("clarification")
	}
	dao.UpdateCols(cd, common.H{"clarification": clarifications})
	c.Set("result", "ok")
}

func newClass(c *gin.Context) {
	name := c.DefaultQuery("name", "")
	if name == "" {
		setError(c, 403, "参数错误")
		return
	}
	password := c.DefaultQuery("password", "")
	class := dao.NewClass(name, password)
	if class == nil {
		setError(c, 403, "创建失败")
		return
	}
	c.Set("class", class)
}

func getClasses(c *gin.Context) {
	c.Set("data", dao.GetClasses())
}

func updateClassInfo(c *gin.Context) {
	id := common.StrToInt64(c.Query("id"))
	name := c.Query("name")
	password := c.Query("password")
	dao.UpdateClassInfo(id, name, password)
	c.Set("result", "ok")
}
