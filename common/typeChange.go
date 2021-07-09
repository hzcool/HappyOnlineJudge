package common

import (
	"reflect"
	"strconv"
	"time"
)

const (
	TIME_FORMAT = "2006-01-02 15:04:05"
)

var (
	cstZone = time.FixedZone("CST", 8*3600)
)

//下面转换不进行错误处理

//字符串转换成整型
func StrToInt(s string) int {
	ret, _ := strconv.ParseInt(s, 10, 64)
	return int(ret)
}

//字符串转64位整型
func StrToInt64(s string) int64 {
	ret, _ := strconv.ParseInt(s, 10, 64)
	return ret
}

//字符串转无符号整型
func StrToUint(s string) uint {
	ret, _ := strconv.ParseUint(s, 10, 64)
	return uint(ret)
}

//字符串转64位无符号整型
func StrToUint64(s string) uint64 {
	ret, _ := strconv.ParseUint(s, 10, 64)
	return ret
}
func StrToBool(s string) bool {
	ret, _ := strconv.ParseBool(s)
	return ret
}
func StrToFloat64(s string) float64 {
	ret, _ := strconv.ParseFloat(s, 64)
	return ret
}
func StrToTime(s string) time.Time {
	t, _ := time.ParseInLocation(TIME_FORMAT, s, time.Local)
	return t
}
func TimeToStr(t time.Time) string {
	return t.In(cstZone).Format(TIME_FORMAT)
}

//将字符串按照json tag 转换成 map
func StructToMapByJsonTag(obj interface{}) map[string]interface{} {
	rType := reflect.TypeOf(obj)
	rVal := reflect.ValueOf(obj)
	if rType.Kind() == reflect.Ptr {
		rType = rType.Elem()
		rVal = rVal.Elem()
	}
	mp := make(map[string]interface{})
	for i := 0; i < rType.NumField(); i++ {
		t := rType.Field(i)
		value := rVal.Field(i).Interface()
		tag := t.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		switch value.(type) {
		case time.Time:
			mp[tag] = value.(time.Time).Format(TIME_FORMAT)
		default:
			mp[tag] = value
		}
	}
	return mp
}
