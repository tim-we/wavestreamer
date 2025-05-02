package utils

// SoftLimitGain applies a gain to an audio sample `x`.
// For gains > 1 a soft limiting function is applied to avoid sharp clipping.
//
// Parameters:
//   - x: The input audio sample in the range [-1, 1].
//   - gain: The gain applied to the signal. Must be 0 ≤ gain ≤ 2.
func SoftLimitGain(x, gain float32) float32 {
	if gain < 1 {
		return gain * x
	}

	absX := abs(x)
	var sign float32 = 1
	if x < 0 {
		sign = -1
	}

	// Blend between the identity function and f(x) = x*(2-x), which has the following properties:
	// 	f(x) ∈ [0,1] for x ∈ [0,1]
	// 	f(0) = 0 and f(1) = 1
	//  f(x) > x for x ∈ [0,1] (amplification)
	// 	f is a strictly monotonic function
	//  f'(1) = 0 (soft top)
	f := absX * (2 - absX)

	// The blending factor will be in [0,1].
	blend := min(2.0, gain) - 1

	// The blended function g will have the following property: g'(0) = gain
	return sign * Lerp(absX, f, blend)
}

type Float interface {
	~float32 | ~float64
}

// Clamp returns value clamped between minimum and maximum.
func Clamp[T Float](minimum, value, maximum T) T {
	if value < minimum {
		return minimum
	}

	if value > maximum {
		return maximum
	}

	return value
}

// Lerp performs linear interpolation between a and b using parameter s in [0, 1].
func Lerp[T Float](a, b, s T) T {
	d := b - a
	return a + s*d
}

func abs[T Float](x T) T {
	if x < 0 {
		return -x
	}
	return x
}
