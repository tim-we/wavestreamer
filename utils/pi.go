package utils

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

// Thresholds for CPU reduction
const defaultTempHigh = 70.0 // °C, start reducing CPU

// ShouldReduceCPU checks the current CPU temperature and throttle state.
// Returns true if the system is currently hot or throttled and you should
// reduce CPU usage (e.g., lower ffmpeg threads).
func ShouldReduceCPU() bool {
	if !isRaspberryPi() {
		// non-Pi devices: never reduce CPU
		return false
	}

	temp, err := readTemp()
	if err != nil {
		log.Println("Warning: could not read CPU temp:", err)
		return false // fail-safe: don’t reduce CPU unnecessarily
	}

	throttle, err := readThrottle()
	if err != nil {
		log.Println("Warning: could not read throttle state:", err)
		return false
	}

	// Check if any current throttle flags are set
	const (
		UnderVoltageNow = 1 << 0
		ThrottledNow    = 1 << 2
		TempLimitNow    = 1 << 3
	)

	throttledNow := (throttle & (UnderVoltageNow | ThrottledNow | TempLimitNow)) != 0

	// Reduce CPU if either temperature is too high or throttling is active
	if temp >= defaultTempHigh || throttledNow {
		return true
	}

	// Otherwise safe to run at normal CPU usage
	return false
}

var (
	isPiOnce sync.Once
	isPi     bool
)

// isRaspberryPi detects if the current device is a Raspberry Pi.
// Checks /proc/cpuinfo for BCM or Raspberry Pi identifiers.
func isRaspberryPi() bool {
	isPiOnce.Do(func() {
		file, err := os.Open("/proc/cpuinfo")
		if err != nil {
			isPi = false
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "Hardware") && strings.Contains(line, "BCM") {
				isPi = true
				return
			}
			if strings.HasPrefix(line, "model name") && strings.Contains(line, "Raspberry Pi") {
				isPi = true
				return
			}
		}
		isPi = false
	})
	return isPi
}

// readTemp reads the CPU temperature from the Linux thermal sysfs interface.
// On Raspberry Pi, the value is exposed in millidegrees Celsius.
// The returned value is converted to degrees Celsius as a float64.
func readTemp() (float64, error) {
	// Read the raw temperature value from sysfs
	data, err := os.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	if err != nil {
		return 0, err
	}

	// The file contains a single integer value as text, e.g. "48234\n"
	raw := strings.TrimSpace(string(data))
	milliC, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}

	// Convert millidegrees Celsius to degrees Celsius
	return float64(milliC) / 1000.0, nil
}

// readThrottle queries the Raspberry Pi firmware for throttling and
// undervoltage status using vcgencmd.
//
// The returned value is a bitmask where individual bits indicate
// current or past power and thermal throttling conditions.
func readThrottle() (uint32, error) {
	// vcgencmd communicates with the VideoCore firmware
	cmd := exec.Command("vcgencmd", "get_throttled")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return 0, err
	}

	// Expected output format:
	// throttled=0x80000
	parts := strings.Split(strings.TrimSpace(out.String()), "=")

	// Parse the hexadecimal bitmask value
	val, err := strconv.ParseUint(parts[1], 0, 32)
	if err != nil {
		return 0, err
	}

	return uint32(val), nil
}
