package common

import (
	"errors"
)

type H = map[string]interface{}

var Avatars []string
var WebHttp string

func Init(cfg H) error {
	if err := initRsaKeys(cfg["rsa"].(H)); err != nil {
		return err
	}
	var ok1 bool
	WebHttp, ok1 = cfg["address"].(string)
	if !ok1 {
		return errors.New("网址加载错误")
	}
	if tmp, ok := cfg["avatars"].([]interface{}); !ok {
		return errors.New("没有头像信息")
	} else {
		Avatars = make([]string, len(tmp))
		for i, item := range tmp {
			var ok2 bool
			Avatars[i], ok2 = item.(string)
			if !ok2 {
				return errors.New("头像信息错误")
			}
		}
	}

	if judge_cfg, ok := cfg["judger_env"].(map[string]interface{}); !ok {
		return errors.New("测评姬环境信息配置错误")
	} else {
		initJudger(judge_cfg)
	}

	return nil
}
