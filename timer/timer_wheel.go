package timer

const TIMERSLOTSCOUNT = 256

type timerWheel struct {
	array    [TIMERSLOTSCOUNT]*timerList
	index    uint8
	bitSize  uint32
	msUnit   uint64
	baseTick uint64

	timerCount uint64

	next_wheel *timerWheel
	prev_wheel *timerWheel
}

func newTimerWheel(ms_unit uint64, bit_size uint32, now uint64, p *timerPool) *timerWheel {
	timer_wheel := &timerWheel{
		index:      0,
		bitSize:    bit_size,
		msUnit:     ms_unit,
		next_wheel: nil,
		prev_wheel: nil,
	}

	if ms_unit == 1 {
		timer_wheel.baseTick = now
	} else {
		timer_wheel.baseTick = now + ms_unit
	}

	for i := 0; i < TIMERSLOTSCOUNT; i++ {
		timer_wheel.array[i] = newTimerList(p)
	}

	return timer_wheel
}

func (tw *timerWheel) tickIdxDelta(runTick uint64) uint8 {
	idxDelta := runTick - tw.baseTick
	idxDelta = idxDelta >> tw.bitSize
	return uint8(idxDelta)
}

func (tw *timerWheel) add_timer(timer *xtimer) {
	if tw.prev_wheel != nil && timer.runTick < tw.baseTick {
		tw.prev_wheel.add_timer(timer)
		return
	}

	idx := uint8(tw.index + tw.tickIdxDelta(timer.runTick))
	timer_list := tw.array[idx]
	timer_list.addTimer(timer)
	tw.timerCount++
}

func (tw *timerWheel) delete_timer(run_tick uint64, seq_id uint32) bool {
	if tw.prev_wheel != nil && run_tick < tw.baseTick {
		return tw.prev_wheel.delete_timer(run_tick, seq_id)
	}
	idx := (tw.index + uint8(tw.tickIdxDelta(run_tick)))
	timer_list := tw.array[idx]
	success := timer_list.deleteTimer(seq_id)
	tw.timerCount--
	return success
}

func (tw *timerWheel) execute(now uint64, count uint32) uint32 {
	if tw.prev_wheel != nil || now < tw.baseTick {
		return 0
	}

	elapsedTime := uint64(now - tw.baseTick)
	loopTimes := uint64(1 + elapsedTime)

	var run_count uint32 = 0
	for ; run_count < count && loopTimes > 0; loopTimes-- {
		timer_list := tw.array[tw.index]
		c, b := timer_list.execute(now, count-run_count)
		run_count = run_count + c
		if !b {
			return run_count
		}
		tw.index++
		tw.baseTick += tw.msUnit
		if tw.index == 0 {
			tw.next_wheel.TurnWheel()
		}
	}
	return run_count
}

func (tw *timerWheel) nextWake() int {
	wakeDelay := 0
	wakeIndex := tw.index
	for wakeDelay < 64 {
		timer_list := tw.array[wakeIndex]
		if timer_list.head != nil {
			break
		}
		if wakeIndex++; wakeIndex == 0 {
			break
		}
		wakeDelay++
	}
	return wakeDelay
}

func (tw *timerWheel) TurnWheel() {
	if tw.prev_wheel == nil {
		return
	}

	timer_list := tw.array[tw.index]
	head_timer := timer_list.head
	var next_timer *xtimer = nil
	for head_timer != nil {
		next_timer = head_timer.next
		head_timer.next = nil
		tw.prev_wheel.add_timer(head_timer)
		head_timer = next_timer
		tw.timerCount--
	}

	timer_list.head = nil
	timer_list.tail = nil

	tw.index++
	tw.baseTick += tw.msUnit

	if tw.index == 0 && tw.next_wheel != nil {
		tw.next_wheel.TurnWheel()
	}
}
