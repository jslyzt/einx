package module

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/jslyzt/einx/slog"
)

var worker_pools_map sync.Map

type WorkerPool interface {
	ForEachModule(func(m Module))
	RegisterRpcHandler(string, RpcHandler)
	RegisterHandler(ProtoTypeID, MsgHandler)
	RpcCall(string, ...interface{})
	Balancer() Module
	Const(string) Module
	Slot(int) Module
}

type ModuleWorkerPool struct {
	modules    []Module
	name       string
	balance_id uint32
	size       uint32
}

func CreateWorkers(name string, size int) WorkerPool {
	w := GetWorkerPool(name).(*ModuleWorkerPool)
	w.size = uint32(size)
	if w.modules == nil {
		w.modules = make([]Module, size)
	}
	for i := 0; i < size; i++ {
		m := NewModule(fmt.Sprintf("%s_worker_%d", name, i+1))
		w.modules[i] = m
	}
	return w
}

func GetWorkerPool(name string) WorkerPool {
	v, ok := worker_pools_map.Load(name)
	if ok {
		return v.(WorkerPool)
	} else {
		w := &ModuleWorkerPool{
			name:       name,
			balance_id: 0,
			size:       0,
		}
		worker_pools_map.Store(name, w)
		return w
	}
}

func (pool *ModuleWorkerPool) Start() {
	for _, m := range pool.modules {
		go func(m Module) { m.(ModuleWoker).Run(&wait_close) }(m)
	}
}

func (pool *ModuleWorkerPool) ForEachModule(f func(m Module)) {
	for _, v := range pool.modules {
		f(v)
	}
}

func (pool *ModuleWorkerPool) Close() {
	slog.LogInfo("worker_pool", "worker_pool [%v] will close.", pool.name)
	for _, v := range pool.modules {
		v.(ModuleWoker).Close()
	}
}

func (pool *ModuleWorkerPool) RegisterRpcHandler(name string, f RpcHandler) {
	for _, v := range pool.modules {
		v.(ModuleRouter).RegisterRpcHandler(name, f)
	}
}

func (pool *ModuleWorkerPool) RegisterHandler(type_id ProtoTypeID, f MsgHandler) {
	for _, v := range pool.modules {
		v.(ModuleRouter).RegisterHandler(type_id, f)
	}
}

func (pool *ModuleWorkerPool) RpcCall(name string, args ...interface{}) {
	var hashkey uint32 = 0
	length := len(name)
	if length > 0 {
		hashkey += uint32(name[0])
		hashkey += uint32(name[length-1])
		hashkey += uint32(name[(length-1)/2])
		hashkey += uint32(length)
	}
	m := pool.modules[hashkey%pool.size] //route the rpc to worker by a simple hash key
	m.RpcCall(name, args...)
}

func (pool *ModuleWorkerPool) Balancer() Module {
	idx := atomic.AddUint32(&pool.balance_id, 1) % pool.size
	return pool.modules[idx]
}

func (pool *ModuleWorkerPool) Const(n string) Module {
	var hashkey uint32 = 0
	length := len(n)
	if length > 0 {
		hashkey += uint32(n[0])
		hashkey += uint32(n[length-1])
		hashkey += uint32(n[(length-1)/2])
		hashkey += uint32(length)
	}
	return pool.modules[hashkey%pool.size]
}

func (pool *ModuleWorkerPool) Slot(n int) Module {
	return pool.modules[uint32(n)%pool.size]
}
