package main

import (
	"HappyOnlineJudge/app"
	"HappyOnlineJudge/common"
	"HappyOnlineJudge/dao"
	"encoding/json"
	"fmt"
)

func init() {
	x, err := common.GetContent("config.json")
	if err != nil {
		panic(err)
	}
	cfg := make(common.H)
	if err := json.Unmarshal([]byte(x), &cfg); err != nil {
		panic(err)
	}
	if err := common.Init(cfg); err != nil {
		panic(err)
	}

	if err := dao.Init(cfg); err != nil {
		panic(err)
	} else {
		fmt.Println("数据库初始化完成")
	}
	app.InitRouters()
	//if err := dao.Test(); err != nil {
	//	fmt.Println("测试失败:", err.Error())
	//} else {
	//	fmt.Println("测试通过")
	//}
}

func main() {
	//fmt.Println("gg")
}
