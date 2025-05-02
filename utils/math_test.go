package utils

import (
	"math"
	"math/rand"
	"testing"
)

func TestSoftLimitNoGain(t *testing.T) {
	gain := float32(1.0)

	for i := range 100 {
		x := float32(i) / 100

		y := SoftLimitGain(x, gain)

		if math.IsNaN(float64(y)) {
			t.Errorf("Unexpected NaN for x=%v", x)
		}

		if !isApproxEqual(x, y, 1e-6) {
			t.Errorf("Expected %v to be %v", y, x)
		}
	}
}

func TestSoftLimitMonotonicity(t *testing.T) {
	seed := int64(42)
	rng := rand.New(rand.NewSource(seed))

	for range 42 {
		gain := rng.Float32() + 0.5

		var lastY float32 = -0.1337

		for i := range 100 {
			x := float32(i) / 100
			y := SoftLimitGain(x, gain)

			if y <= lastY {
				t.Errorf("Expected SoftLimit to be strictly monotonic but its not for gain %v", gain)
			}

			lastY = y
		}
	}
}

func TestSoftLimitOutputRange(t *testing.T) {
	seed := int64(13)
	rng := rand.New(rand.NewSource(seed))

	gains := make([]float32, 42)

	for i := range len(gains) {
		gains[i] = rng.Float32() + 0.5
	}

	gains[0] = 1

	for _, gain := range gains {
		for i := range 100 {
			x := float32(i) / 100
			y := SoftLimitGain(x, gain)

			if y < 0 || y > 1 {
				t.Errorf("Incorrect output range (got %v)", y)
			}

			if math.IsNaN(float64(y)) {
				t.Errorf("Unexpected NaN for gain=%v and x=%v", gain, x)
			}
		}

		if !isApproxEqual(0, SoftLimitGain(0, gain), 1e-6) {
			t.Errorf("SoftLimitGain(0) = %v but should have been 0", SoftLimitGain(0, gain))
		}

		if gain > 1 && !isApproxEqual(1, SoftLimitGain(1, gain), 1e-6) {
			t.Errorf("SoftLimitGain(1) = %v but should have been 1", SoftLimitGain(1, gain))
		}
	}
}

func isApproxEqual(a, b, delta float32) bool {
	d := abs(a - b)
	return d <= delta
}
