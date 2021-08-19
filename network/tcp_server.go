package network

import (
	"net"
	"sync/atomic"
	"time"

	"github.com/jslyzt/einx/event"
	"github.com/jslyzt/einx/slog"
)

const TCP_ACCEPT_SLEEP = 150

type TcpServerMgr struct {
	name         string
	listener     net.Listener
	componentID  ComponentID
	module       EventReceiver
	agentHandler SessionHandler
	addr         string
	closeFlag    int32
	option       TransportOption
}

func NewTcpServerMgr(opts ...Option) Component {
	tcpServer := &TcpServerMgr{
		componentID: GenComponentID(),
		closeFlag:   0,
		option:      newTransportOption(),
	}

	for _, opt := range opts {
		opt(tcpServer)
	}

	if tcpServer.agentHandler == nil {
		panic("option agent handler is nil")
	}

	if tcpServer.module == nil {
		panic("option agent handler is nil")
	}

	return tcpServer
}

func (mgr *TcpServerMgr) GetID() ComponentID {
	return mgr.componentID
}

func (mgr *TcpServerMgr) GetType() ComponentType {
	return COMPONENT_TYPE_TCP_SERVER
}

func (mgr *TcpServerMgr) Address() net.Addr {
	if mgr.listener == nil {
		return nil
	}
	return mgr.listener.Addr()
}

func (mgr *TcpServerMgr) Start() bool {
	listener, err := net.Listen("tcp", mgr.addr)
	if err != nil {
		slog.LogError("tcp_server", "ListenTCP addr:[%s],Error:%s", mgr.addr, err.Error())
		return false
	}
	mgr.listener = listener
	go mgr.doTcpAccept()
	return true
}

func (mgr *TcpServerMgr) Close() {
	if atomic.CompareAndSwapInt32(&mgr.closeFlag, 0, 1) {
		if mgr.listener == nil {
			return
		}
		_ = mgr.listener.Close()
	}
}

func (mgr *TcpServerMgr) isRunning() bool {
	close_flag := atomic.LoadInt32(&mgr.closeFlag)
	return close_flag == 0
}

func (mgr *TcpServerMgr) doTcpAccept() {
	m := mgr.module
	h := mgr.agentHandler
	listener := mgr.listener

	for mgr.isRunning() {
		rawConn, err := listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				time.Sleep(TCP_ACCEPT_SLEEP)
			} else if mgr.isRunning() {
				slog.LogError("tcp_server", "Accept Error: %v", err)
			}
			continue
		}

		tcpAgent := newTcpConn(rawConn, h, Linker_TCP_InComming, &mgr.option)
		m.PostEvent(event.EVENT_TCP_ACCEPTED, tcpAgent, mgr.componentID)

		go func() {
			pingMgr.AddPing(tcpAgent)
			err := tcpAgent.Run()
			pingMgr.RemovePing(tcpAgent)
			m.PostEvent(event.EVENT_TCP_CLOSED, tcpAgent, mgr.componentID, err)
		}()
	}
}

func (mgr *TcpServerMgr) GetOption() *TransportOption {
	return &mgr.option
}
