package mysql

import (
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
)

// 错误定义
var (
	ErrSessionNil     = errors.New("mysql mgr session nil")
	ErrNotFound       = errors.New("not found")
	ErrDbFindAll      = errors.New("mysql mgr found error")
	ErrGetNamedResult = errors.New("mysql get result type error")
)

// MysqlConnInfo 连接信息
type MysqlConnInfo struct {
	mysql_conn_info *mysql.Config
}

// NewMysqlConnInfo 新建连接信息
func NewMysqlConnInfo(host string, port int, name, user, pass string) *MysqlConnInfo {
	cfg := mysql.NewConfig()
	cfg.User = user
	cfg.Passwd = pass
	cfg.Addr = fmt.Sprintf("%s:%d", host, port)
	cfg.Net = "tcp"
	cfg.DBName = name
	cfg.MultiStatements = true
	cfg.InterpolateParams = true
	return &MysqlConnInfo{mysql_conn_info: cfg}
}

// String 字符串
func (con *MysqlConnInfo) String() string {
	return con.mysql_conn_info.FormatDSN()
}
