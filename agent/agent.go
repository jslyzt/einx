package agent

import (
	"sync/atomic"
)

type (
	AgentID     = uint64
	ProtoTypeID = uint32
	EventType   = int

	Agent interface {
		GetID() AgentID
		Close()
	}
)

var agent_id uint64 = 0

func GenAgentID() AgentID {
	return AgentID(atomic.AddUint64(&agent_id, 1))
}
