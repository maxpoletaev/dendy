package apu

type filter interface {
	filter(input float32) float32
}

type filterChain struct {
	filters []filter
}

func (c *filterChain) filter(input float32) float32 {
	for _, f := range c.filters {
		input = f.filter(input)
	}

	return input
}

type lowPassFilter struct {
	previous float32
	alpha    float32
}

func (f *lowPassFilter) filter(input float32) float32 {
	output := f.alpha*input + (1.0-f.alpha)*f.previous
	f.previous = output
	return output
}
