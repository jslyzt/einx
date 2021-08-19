package timer

import (
	"time"
)

const TIMERWHEELCOUNT = 5

type TimerManager struct {
	seqIDIndex   uint32
	timer_wheels [5]*timerWheel
	pool         *timerPool
}

var beginTime = time.Now()

func UnixTS() uint64 {
	tick := time.Since(beginTime).Nanoseconds()
	return uint64(tick / 1e6)
}

func NewTimerManager() *TimerManager {

	m := &TimerManager{
		seqIDIndex: 0,
		pool:       newTimerPool(),
	}

	now := UnixTS()

	timer_wheels := &m.timer_wheels

	timer_wheels[0] = newTimerWheel(1, 0, now, m.pool)
	timer_wheels[1] = newTimerWheel(0xff+1, 8, now, m.pool)
	timer_wheels[2] = newTimerWheel(0xffff+1, 16, now, m.pool)
	timer_wheels[3] = newTimerWheel(0xffffff+1, 24, now, m.pool)
	timer_wheels[4] = newTimerWheel(0xffffffff+1, 32, now, m.pool)

	timer_wheels[0].next_wheel = timer_wheels[1]
	timer_wheels[1].prev_wheel = timer_wheels[0]
	timer_wheels[1].next_wheel = timer_wheels[2]
	timer_wheels[2].prev_wheel = timer_wheels[1]
	timer_wheels[2].next_wheel = timer_wheels[3]
	timer_wheels[3].prev_wheel = timer_wheels[2]
	timer_wheels[3].next_wheel = timer_wheels[4]
	timer_wheels[4].prev_wheel = timer_wheels[3]

	return m
}

func (mgr *TimerManager) GetSeqID() uint32 {
	mgr.seqIDIndex++
	if mgr.seqIDIndex == 0 || mgr.seqIDIndex >= 0xffffff {
		mgr.seqIDIndex = 1
	}
	return mgr.seqIDIndex
}

func (mgr *TimerManager) AddTimer(delay uint64, op TimerHandler, args ...interface{}) uint64 {
	seqID := mgr.GetSeqID()
	run_tick := UnixTS() + delay

	if run_tick > 0x000000ffffffffff {
		run_tick = run_tick & 0x000000ffffffffff
	}

	xtimer := mgr.pool.Get()
	xtimer.args = args
	xtimer.handler = op
	xtimer.next = nil
	xtimer.running = false
	xtimer.seqID = seqID
	xtimer.runTick = run_tick

	mgr.timer_wheels[4].add_timer(xtimer)

	return xtimer.getTimerId()
}

func (mgr *TimerManager) DeleteTimer(timerID uint64) bool {
	if timerID == 0 {
		return false
	}
	return mgr.timer_wheels[4].delete_timer(timerID>>24, uint32(timerID&0xffffff))
}

func (mgr *TimerManager) Execute(count uint32) int {
	now := UnixTS()
	wheel := mgr.timer_wheels[0]
	wheel.execute(now, count)
	return wheel.nextWake()
}
