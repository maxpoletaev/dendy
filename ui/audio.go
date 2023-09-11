package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type AudioOut struct {
	stream   rl.AudioStream
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
	}
}

func (s *AudioOut) SetVolume(volume float32) {
	rl.SetMasterVolume(volume)
}

func (s *AudioOut) Mute() {
	rl.SetMasterVolume(0)
}

func (s *AudioOut) Close() {
	rl.StopAudioStream(s.stream)
	rl.CloseAudioDevice()
}

func (s *AudioOut) UpdateStream(buf []float32) {
	for !rl.IsAudioStreamProcessed(s.stream) {
		// sync to audio stream
	}

	rl.UpdateAudioStream(s.stream, buf)
}
