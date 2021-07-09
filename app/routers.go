package app

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"net/http"
)

//路由
func InitRouters() {

	init_config()
	gin.SetMode(gin.ReleaseMode)
	//r := gin.Default()
	r := gin.New()
	r.Use(gin.Recovery())

	r.LoadHTMLFiles("./dist/spa/index.html")
	r.StaticFS("/statics", http.Dir("./dist/spa/statics"))
	r.StaticFS("/js", http.Dir("./dist/spa/js"))
	r.StaticFS("/css", http.Dir("./dist/spa/css"))
	r.StaticFS("/fonts", http.Dir("./dist/spa/fonts"))
	r.StaticFS("/img", http.Dir("./dist/spa/img"))

	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	store := cookie.NewStore([]byte("secret")) //启用cookie和session
	store.Options(sessions.Options{
		MaxAge: int(SESSION_EXPIRE), //3天的过期时间
	})

	r.Use(jsonResponse)
	r.Use(sessions.Sessions("ginSession", store))

	initUserRouters(r)
	initAdminRouters(r)
	if err := r.Run(":9999"); err != nil {
		fmt.Println("路由初始化错误\n", err.Error())
	}
}

//用户基础的路由
func initUserRouters(r *gin.Engine) {
	g0 := r.Group("/api") // 无需任何条件的请求
	{
		//g0.GET("/", hello)
		g0.GET("ping", ping)
		g0.POST("login", login)
		g0.POST("register", register)
		g0.GET("autologin", autologin)
		g0.GET("getUserInfo", getUserInfo)
		g0.GET("searchUsers", searchUsers)
		g0.GET("showAvatars", showAvatars)

		//post
		g0.GET("getPostList", getPostList)
		g0.GET("getPost", getPost)
		g0.GET("getReplies", getReplies)
		g0.GET("getComments", getComments)

		//problem
		g0.POST("getProblemset", getProblemset)
		g0.GET("getOneProblem", getOneProblem)
		g0.GET("getProblemTitle", getProblemTitle)
		g0.POST("getProblemAsWants", getProblemAsWants)

		//submission
		g0.GET("getSubmissionHistory", getSubmissionHistory)
		g0.POST("searchSubmissions", searchSubmissions)
		g0.POST("getContests", getContests)

		//contest
		g0.GET("enterContest", enterContest)
		g0.GET("getContestContent", getContestContent)
		g0.GET("loadCproblems", loadCproblems)
		g0.POST("searchCsubmissions", searchCsubmissions)
		g0.POST("getOneCproblem", getOneCproblem)
		g0.GET("getCSubmissionHistory", getCSubmissionHistory)
		g0.GET("getCproblemLabels", getCproblemLabels)
		g0.GET("getRankList", getRankList)
		g0.GET("getClarification", getClarification)

	}

	g1 := r.Group("/api") //需要登陆才能进行的请求
	g1.Use(AuthLogin)     //authLogin 登陆验证中间件
	{
		g1.GET("logout", logout)
		g1.POST("update", update)
		g1.POST("addImg", addImg)
		g1.GET("changeAvatar", changeAvatar)

		//message
		g1.GET("findContactPerson", findContactPerson)
		g1.GET("getMessages", getMessages)
		g1.POST("sendOneMessage", sendOneMessage)
		g1.GET("getContacts", getContacts)
		g1.GET("delOneContact", delOneContact)

		//post
		g1.POST("newPost", newPost)
		g1.GET("onePost", onePost)
		g1.POST("updatePost", updatePost)
		g1.GET("countMyPost", countMyPost)
		g1.GET("getUserPostList", getUserPostList)
		g1.GET("deletePost", deletePost)
		g1.POST("sendOneReply", sendOneReply)
		g1.POST("sendOneComment", sendOneComment)
		g1.GET("addPostAttitude", addPostAttitude)

		//submission
		g1.POST("submitCode", submitCode)
		g1.GET("getOneSubmissionBase", getOneSubmissionBase)
		g1.GET("getSubmission", getSubmission)

		//contest
		g1.GET("checkContestPassword", checkContestPassword)
		g1.POST("submitContestCode", submitContestCode)
		g1.GET("getOneCSubmissionBase", getOneCSubmissionBase)
		g1.GET("showCsubmission", showCsubmission)
		g1.GET("getCproblemIndexAndTitle", getCproblemIndexAndTitle)
		g1.GET("getContestDiscussList", getContestDiscussList)
		g1.GET("checkExistTeam", checkExistTeam)

	}
}
func initAdminRouters(R *gin.Engine) {
	r := R.Group("/api")
	r.POST("newProblem", __CreateProblem, newProblem)
	r.GET("getProblems", __ScanPrivateProblem, getProblems)
	r.POST("updateProblemByJson", __UpdateProblem, updateProblemByJson)
	r.POST("addSpjFile", __UpdateProblem, addSpjFile)
	r.POST("uploadTestdatas", __ScanTestdata, uploadTestdatas)
	r.GET("removeTestdatas", __UpdateTestdata, removeTestdatas)
	r.GET("getTestdatas", __ScanTestdata, getTestdatas)
	r.GET("showOneTestdata", __ScanTestdata, showOneTestdata)
	r.GET("copyProblem", __CopyProblem, copyProblem)
	r.GET("delProblem", __DeleteProblem, delProblem)
	r.POST("updateProblemTags", __UpdateProblem, updateProblemTags)
	r.POST("downloadDatas", __ScanTestdata, downloadDatas)
	r.POST("newContest", __CreateContest, newContest)
	r.POST("updateContest", __UpdateContest, updateContest)
	r.POST("addCproblems", __UpdateContest, addCproblems)
	r.GET("deleteContest", __DeleteProblem, deleteContest)
	r.POST("forCsubmissions", __ScanContestSubmission, forCsubmissions)
	r.POST("rejudge", __RejudgeContestProblem, rejudge)
	r.GET("getUsers", __ScanUserInfo, getUsers)

	r.POST("updatePrivileges", __SetUserPrivilege, updatePrivileges)
	g := r.Group("", AuthAdmin)
	{
		g.GET("forAdminPage", forAdminPage)
		g.GET("getProblemAllInfo", getProblemAllInfo)
		g.POST("getOneProblemInfoByJson", getOneProblemInfoByJson)
		g.POST("searchContests", searchContests)
		g.GET("getContestInfo", getContestInfo)

		g.GET("getCproblems", getCproblems)
		g.GET("searchProbelmTitle", searchProbelmTitle)

		g.GET("forCproblemLabels", forCproblemLabels)
		g.GET("forCproblemTitles", forCproblemTitles)
		g.GET("forRankList", forRankList)
		g.GET("getPrivileges", getPrivileges)

		g.POST("updateClarification", updateClarification)

		g.GET("newClass", newClass)
		g.GET("getClasses", getClasses)
		g.GET("updateClassInfo", updateClassInfo)
	}
}
