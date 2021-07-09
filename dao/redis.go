package dao

import (
	"HappyOnlineJudge/common"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"time"
)

//将各个域存到redis中 obj必须是结构体指针,按照json标签存入redis,  expire为0时永久保存
func typeAnalyzed(x interface{}) interface{} {
	switch x.(type) {
	case string, int64, int, uint, uint64, bool, float32, float64, []byte:
		return x
	case time.Time:
		t := x.(time.Time)
		return common.TimeToStr(t)
	default:
		jsonValue, _ := json.Marshal(x)
		return jsonValue
	}
}
func putObjToRedis(key string, obj interface{}, expire time.Duration) error {
	objType := reflect.TypeOf(obj)
	objVal := reflect.ValueOf(obj)
	if objType.Kind() == reflect.Ptr {
		if objVal.IsNil() {
			return errors.New("空指针错误")
		}
		objType = objType.Elem()
		objVal = objVal.Elem()
		if objType.Kind() != reflect.Struct {
			return errors.New("传入的不是结构体")
		}
	} else {
		return errors.New("传入对象不是结构体指针")
	}
	var args []interface{}
	num := objType.NumField()
	for i := 0; i < num; i++ {
		t := objType.Field(i)
		v := objVal.Field(i)
		tag := t.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		args = append(args, tag, typeAnalyzed(v.Interface()))
	}
	if _, err := rdb.HMSet(ctx, key, args...).Result(); err != nil {
		return err
	}
	if expire != 0 {
		rdb.Expire(ctx, key, expire)
	}
	return nil
}

//从redis中获取结构体对象,obj必须是结构体指针,按照json标签读取结构体
func GetObjFromRedis(key string, obj interface{}) error {
	objType := reflect.TypeOf(obj)
	objVal := reflect.ValueOf(obj)
	if objType.Kind() == reflect.Ptr {
		if objVal.IsNil() {
			return errors.New("空指针错误")
		}
		objType = objType.Elem()
		if objType.Kind() != reflect.Struct {
			return errors.New("传入的不是结构体")
		}
	} else {
		return errors.New("传入对象不是结构体指针")
	}
	var mp map[string]string
	var err error
	mp, err = rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return err
	}
	v := reflect.Indirect(objVal)
	num := v.NumField()
	for i := 0; i < num; i++ {
		valueInterface := v.Field(i).Interface()
		tag := objType.Field(i).Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		rawValue, ok := mp[tag]
		if !ok {
			continue
		}
		switch valueInterface.(type) {
		case string:
			v.Field(i).SetString(rawValue)
		case int64, int:
			v.Field(i).SetInt(common.StrToInt64(rawValue))
		case uint64, uint:
			v.Field(i).SetUint(common.StrToUint64(rawValue))
		case bool:
			v.Field(i).SetBool(common.StrToBool(rawValue))
		case float64, float32:
			num, _ := strconv.ParseFloat(rawValue, 64)
			v.Field(i).SetFloat(num)
		case time.Time:
			v.Field(i).Set(reflect.ValueOf(common.StrToTime(rawValue)))
		case []int64:
			var x []int64
			if err := json.Unmarshal([]byte(rawValue), &x); err != nil {
				return err
			}
			v.Field(i).Set(reflect.ValueOf(x))
		case []string:
			var x []string
			if err := json.Unmarshal([]byte(rawValue), &x); err != nil {
				return err
			}
			v.Field(i).Set(reflect.ValueOf(x))
		case map[string]int64:
			x := make(map[string]int64)
			if err := json.Unmarshal([]byte(rawValue), &x); err != nil {
				return err
			}
			v.Field(i).Set(reflect.ValueOf(x))
		case map[string]interface{}:
			x := make(map[string]interface{})
			if err := json.Unmarshal([]byte(rawValue), &x); err != nil {
				return err
			}
			v.Field(i).Set(reflect.ValueOf(x))
		case map[int64]int64:
			var x map[int64]int64
			if err := json.Unmarshal([]byte(rawValue), &x); err != nil {
				return err
			}
			v.Field(i).Set(reflect.ValueOf(x))

		default:
			return errors.New("出现了未讨论的redis类型")
		}
	}
	return nil
}

func HMSetMap(key string, mp map[string]interface{}, expire time.Duration) error {
	var args []interface{}
	for k, v := range mp {
		args = append(args, k, typeAnalyzed(v))
	}
	if _, err := rdb.HMSet(ctx, key, args...).Result(); err != nil {
		return err
	}
	if expire > 0 {
		rdb.Expire(ctx, key, expire)
	}
	return nil
}
func SetKeyExpire(key string, expire time.Duration) {
	rdb.Expire(ctx, key, expire)
}
