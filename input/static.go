package input

import "github.com/maxpoletaev/dendy/internal/binario"

type DeviceType uint8

const (
	DeviceTypeJoystick DeviceType = iota
	DeviceTypeZapper
)

var (
	_ Device = (*StaticDevice)(nil)
)

type StaticDevice struct {
	t DeviceType
	v any
}

func NewStaticDevice(t DeviceType) *StaticDevice {
	switch t {
	case DeviceTypeJoystick:
		return &StaticDevice{t: t, v: NewJoystick()}
	case DeviceTypeZapper:
		return &StaticDevice{t: t, v: NewZapper()}
	default:
		panic("unreachable")
	}
}

func (d *StaticDevice) Read() uint8 {
	switch d.t {
	case DeviceTypeJoystick:
		return d.v.(*Joystick).Read()
	case DeviceTypeZapper:
		return d.v.(*Zapper).Read()
	default:
		panic("unreachable")
	}
}

func (d *StaticDevice) Write(data uint8) {
	switch d.t {
	case DeviceTypeJoystick:
		d.v.(*Joystick).Write(data)
	case DeviceTypeZapper:
		d.v.(*Zapper).Write(data)
	default:
		panic("unreachable")
	}
}

func (d *StaticDevice) Reset() {
	switch d.t {
	case DeviceTypeJoystick:
		d.v.(*Joystick).Reset()
	case DeviceTypeZapper:
		d.v.(*Zapper).Reset()
	default:
		panic("unreachable")
	}
}

func (d *StaticDevice) SaveState(w *binario.Writer) error {
	switch d.t {
	case DeviceTypeJoystick:
		return d.v.(*Joystick).SaveState(w)
	case DeviceTypeZapper:
		return d.v.(*Zapper).SaveState(w)
	default:
		panic("unreachable")
	}
}

func (d *StaticDevice) LoadState(r *binario.Reader) error {
	switch d.t {
	case DeviceTypeJoystick:
		return d.v.(*Joystick).LoadState(r)
	case DeviceTypeZapper:
		return d.v.(*Zapper).LoadState(r)
	default:
		panic("unreachable")
	}
}
