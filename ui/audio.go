package ui

import (
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type AudioOut struct {
	stream   rl.AudioStream
	volume   float32
	muted    bool
	channels int
}

func CreateAudio(sampleRate, sampleSize, channels, bufferSize int) *AudioOut {
	rl.SetAudioStreamBufferSizeDefault(int32(bufferSize))
	rl.InitAudioDevice()

	stream := rl.LoadAudioStream(uint32(sampleRate), uint32(sampleSize), uint32(channels))
	rl.SetAudioStreamVolume(stream, 1.0)
	rl.PlayAudioStream(stream)

	return &AudioOut{
		channels: channels,
		stream:   stream,
		volume:   1.0,
	}
}

func (s *AudioOut) SetVolume(volume float32) {
	if volume > 1.0 {
		volume = 1.0
	}

	rl.SetMasterVolume(volume)
	s.volume = volume
}

func (s *AudioOut) Close() {
	rl.StopAudioStream(s.stream)
	rl.CloseAudioDevice()
}

func (s *AudioOut) IsStreamProcessed() bool {
	return rl.IsAudioStreamProcessed(s.stream)
}

func (s *AudioOut) WaitStreamProcessed() {
	for !rl.IsAudioStreamProcessed(s.stream) {
		time.Sleep(time.Millisecond * 16)
	}
}

func (s *AudioOut) UpdateStream(buf []float32) {
	rl.UpdateAudioStream(s.stream, buf)
}

func (s *AudioOut) Mute(m bool) {
	s.muted = m

	if s.muted {
		rl.SetMasterVolume(0)
	} else {
		rl.SetMasterVolume(s.volume)
	}
}

func (s *AudioOut) ToggleMute() {
	s.muted = !s.muted

	if s.muted {
		rl.SetMasterVolume(0)
	} else {
		rl.SetMasterVolume(s.volume)
	}
}
