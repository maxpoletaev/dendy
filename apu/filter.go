package apu

import (
	"math"
)

// filter is a simple IIR filter. No idea how it works, though. This part was copied
// from fogleman’s nes emulator, since I miserably failed to implement it myself.
type filter struct {
	b0    float32
	b1    float32
	a1    float32
	prevX float32
	prevY float32
}

func (f *filter) do(x float32) float32 {
	y := f.b0*x + f.b1*f.prevX - f.a1*f.prevY
	f.prevX, f.prevY = x, y

	return y
}

func lowPassFilter(sampleRate float32, cutoffFreq float32) *filter {
	c := sampleRate / math.Pi / cutoffFreq
	a0i := 1 / (1 + c)

	return &filter{
		b0: a0i,
		b1: a0i,
		a1: (1 - c) * a0i,
	}
}

func highPassFilter(sampleRate float32, cutoffFreq float32) *filter {
	c := sampleRate / math.Pi / cutoffFreq
	a0i := 1 / (1 + c)

	return &filter{
		b0: c * a0i,
		b1: -c * a0i,
		a1: (1 - c) * a0i,
	}
}
