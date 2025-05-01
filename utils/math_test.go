package utils

import (
	"math"
	"math/rand"
	"testing"
)

func TestSoftLimitNoGain(t *testing.T) {
	gain := float32(1.0)
	xThreshold, alpha := SoftLimitParameters(gain)

	for i := range 100 {
		x := float32(i) / 100

		y := SoftLimit(x, xThreshold, gain, alpha)

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
		xThreshold, alpha := SoftLimitParameters(gain)

		var lastY float32 = -0.1337

		for i := range 100 {
			x := float32(i) / 100
			y := SoftLimit(x, xThreshold, gain, alpha)

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

	for range 42 {
		gain := rng.Float32() + 0.5
		if rng.Intn(13) < 2 {
			gain = 1.0
		}
		xThreshold, alpha := SoftLimitParameters(gain)

		if gain > 1 && alpha > 0 {
			t.Errorf("Expected to alpha to be negative for gain %v but was: %v", gain, alpha)
		}

		for i := range 100 {
			x := float32(i) / 100
			y := SoftLimit(x, xThreshold, gain, alpha)

			if y < 0 || y > 1 {
				t.Errorf("Incorrect output range (got %v)", y)
			}

			if math.IsNaN(float64(y)) {
				t.Errorf("Unexpected NaN for gain=%v and x=%v", gain, x)
			}
		}
	}
}

func isApproxEqual(a, b, delta float32) bool {
	d := f32abs(a - b)
	return d <= delta
}
