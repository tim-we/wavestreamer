package utils

// SoftLimit applies a soft limiting function to an audio sample `x`.
//
// This function scales the input linearly by `gain` as long as the absolute
// value of the sample is below the `xThreshold`. Beyond this threshold, it
// smoothly bends the output using a quadratic curve to avoid hard clipping,
// ensuring a continuous and differentiable transition.
//
// Parameters:
//   - x: The input audio sample (typically in the range [-1, 1]).
//   - xThreshold: The input threshold beyond which the soft limiting begins.
//     This is usually derived from a desired output threshold.
//   - gain: The linear gain applied below the threshold. Must be ≥ 1.
//   - alpha: A precomputed shaping parameter that controls the curvature
//     of the limiting section. It ensures that SoftLimit(1, ...) == 1
//     and that the function's derivative at x == 1 is zero.
//
// Optimal values for xThreshold and alpha can be computed with SoftLimitParameters.
//
// Returns:
//   - A softly limited version of `x`, preserving signal dynamics below
//     the threshold and avoiding sharp clipping above it.
func SoftLimit(x, xThreshold, gain, alpha float32) float32 {
	if gain < 1 {
		return gain * x
	}

	absX := f32abs(x)
	if absX < xThreshold {
		return gain * x
	}

	var sign float32 = 1.0
	if x < 0 {
		sign = -1.0
	}

	y := x - 1
	y = alpha*y*y + 1 // alpha * (x-1)² + 1

	return sign * y
}

// SoftLimitParameters computes the xThreshold and alpha parameters for SoftLimit.
func SoftLimitParameters(gain float32) (float32, float32) {
	if gain <= 1 {
		// We only support gain >= 1, for gain < 1 no soft limiting is required.
		// xThreshold = 1 means SoftLimit becomes the identity function.
		// alpha = 0 means the constant 1 function but is not relevant here.
		return 1, 0
	}

	// Compute xThreshold s.t. there is smooth transition from the linear to quadratic piece
	var xThreshold float32 = 1 / (2*gain - 1)
	var yThreshold float32 = gain * xThreshold

	if yThreshold < 0.5 {
		// If the threshold is too low we ignore the smoothness requirement
		yThreshold = 0.5
		xThreshold = yThreshold / gain
	}

	// Compute alpha = (yThreshold - 1) / (xThreshold -1)²
	// SoftLimit(x) = alpha * (x-1)² + 1 for x >= xThreshold
	// The formula has been derived from
	//  	SoftLimit(xThreshold) = yThreshold
	//		SoftLimit(1) = 1
	//		SoftLimit'(1) = 0
	tmp := 1 - xThreshold
	var alpha float32 = (yThreshold - 1) / max(0.001, tmp*tmp)

	return xThreshold, alpha
}

func f32abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
