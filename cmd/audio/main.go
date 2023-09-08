package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	maxSamples          = 22050
	maxSamplesPerUpdate = 4096
)

func main() {
	rl.InitWindow(800, 450, "raylib [audio] example - raw audio streaming")
	rl.SetAudioStreamBufferSizeDefault(maxSamplesPerUpdate)
	rl.InitAudioDevice()

	stream := rl.LoadAudioStream(22050, 32, 1)
	rl.SetAudioStreamVolume(stream, 0.2)
	rl.PlayAudioStream(stream)
	rl.SetTargetFPS(30)

	for !rl.WindowShouldClose() {
		if rl.IsAudioStreamProcessed(stream) {
			buf := make([]float32, 100)
			for i := range buf {
				buf[i] = 100
			}

			rl.UpdateAudioStream(stream, buf)
		}
	}

	rl.UnloadAudioStream(stream)
	rl.CloseAudioDevice()
	rl.CloseWindow()
}
