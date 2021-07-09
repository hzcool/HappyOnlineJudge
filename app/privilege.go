package app

type PrivilegeType = uint64 //最多细分64个权限

const (
	ScanProblemSubmission PrivilegeType = 1 << iota
	ScanPrivateProblem
	CreateProblem
	UpdateProblem
	CopyProblem
	DeleteProblem

	ScanTestdata
	UpdateTestdata

	CreateContest
	UpdateContest
	DeleteContest
	ScanContestSubmission
	RejudgeContestProblem

	ScanUserInfo
	SetUserPrivilege
)

var PrivilegeDesc []string = []string{
	"查看公共题库的提交代码", //1
	"查看私有题库题面",    //2
	"创建题目",        //4
	"修改题目",        //8
	"拷贝题目",        //16
	"删除题目",        //32

	"查看测试数据", //64
	"修改数据",   //128

	"创建比赛",     //256
	"修改比赛",     //512
	"删除比赛",     //1024
	"查看比赛的代码",  //2048
	"重新测评考试题目", //4096

	"查看用户信息", // 8192
	"设置用户权限", //16384

}
