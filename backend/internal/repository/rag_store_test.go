package repository

import "testing"

func TestVectorLiteral(t *testing.T) {
	got := vectorLiteral([]float32{0.1, -2.5, 3})
	want := "[0.1,-2.5,3]"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
