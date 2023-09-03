package screen

import (
	"log"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type AudioOut struct {
	stream   rl.AudioStream
	source   <-chan float32
	buffer   []float32
	channels int
}

func CreateAudio(sampleRate, sampleSize, channels, bufferSize int) *AudioOut {
	rl.SetAudioStreamBufferSizeDefault(int32(bufferSize))
	rl.InitAudioDevice()

	stream := rl.LoadAudioStream(uint32(sampleRate), uint32(sampleSize), uint32(channels))
	rl.SetAudioStreamVolume(stream, 0.1)
	rl.PlayAudioStream(stream)

	return &AudioOut{
		buffer:   make([]float32, bufferSize),
		channels: channels,
		stream:   stream,
	}
}

func (s *AudioOut) SetChannel(ch <-chan float32) {
	s.source = ch
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

func (s *AudioOut) drainSourceCh(out []float32) (n int) {
loop:
	for i := 0; i < len(out); i += s.channels {
		select {
		case sample := <-s.source:
			for j := 0; j < s.channels; j++ {
				out[i+j] = sample
			}

			n += s.channels
		default:
			log.Printf("[WARN] audio buffer underrun, want:%d, got:%d", len(out), n)
			break loop
		}
	}

	return n
}

func (s *AudioOut) Update() {
	if rl.IsAudioStreamProcessed(s.stream) {
		if f := s.drainSourceCh(s.buffer); f > 0 {
			rl.UpdateAudioStream(s.stream, s.buffer[:f])
		}
	}
}
