package testutil

import "testing"

func Panic(t *testing.T, f func()) {
	defer func() {
		_ = recover()
	}()

	f() // jumps to defer above if panics
	t.Fatalf("should have panicked")
}

func Equal[T comparable](t *testing.T, got, want T) {
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}
