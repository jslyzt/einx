package module

import (
	"sync"

	"github.com/jslyzt/einx/agent"
	"github.com/jslyzt/einx/event"
	"github.com/jslyzt/einx/timer"
)

//var module_map map[string]Module = make(map[string]Module)
var module_map sync.Map
var wait_close sync.WaitGroup
var PerfomancePrint bool = false

func GenModuleID() AgentID {
	return agent.GenAgentID()
}

func Close() {
	module_map.Range(func(k interface{}, m interface{}) bool {
		m.(ModuleWoker).Close()
		return true
	})

	worker_pools_map.Range(func(k interface{}, m interface{}) bool {
		m.(*ModuleWorkerPool).Close()
		return true
	})

	wait_close.Wait()
}

func NewModule(name string) Module {
	m := &module{
		id:            GenModuleID(),
		evQueue:       event.NewEventQueue(),
		name:          name,
		timerManager:  timer.NewTimerManager(),
		msgHandlerMap: make(map[ProtoTypeID]MsgHandler),
		rpcHandlerMap: make(map[string]RpcHandler),
		agentMap:      make(map[AgentID]Agent),
		commgrMap:     make(map[ComponentID]ComponentMgr),
		componentMap:  make(map[ComponentID]Component),
		rpcMsgPool:    &sync.Pool{New: func() interface{} { return new(RpcEventMsg) }},
		dataMsgPool:   &sync.Pool{New: func() interface{} { return new(DataEventMsg) }},
		eventMsgPool:  &sync.Pool{New: func() interface{} { return new(SessionEventMsg) }},
		awaitMsgPool:  &sync.Pool{New: func() interface{} { return new(AwaitRpcEventMsg) }},
		closeChan:     make(chan bool),
		eventList:     make([]interface{}, MODULE_EVENT_LENGTH),
	}
	m.context = &ModuleContext{m: m}
	return m
}

func GetModule(name string) Module {
	var m Module
	v, ok := module_map.Load(name)
	if ok == false {
		m = NewModule(name)
		module_map.Store(name, m)
	} else {
		m = v.(Module)
	}
	return m
}

func FindModule(name string) Module {
	var m Module
	v, ok := module_map.Load(name)
	if ok == true {
		m = v.(Module)
		return m
	}
	return nil
}

func Start() {
	module_map.Range(func(k interface{}, m interface{}) bool {
		go func(m interface{}) { m.(ModuleWoker).Run(&wait_close) }(m)
		return true
	})

	worker_pools_map.Range(func(k interface{}, m interface{}) bool {
		m.(*ModuleWorkerPool).Start()
		return true
	})
}
