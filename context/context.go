package context

import (
	"github.com/jslyzt/einx/agent"
	"github.com/jslyzt/einx/component"
)

type (
	Agent     = agent.Agent
	AgentID   = agent.AgentID
	Component = component.Component

	Context interface {
		GetModule() Module
		GetSender() Agent
		GetComponent() Component
		GetAttach() interface{}
		Store(int, interface{})
		Get(int) interface{}
		Done(args ...interface{})
	}
)
