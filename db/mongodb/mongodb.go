package mongodb

import (
	"time"

	"github.com/jslyzt/einx/component"
	"github.com/jslyzt/einx/event"
	"github.com/jslyzt/einx/module"
	"github.com/jslyzt/einx/slog"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type (
	M             = bson.M
	EventReceiver = event.EventReceiver
)

const (
	Strong    = 1
	Monotonic = 2
)

type MongoDBMgr struct {
	session      *mgo.Session
	timeout      time.Duration
	dbcfg        *MongoDBInfo
	component_id component.ComponentID
	m            EventReceiver
}

func NewMongoDBMgr(m module.Module, dbcfg *MongoDBInfo, timeout time.Duration) *MongoDBMgr {
	return &MongoDBMgr{
		session:      nil,
		timeout:      timeout,
		dbcfg:        dbcfg,
		component_id: component.GenComponentID(),
		m:            m.(event.EventReceiver),
	}
}

func (mgr *MongoDBMgr) GetID() component.ComponentID {
	return mgr.component_id
}

func (mgr *MongoDBMgr) GetType() component.ComponentType {
	return component.COMPONENT_TYPE_DB_MONGODB
}

func (mgr *MongoDBMgr) Start() bool {
	var err error
	mgr.session, err = mgo.DialWithTimeout(mgr.dbcfg.String(), mgr.timeout)
	if err != nil {
		e := &event.ComponentEventMsg{}
		e.MsgType = event.EVENT_COMPONENT_ERROR
		e.Sender = mgr
		e.Attach = err
		mgr.m.PushEventMsg(e)
		slog.LogInfo("mongodb", "MongoDB Connect failed.")
		return false
	}

	mgr.session.SetMode(mgo.Monotonic, true)
	return true
}

func (mgr *MongoDBMgr) Close() {
	if mgr.session != nil {
		mgr.session.DB("").Logout()
		mgr.session.Close()
		mgr.session = nil
		slog.LogInfo("mongodb", "Disconnect mongodb url: ", mgr.dbcfg.String())
	}
}

func (mgr *MongoDBMgr) Ping() error {
	if mgr.session != nil {
		return mgr.session.Ping()
	}
	return ErrSessionNil
}

func (mgr *MongoDBMgr) RefreshSession() {
	mgr.session.Refresh()
}

func (mgr *MongoDBMgr) GetDbSession() *mgo.Session {
	return mgr.session
}

func (mgr *MongoDBMgr) SetMode(mode int, refresh bool) {
	status := mgo.Monotonic
	if mode == Strong {
		status = mgo.Strong
	} else {
		status = mgo.Monotonic
	}

	mgr.session.SetMode(status, refresh)
}

func (mgr *MongoDBMgr) Insert(collection string, doc interface{}) error {
	if mgr.session == nil {
		return ErrSessionNil
	}

	db_session := mgr.session.Copy()
	defer db_session.Close()

	c := db_session.DB("").C(collection)

	return c.Insert(doc)
}

func (mgr *MongoDBMgr) Update(collection string, cond interface{}, change interface{}) error {
	if mgr.session == nil {
		return ErrSessionNil
	}

	db_session := mgr.session.Copy()
	defer db_session.Close()

	c := db_session.DB("").C(collection)

	return c.Update(cond, bson.M{"$set": change})
}

func (mgr *MongoDBMgr) UpdateInsert(collection string, cond interface{}, doc interface{}) error {
	if mgr.session == nil {
		return ErrSessionNil
	}

	db_session := mgr.session.Copy()
	defer db_session.Close()

	c := db_session.DB("").C(collection)
	_, err := c.Upsert(cond, bson.M{"$set": doc})
	if err != nil {
		slog.LogInfo("mongodb", "UpdateInsert failed collection is:%s. cond is:%v", collection, cond)
	}

	return err
}

func (mgr *MongoDBMgr) RemoveOne(collection string, cond_name string, cond_value int64) error {
	if mgr.session == nil {
		return ErrSessionNil
	}

	db_session := mgr.session.Copy()
	defer db_session.Close()

	c := db_session.DB("").C(collection)
	err := c.Remove(bson.M{cond_name: cond_value})
	if err != nil && err != mgo.ErrNotFound {
		slog.LogInfo("mongodb", "remove failed from collection:%s. name:%s-value:%d", collection, cond_name, cond_value)
	}

	return err
}

func (mgr *MongoDBMgr) RemoveOneByCond(collection string, cond interface{}) error {
	if mgr.session == nil {
		return ErrSessionNil
	}

	db_session := mgr.session.Copy()
	defer db_session.Close()

	c := db_session.DB("").C(collection)
	err := c.Remove(cond)

	if err != nil && err != mgo.ErrNotFound {
		slog.LogInfo("mongodb", "remove failed from collection:%s. cond :%v, err: %v.", collection, cond, err)
	}

	return err
}

func (mgr *MongoDBMgr) RemoveAll(collection string, cond interface{}) error {
	if mgr.session == nil {
		return ErrSessionNil
	}

	db_session := mgr.session.Copy()
	defer db_session.Close()

	c := db_session.DB("").C(collection)
	change, err := c.RemoveAll(cond)
	if err != nil && err != mgo.ErrNotFound {
		slog.LogInfo("mongodb", "MongoDBMgr RemoveAll failed : %s, %v", collection, cond)
		return err
	}
	slog.LogInfo("mongodb", "MongoDBMgr RemoveAll: %v, %v", change.Updated, change.Removed)
	return nil
}

func (mgr *MongoDBMgr) DBQuery(collection string, cond interface{}, result *[]map[string]interface{}) error {
	if mgr.session == nil {
		return ErrSessionNil
	}

	db_session := mgr.session.Copy()
	defer db_session.Close()

	c := db_session.DB("").C(collection)
	q := c.Find(cond)

	if nil == q {
		return ErrDbFindAll
	}

	q.All(result)
	return nil
}

func (mgr *MongoDBMgr) DBQueryOneResult(collection string, cond interface{}, result map[string]interface{}) error {
	if mgr.session == nil {
		return ErrSessionNil
	}

	db_session := mgr.session.Copy()
	defer db_session.Close()

	c := db_session.DB("").C(collection)
	q := c.Find(cond)

	if nil == q {
		return ErrDbFindAll
	}

	q.One(result)
	return nil
}
