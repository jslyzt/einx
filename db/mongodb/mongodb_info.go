package mongodb

import (
	"errors"
	"fmt"
)

// 错误定义
var (
	ErrSessionNil = errors.New("MongoDBMgr session nil")
	ErrNotFound   = errors.New("not found")
	ErrDbFindAll  = errors.New("MongoDBMgr found error")
)

// MongoDBInfo mongo db info
type MongoDBInfo struct {
	DbHost string
	DbPort int
	DbName string
	DbUser string
	DbPass string
}

// NewMongoDBInfo new func
func NewMongoDBInfo(host string, port int, name, user, pass string) *MongoDBInfo {
	return &MongoDBInfo{
		DbHost: host,
		DbPort: port,
		DbName: name,
		DbUser: user,
		DbPass: pass,
	}
}

// String 字符串
func (mgr *MongoDBInfo) String() string {
	url := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
		mgr.DbUser, mgr.DbPass, mgr.DbHost, mgr.DbPort, mgr.DbName)
	if mgr.DbUser == "" || mgr.DbPass == "" {
		url = fmt.Sprintf("mongodb://%s:%d/%s", mgr.DbHost, mgr.DbPort, mgr.DbName)
	}
	return url
}
