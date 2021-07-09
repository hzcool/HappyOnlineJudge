package dao

import (
	"HappyOnlineJudge/common"
	"HappyOnlineJudge/model"
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"xorm.io/core"
)

type H = map[string]interface{}

var (
	engine *xorm.Engine    //数据库引擎(这里用的mysql)
	rdb    *redis.Client   //redis
	ctx    context.Context //默认值
)

//连接mysql数据库和redis
func connect(cfg H) error {
	var err error
	if mysql, ok := cfg["mysql"].(H); !ok {
		return errors.New("读取mysql配置失败")
	} else {
		dataSourceName := mysql["name"].(string) + ":" + mysql["password"].(string) + "@tcp(" + mysql["host"].(string) + ")/" + mysql["database"].(string) + "?charset=utf8"
		fmt.Println(dataSourceName)
		//数据库连接, "root:root@/hoj_db?charset=utf8"
		engine, err = xorm.NewEngine("mysql", dataSourceName)
		if err != nil {
			return err
		}
		err = engine.Ping()
		if err != nil {
			return err
		}
		engine.SetMapper(core.GonicMapper{})
	}

	if rds, ok := cfg["redis"].(H); !ok {
		return errors.New("读取redis配置失败")
	} else {
		//redis连接
		rdb = redis.NewClient(&redis.Options{
			Addr:     rds["addr"].(string),     //"localhost:6379"
			Password: rds["password"].(string), // no password set
			DB:       0,                        // use default DB
		})
		ctx = context.TODO()
		if pong, err := rdb.Ping(ctx).Result(); err != nil {
			return err
		} else {
			fmt.Println(pong, err)
		}
	}
	return nil
}

// mysql表同步和redis初始化
func sync(cfg H) error {
	if err := engine.Sync2(new(model.User)); err != nil {
		return err
	}
	if err := engine.Sync2(new(model.Message)); err != nil {
		return err
	}
	if err := engine.Sync2(new(model.Post)); err != nil {
		return err
	}
	if err := engine.Sync2(new(model.Attitude)); err != nil {
		return err
	}
	if err := engine.Sync2(new(model.Reply)); err != nil {
		return err
	}
	if err := engine.Sync2(new(model.Comment)); err != nil {
		return err
	}
	if err := engine.Sync2(new(model.Problem)); err != nil {
		return err
	}
	if err := engine.Sync2(new(model.Submission)); err != nil {
		return err
	}
	if err := engine.Sync2(new(model.Contest)); err != nil {
		return err
	}
	if err := engine.Sync2(new(model.Team)); err != nil {
		return err
	}
	if err := engine.Sync2(new(model.Cproblem)); err != nil {
		return err
	}
	if err := engine.Sync2(new(model.Csubmission)); err != nil {
		return err
	}
	if err := engine.Sync2(new(model.Homework)); err != nil {
		return err
	}
	if err := engine.Sync2(new(model.Class)); err != nil {
		return err
	}
	if err := engine.Sync2(new(model.Student)); err != nil {
		return err
	}
	//if err := engine.Sync2(new(model.Announcement)); err != nil {
	//	return err
	//}

	userInitRedis()
	problemInitRedis()
	contestInit()
	//设置管理员
	if superAdmin, ok := cfg["super_admin"].(H); !ok {
		return errors.New("读取super_admin配置失败")
	} else {
		ud := &UserDao{Username: superAdmin["name"].(string)}
		if ud.GetID() == 0 {
			ud.User = &User{
				Username:     superAdmin["name"].(string),
				Password:     common.GetMD5Password(superAdmin["password"].(string)),
				IsSuperAdmin: true,
				IsAdmin:      true,
				Privilege:    (uint64(1) << 32) - 1,
				Email:        superAdmin["email"].(string),
				Avatar:       superAdmin["avatar"].(string),
			}
			if err := ud.Create(); err != nil {
				return err
			}
			fmt.Println("超级管理初始化创建完成!!!")
		}
	}
	return nil
}

func Init(cfg H) error {
	if err := connect(cfg); err != nil {
		return err
	}
	if err := sync(cfg); err != nil {
		return err
	}
	return nil
}
