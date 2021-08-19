package einx

import (
	"sync"

	"github.com/jslyzt/einx/agent"
	"github.com/jslyzt/einx/component"
	"github.com/jslyzt/einx/context"
	"github.com/jslyzt/einx/event"
	lua_state "github.com/jslyzt/einx/lua"
	"github.com/jslyzt/einx/module"
	"github.com/jslyzt/einx/network"
	"github.com/jslyzt/einx/timer"
)

type (
	Agent            = agent.Agent
	AgentID          = agent.AgentID
	Module           = context.Module
	Context          = context.Context
	EventMsg         = event.EventMsg
	EventType        = event.EventType
	ArgsVar          = module.ArgsVar
	MsgHandler       = module.MsgHandler
	RpcHandler       = module.RpcHandler
	WorkerPool       = module.WorkerPool
	Component        = component.Component
	ComponentID      = component.ComponentID
	ModuleRouter     = module.ModuleRouter
	ComponentMgr     = module.ComponentMgr
	SessionEventMsg  = event.SessionEventMsg
	LuaRuntime       = lua_state.LuaRuntime
	NetLinker        = network.NetLinker
	ProtoTypeID      = network.ProtoTypeID
	SessionMgr       = network.SessionMgr
	SessionHandler   = network.SessionHandler
	ITcpClientMgr    = network.ITcpClientMgr
	ITcpServerMgr    = network.ITcpServerMgr
	TimerHandler     = timer.TimerHandler
	EventReceiver    = event.EventReceiver
	ITranMsgMultiple = network.ITranMsgMultiple

	einx struct {
		endWait   sync.WaitGroup
		closeChan chan bool
		onClose   func()
	}
)

func (e *einx) doClose() {
	onClose := e.onClose
	if onClose != nil {
		onClose()
		<-e.closeChan
	}
}

func (e *einx) close() {
	module.Close()
	e.endWait.Wait()
}

func (e *einx) continueClose() {
	e.closeChan <- true
}
