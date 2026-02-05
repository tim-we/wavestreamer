package config

const SAMPLE_RATE = 44100

// The size of an AudioChunk. A frame consists of a sample for each channel.
const FRAMES_PER_BUFFER = 1024

// Unfortunately the number of channels cannot be changed (easily),
// as it requires changes in other places (for example AudioChunk) as well.
const CHANNELS = 2

// TARGET_MIN_RMS defines the minimum acceptable RMS (Root Mean Square) level for audio signals.
// If the signal's RMS falls below this threshold, automatic amplification will be applied
// to bring it closer to a consistent perceived loudness. Increase this value to make quiet
// signals louder; decrease it to allow lower-volume content without amplification.
const TARGET_MIN_RMS = 0.25

// When the signal is low and normalization is enabled (default) we will boost it
// by basically scaling it with a factor (its a bit more complicated but thats the rough idea).
// The maximum of this factor can be controlled here.
const MAX_AMPLIFICATION = 1.82
