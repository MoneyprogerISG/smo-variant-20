package smo

import (
	"math/rand"
	"testing"
)

func newTestPool(m int) *DevicePool {
	rng := rand.New(rand.NewSource(1))
	devices := make([]*Device, m)
	for i := range devices {
		devices[i] = NewDevice(i+1, 1.0, rng)
	}
	return NewDevicePool(devices)
}

func occupy(d *Device) {
	d.StartService(NewRequest(1, 1, 0), 0)
}

func TestDevicePoolNew(t *testing.T) {
	p := newTestPool(3)

	if p.Pointer() != 0 {
		t.Errorf("Pointer() нового пула: получили %d, ожидали 0", p.Pointer())
	}
	if len(p.Devices()) != 3 {
		t.Errorf("len(Devices()): получили %d, ожидали 3", len(p.Devices()))
	}
	if !p.HasFreeDevice() {
		t.Errorf("HasFreeDevice() нового пула: получили false, ожидали true")
	}
}

func TestDevicePoolNewPanics(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	cases := map[string][]*Device{
		"nil":        nil,
		"пустой":     {},
		"nil-прибор": {NewDevice(1, 1.0, rng), nil},
	}

	for name, devices := range cases {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("NewDevicePool(%s): ожидали panic", name)
				}
			}()
			NewDevicePool(devices)
		}()
	}
}

func TestDevicePoolRoundRobin(t *testing.T) {
	p := newTestPool(3)

	wantIDs := []int{1, 2, 3, 1}
	for i, want := range wantIDs {
		d := p.FindFreeDevice()
		if d == nil {
			t.Fatalf("вызов %d: получили nil, ожидали прибор %d", i, want)
		}
		if d.ID() != want {
			t.Errorf("вызов %d: прибор %d, ожидали %d", i, d.ID(), want)
		}
	}
}

func TestDevicePoolSkipsBusy(t *testing.T) {
	p := newTestPool(3)
	occupy(p.Devices()[0])
	occupy(p.Devices()[1])

	d := p.FindFreeDevice()
	if d == nil || d.ID() != 3 {
		t.Fatalf("FindFreeDevice(): получили %v, ожидали прибор 3", d)
	}
	if p.Pointer() != 0 {
		t.Errorf("Pointer() после выбора прибора 3 из 3: получили %d, ожидали 0", p.Pointer())
	}
}

func TestDevicePoolWrapsAroundRing(t *testing.T) {
	p := newTestPool(3)
	p.pointer = 2
	occupy(p.Devices()[2])

	d := p.FindFreeDevice()
	if d == nil || d.ID() != 1 {
		t.Fatalf("FindFreeDevice(): получили %v, ожидали прибор 1", d)
	}
	if p.Pointer() != 1 {
		t.Errorf("Pointer(): получили %d, ожидали 1", p.Pointer())
	}
}

func TestDevicePoolPointerFollowsChosen(t *testing.T) {
	p := newTestPool(4)
	occupy(p.Devices()[0])
	occupy(p.Devices()[1])

	d := p.FindFreeDevice()
	if d == nil || d.ID() != 3 {
		t.Fatalf("FindFreeDevice(): получили %v, ожидали прибор 3", d)
	}
	if p.Pointer() != 3 {
		t.Errorf("Pointer(): получили %d, ожидали 3", p.Pointer())
	}
}

func TestDevicePoolAllBusy(t *testing.T) {
	p := newTestPool(3)
	for _, d := range p.Devices() {
		occupy(d)
	}
	p.pointer = 1

	if d := p.FindFreeDevice(); d != nil {
		t.Errorf("FindFreeDevice() при всех занятых: получили %v, ожидали nil", d)
	}
	if p.Pointer() != 1 {
		t.Errorf("Pointer() после неудачного поиска: получили %d, ожидали 1", p.Pointer())
	}
	if p.HasFreeDevice() {
		t.Errorf("HasFreeDevice() при всех занятых: получили true, ожидали false")
	}
}

func TestDevicePoolSingleDevice(t *testing.T) {
	p := newTestPool(1)

	for i := 0; i < 3; i++ {
		if d := p.FindFreeDevice(); d == nil || d.ID() != 1 {
			t.Fatalf("вызов %d: получили %v, ожидали прибор 1", i, d)
		}
		if p.Pointer() != 0 {
			t.Errorf("вызов %d: Pointer() %d, ожидали 0", i, p.Pointer())
		}
	}
}

func TestDevicePoolHasFreeDoesNotMovePointer(t *testing.T) {
	p := newTestPool(3)
	p.pointer = 2

	if !p.HasFreeDevice() {
		t.Errorf("HasFreeDevice(): получили false, ожидали true")
	}
	if p.Pointer() != 2 {
		t.Errorf("Pointer() после HasFreeDevice(): получили %d, ожидали 2", p.Pointer())
	}
}

func TestDevicePoolFreedDeviceIsFoundAgain(t *testing.T) {
	p := newTestPool(2)
	d1 := p.FindFreeDevice()
	occupy(d1)
	d2 := p.FindFreeDevice()
	occupy(d2)

	if p.HasFreeDevice() {
		t.Fatalf("HasFreeDevice() при обоих занятых: получили true, ожидали false")
	}

	d1.FinishService(1)
	if d := p.FindFreeDevice(); d != d1 {
		t.Errorf("FindFreeDevice() после освобождения: получили %v, ожидали первый прибор", d)
	}
}
