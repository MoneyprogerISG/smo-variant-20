package smo

type DevicePool struct {
	devices []*Device
	pointer int
}

func NewDevicePool(devices []*Device) *DevicePool {
	if len(devices) == 0 {
		panic("device_pool: devices must not be empty")
	}
	for _, d := range devices {
		if d == nil {
			panic("device_pool: device must not be nil")
		}
	}
	return &DevicePool{
		devices: devices,
	}
}

func (p *DevicePool) FindFreeDevice() *Device {
	m := len(p.devices)
	for k := 0; k < m; k++ {
		j := (p.pointer + k) % m
		if p.devices[j].IsFree() {
			p.pointer = (j + 1) % m
			return p.devices[j]
		}
	}
	return nil
}

func (p *DevicePool) HasFreeDevice() bool {
	for _, d := range p.devices {
		if d.IsFree() {
			return true
		}
	}
	return false
}

func (p *DevicePool) Pointer() int {
	return p.pointer
}

func (p *DevicePool) Devices() []*Device {
	return p.devices
}
