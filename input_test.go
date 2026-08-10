package main

import "testing"

func TestSignificantMouseMovementRequiresOneMeaningfulDelta(t *testing.T) {
	if significantMouseMovement(3) {
		t.Fatal("tiny mouse movement was treated as significant")
	}
	if !significantMouseMovement(4) {
		t.Fatal("boundary mouse movement was ignored")
	}
}
