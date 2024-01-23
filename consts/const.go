package consts

const (
	SampleSize        = 32
	FramesPerSecond   = 60
	SamplesPerSecond  = 44100
	CPUTicksPerSecond = 1789773
	TicksPerSecond    = CPUTicksPerSecond * 3
	SamplesPerFrame   = SamplesPerSecond / FramesPerSecond
	TicksPerSample    = TicksPerSecond / SamplesPerSecond
	AudioBufferSize   = SamplesPerFrame * 3
)
