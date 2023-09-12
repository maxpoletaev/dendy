package apu

import "math"

type filter interface {
	filter(input float32) float32
}

type butterworthFilter struct {
	b0    float32
	b1    float32
	a1    float32
	prevX float32
	prevY float32
}

func (f *butterworthFilter) filter(x float32) float32 {
	y := f.b0*x + f.b1*f.prevX - f.a1*f.prevY
	f.prevY = y
	f.prevX = x
	return y
}

func lowPassFilter(sampleRate, cutoffFrequency float32) *butterworthFilter {
	c := sampleRate / math.Pi / cutoffFrequency
	a0i := 1 / (1 + c)

	return &butterworthFilter{
		b0: a0i,
		b1: a0i,
		a1: (1 - c) * a0i,
	}
}

func highPassFilter(sampleRate, cutoffFrequency float32) *butterworthFilter {
	c := sampleRate / math.Pi / cutoffFrequency
	a0i := 1 / (1 + c)

	return &butterworthFilter{
		b0: c * a0i,
		b1: -c * a0i,
		a1: (1 - c) * a0i,
	}
}
