package main

const minMouseMovePixels = 4

func significantMouseMovement(distance int32) bool {
	return distance >= minMouseMovePixels
}
