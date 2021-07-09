package dao

import (
	"HappyOnlineJudge/common"
	"encoding/json"
	"reflect"
	"time"
)

//提取数据库的某一列, 提供方法转化为对应类型,不提供错误,自己注意. 另外redis获取的结果都是字符串,需要特判转化
type Col struct {
	data interface{}
}

func (c *Col) ToString() string {
	if s, ok := c.data.(string); ok {
		return s
	}
	return string(c.data.([]byte))
}
func (c *Col) ToInt64() int64 {
	if s, ok := c.data.(string); ok {
		return common.StrToInt64(s)
	}
	return c.data.(int64)
}
func (c *Col) ToInt() int {
	if s, ok := c.data.(string); ok {
		return common.StrToInt(s)
	}
	return int(c.ToInt64())
}
func (c *Col) ToBool() bool {
	if s, ok := c.data.(string); ok {
		return common.StrToBool(s)
	}
	if c.data.(int64) == 1 {
		return true
	}
	return false
}
func (c *Col) ToUint() uint {
	if s, ok := c.data.(string); ok {
		return common.StrToUint(s)
	}
	return uint(c.ToInt64())
}
func (c *Col) ToUint64() uint64 {
	if s, ok := c.data.(string); ok {
		return common.StrToUint64(s)
	}
	return uint64(c.ToInt64())
}
func (c *Col) ToFloat64() float64 {
	if s, ok := c.data.(string); ok {
		return common.StrToFloat64(s)
	}
	return c.data.(float64)
}
func (c *Col) ToTime() time.Time {
	t := c.ToString()
	return common.StrToTime(t)
}

//数据库是json序列化的,所以需要反序列化
func (c *Col) GetByteSlice() []byte {
	if reflect.TypeOf(c.data).Kind() == reflect.String {
		return []byte(c.data.(string))
	}
	return c.data.([]byte)
}
func (c *Col) ToInt64Slice() []int64 {
	var res []int64
	json.Unmarshal(c.GetByteSlice(), &res)
	if res == nil {
		return make([]int64, 0)
	}
	return res
}
func (c *Col) ToInt64MapInt64Silce() map[int64][]int64 {
	var res map[int64][]int64
	json.Unmarshal(c.GetByteSlice(), &res)
	if res == nil {
		return make(map[int64][]int64)
	}
	return res
}
func (c *Col) ToStringSlice() []string {
	var res []string
	json.Unmarshal(c.GetByteSlice(), &res)
	if res == nil {
		return make([]string, 0)
	}
	return res
}
func (c *Col) ToStringMapString() map[string]string {
	var res map[string]string
	json.Unmarshal(c.GetByteSlice(), &res)
	if res == nil {
		return make(map[string]string)
	}
	return res
}
func (c *Col) ToStringMapAny() map[string]interface{} {
	res := make(map[string]interface{})
	json.Unmarshal(c.GetByteSlice(), &res)
	if res == nil {
		return make(map[string]interface{})
	}
	return res
}
func (c *Col) ToInt64MapInt64() map[int64]int64 {
	res := make(map[int64]int64, 0)
	json.Unmarshal(c.GetByteSlice(), &res)
	if res == nil {
		return make(map[int64]int64, 0)
	}
	return res
}
func (c *Col) ToStringMapInt64() map[string]int64 {
	res := make(map[string]int64, 0)
	json.Unmarshal(c.GetByteSlice(), &res)
	if res == nil {
		return make(map[string]int64, 0)
	}
	return res
}

//原生sql语句构造
func ToSqlConditions(cols []string) string {
	n := len(cols)
	sql := cols[0] + " = ?"
	for i := 1; i < n; i++ {
		sql += " and " + cols[i] + " = ?"
	}
	return sql
}
func ToSqlSelect(wants ...string) string {
	n := len(wants)
	sql := "select " + wants[0]
	for i := 1; i < n; i++ {
		sql += "," + wants[i]
	}
	return sql
}
