package mysql

import (
	"database/sql"
	"time"

	"github.com/jslyzt/einx/component"
	"github.com/jslyzt/einx/event"
	"github.com/jslyzt/einx/module"
	"github.com/jslyzt/einx/slog"
)

// 类型定义
type (
	EventReceiver = event.EventReceiver

	MysqlMgr struct {
		session      *sql.DB
		timeout      time.Duration
		dbcfg        *MysqlConnInfo
		component_id component.ComponentID
		m            EventReceiver
	}
)

// NewMysqlMgr 新建
func NewMysqlMgr(m module.Module, dbcfg *MysqlConnInfo, timeout time.Duration) *MysqlMgr {
	return &MysqlMgr{
		session:      nil,
		timeout:      timeout,
		dbcfg:        dbcfg,
		component_id: component.GenComponentID(),
		m:            m.(event.EventReceiver),
	}
}

func (mgr *MysqlMgr) GetID() component.ComponentID {
	return mgr.component_id
}

func (mgr *MysqlMgr) GetType() component.ComponentType {
	return component.COMPONENT_TYPE_DB_MYSQL
}

func (mgr *MysqlMgr) Start() bool {
	var err error
	mgr.session, err = sql.Open("mysql", mgr.dbcfg.String())
	if err != nil {
		e := &event.ComponentEventMsg{}
		e.MsgType = event.EVENT_COMPONENT_ERROR
		e.Sender = mgr
		e.Attach = err
		mgr.m.PushEventMsg(e)
		slog.LogInfo("mysql", "mysql connect failed.")
		return false
	}
	return true
}

func (mgr *MysqlMgr) Close() {
	if mgr.session != nil {
		mgr.session.Close()
		mgr.session = nil
		slog.LogInfo("mysql", "mysql disconnect")
	}
}

func (mgr *MysqlMgr) Ping() error {
	if mgr.session != nil {
		return mgr.session.Ping()
	}
	return ErrSessionNil
}

func (mgr *MysqlMgr) GetSession() *sql.DB {
	return mgr.session
}

func (mgr *MysqlMgr) GetNamedRows(query interface{}) ([]map[string]interface{}, error) {
	return GetNamedRows(query)
}

func GetNamedRows(query interface{}) ([]map[string]interface{}, error) {
	row, ok := query.(*sql.Rows)
	if !ok {
		return nil, ErrGetNamedResult
	}
	var results []map[string]interface{}
	column_types, err := row.ColumnTypes()
	if err != nil {
		slog.LogError("mysql", "columns error:%v", err)
		return nil, err
	}

	values := make([]interface{}, len(column_types))

	for c := true; c || row.NextResultSet(); c = false {

		//maybe this way is better
		//for k, c := range column_types {
		//	scans[k] = reflect.New(c.ScanType()).Interface()
		//}

		for k, c := range column_types {
			switch c.DatabaseTypeName() {
			case "INT":
				values[k] = new(int32)
			case "BIGINT":
				values[k] = new(int64)
			case "DOUBLE", "FLOAT":
				values[k] = new(float64)
			case "VARCHAR":
				values[k] = new(string)
			case "BLOB":
				values[k] = new([]byte)
			default:
				values[k] = new([]byte)
			}
		}

		for row.Next() {
			if err = row.Scan(values...); err != nil {
				slog.LogError("mysql", "Scan error:%v", err)
				return nil, err
			}

			result := make(map[string]interface{})
			for k, v := range values {
				key := column_types[k]
				switch s := v.(type) {
				case *int64:
					result[key.Name()] = *s
				case *float64:
					result[key.Name()] = *s
				case *string:
					result[key.Name()] = *s
				case *[]byte:
					result[key.Name()] = *s
				}
			}
			results = append(results, result)
		}
	}
	_ = row.Close()
	return results, nil
}
