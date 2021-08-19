package network

import (
	"net"

	"github.com/jslyzt/einx/event"
	"github.com/jslyzt/einx/slog"
)

type TcpClientMgr struct {
	name          string
	component_id  ComponentID
	module        EventReceiver
	agent_handler SessionHandler
	option        TransportOption
}

func NewTcpClientMgr(opts ...Option) Component {
	tcp_client := &TcpClientMgr{
		component_id: GenComponentID(),
		option: TransportOption{
			msg_max_length: MSG_MAX_BODY_LENGTH,
			msg_max_count:  MSG_DEFAULT_COUNT,
		},
	}

	for _, opt := range opts {
		opt(tcp_client)
	}

	if tcp_client.agent_handler == nil {
		panic("option agent handler is nil")
	}

	if tcp_client.module == nil {
		panic("option agent handler is nil")
	}

	return tcp_client
}

func (mgr *TcpClientMgr) GetID() ComponentID {
	return mgr.component_id
}

func (mgr *TcpClientMgr) GetType() ComponentType {
	return COMPONENT_TYPE_TCP_CLIENT
}

func (mgr *TcpClientMgr) Start() bool {
	return true
}

func (mgr *TcpClientMgr) Close() {

}

func (mgr *TcpClientMgr) Connect(addr string, user_type interface{}) {
	go mgr.connect(addr, user_type)
}

func (mgr *TcpClientMgr) connect(addr string, user_type interface{}) {
	raw_conn, err := net.Dial("tcp", addr)
	if err != nil {
		slog.LogWarning("tcp_client", "tcp connect failed %v", err)
		e := &event.ComponentEventMsg{}
		e.MsgType = event.EVENT_COMPONENT_ERROR
		e.Sender = mgr
		e.Attach = user_type
		e.Err = err
		mgr.module.PushEventMsg(e)
		return
	}

	m := mgr.module
	h := mgr.agent_handler

	tcp_agent := newTcpConn(raw_conn, h, Linker_TCP_OutGoing, &mgr.option)
	tcp_agent.SetUserType(user_type)
	m.PostEvent(event.EVENT_TCP_CONNECTED, tcp_agent, mgr.component_id)

	go func() {
		pingMgr.AddPing(tcp_agent)
		err := tcp_agent.Run()
		pingMgr.RemovePing(tcp_agent)
		m.PostEvent(event.EVENT_TCP_CLOSED, tcp_agent, mgr.component_id, err)
	}()
}

func (mgr *TcpClientMgr) GetOption() *TransportOption {
	return &mgr.option
}
