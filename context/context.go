package context

import (
	"github.com/jslyzt/einx/agent"
	"github.com/jslyzt/einx/component"
)

type Agent = agent.Agent
type AgentID = agent.AgentID
type Component = component.Component

type Context interface {
	GetModule() Module
	GetSender() Agent
	GetComponent() Component
	GetAttach() interface{}
	Store(int, interface{})
	Get(int) interface{}
	Done(args ...interface{})
}
