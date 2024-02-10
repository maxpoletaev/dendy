package consts

import "time"

const (
	Speed             = 1
	SampleSize        = 32
	FramesPerSecond   = 60 * Speed
	SamplesPerSecond  = 44100 * Speed
	CPUTicksPerSecond = 1789773 * Speed
	TicksPerSecond    = CPUTicksPerSecond * 3
	SamplesPerFrame   = SamplesPerSecond / FramesPerSecond
	TicksPerSample    = TicksPerSecond / SamplesPerSecond
	FrameDuration     = time.Second / FramesPerSecond
	AudioBufferSize   = SamplesPerFrame * 3
	DefaultRelayAddr  = "159.223.15.170:1234" // TODO: need FQDN for this
)
