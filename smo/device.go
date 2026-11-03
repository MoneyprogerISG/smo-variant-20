package smo

import (
	"math"
	"math/rand"
)

type Device struct {
	id          int
	lambda      float64
	random      *rand.Rand
	current     *Request
	releaseTime float64
	busyTime    float64
}

func NewDevice(id int, lambda float64, rng *rand.Rand) *Device {
	if lambda <= 0 {
		panic("device: lambda must be positive")
	}
	if rng == nil {
		panic("device: rng must not be nil")
	}
	return &Device{
		id:     id,
		lambda: lambda,
		random: rng,
	}
}

func (d *Device) StartService(r *Request, now float64) float64 {
	if d.current != nil {
		panic("device: StartService on a busy device")
	}
	t := -(1 / d.lambda) * math.Log(1-d.random.Float64())
	r.SetServiceStartTime(now)
	d.current = r
	d.releaseTime = now + t
	d.busyTime += t
	return d.releaseTime
}

func (d *Device) FinishService(now float64) *Request {
	if d.current == nil {
		panic("device: FinishService on a free device")
	}
	finished := d.current
	finished.SetServiceEndTime(now)
	d.current = nil
	return finished
}

func (d *Device) IsFree() bool {
	return d.current == nil
}

func (d *Device) ID() int {
	return d.id
}

func (d *Device) ReleaseTime() float64 {
	return d.releaseTime
}

func (d *Device) Current() *Request {
	return d.current
}

func (d *Device) BusyTime() float64 {
	return d.busyTime
}
