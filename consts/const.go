package consts

import "time"

const (
	Speed             = 1
	FramesPerSecond   = 60 * Speed
	CPUTicksPerSecond = 1789773 * Speed
	TicksPerSecond    = CPUTicksPerSecond * 3
	FrameDuration     = time.Second / FramesPerSecond

	AudioSampleSize       = 32
	AudioSamplesPerSecond = 44100 * Speed
	AudioSamplesPerFrame  = AudioSamplesPerSecond / FramesPerSecond
	TicksPerAudioSample   = TicksPerSecond / AudioSamplesPerSecond
	AudioBufferSize       = AudioSamplesPerFrame * 3

	DefaultRelayAddr = "159.223.15.170:1234" // TODO: need FQDN for this
)
