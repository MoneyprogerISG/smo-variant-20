package smo

type SelectionDispatcher struct {
	buffer     *Buffer
	devicePool *DevicePool
	calendar   *EventCalendar
	statistics *Statistics
}

func NewSelectionDispatcher(b *Buffer, p *DevicePool, c *EventCalendar, s *Statistics) *SelectionDispatcher {
	if b == nil {
		panic("selection: buffer must not be nil")
	}
	if p == nil {
		panic("selection: device pool must not be nil")
	}
	if c == nil {
		panic("selection: calendar must not be nil")
	}
	if s == nil {
		panic("selection: statistics must not be nil")
	}
	return &SelectionDispatcher{
		buffer:     b,
		devicePool: p,
		calendar:   c,
		statistics: s,
	}
}

func (d *SelectionDispatcher) OnDeviceFreed(now float64) {
	if d.buffer.IsEmpty() {
		return
	}

	r := d.selectRequest()
	d.statistics.RecordWait(r, now-r.bufferInTime)

	dev := d.selectDevice()
	if dev == nil {
		panic("selection: no free device after release")
	}

	releaseTime := dev.StartService(r, now)
	d.calendar.Schedule(NewReleaseEvent(releaseTime, dev))
}

func (d *SelectionDispatcher) selectRequest() *Request {
	return d.buffer.TakeFirst()
}

func (d *SelectionDispatcher) selectDevice() *Device {
	return d.devicePool.FindFreeDevice()
}
