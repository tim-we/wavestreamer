package gpio

import (
	"log"
	"time"

	"github.com/tim-we/wavestreamer/player"
	"github.com/tim-we/wavestreamer/player/clips"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

// Wiring instructions for QIACHIP RX480e to Raspberry Pi 3B+:
//
// RX480e Module -> Raspberry Pi
// --------------------------------
// GND           -> Pin 6  (GND) - or any other ground pin
// +V            -> Pin 2  (5V)  - powers the receiver module (3.3V should also work)
// D2            -> Pin 11 (GPIO17) - data signal from receiver
//
// The 433MHz transmitter sends repeated pulses when button is held.
// We detect "button pressed" by seeing pulses, and "released" when pulses stop.

const (
	// If no pulse detected for this long, consider button released
	pulseTimeout = 15 * time.Millisecond

	// If button released within this time, cancel the pause (just skip)
	longPressThreshold = 1 * time.Second
)

type buttonEvent int

const (
	buttonPulse buttonEvent = iota
	buttonReleased
)

func InitGPIOButton(pinName string) {
	// Initialize periph.io
	if _, err := host.Init(); err != nil {
		log.Fatal("Failed to initialize periph.io:", err)
	}

	if pinName == "" {
		pinName = "GPIO17"
	}

	// Get pin (by default GPIO17)
	pin := gpioreg.ByName(pinName)
	if pin == nil {
		log.Fatalf("Failed to find pin '%s'", pinName)
	}

	// Configure as input with pull-up, detect rising edges only
	if err := pin.In(gpio.PullUp, gpio.RisingEdge); err != nil {
		log.Fatal("Failed to configure pin:", err)
	}

	// Channel for thread-safe communication between goroutines
	events := make(chan buttonEvent, 10)

	// Goroutine 1: Detect GPIO pulses
	go func() {
		for {
			// Indefinitely wait for a rising edge
			pin.WaitForEdge(-1)

			// Verify it's actually HIGH (filter out noise/glitches)
			if pin.Read() != gpio.High {
				continue
			}

			// Send pulse event
			events <- buttonPulse
		}
	}()

	// Goroutine 2: Process button events with state management
	go func() {
		buttonPressed := false
		var releaseTimer *time.Timer
		var pressStartTime time.Time

		for event := range events {
			switch event {
			case buttonPulse:
				if buttonPressed {
					// Button already pressed - just reset the timer
					if releaseTimer != nil {
						releaseTimer.Reset(pulseTimeout)
					}
				} else {
					// First pulse detected - button just pressed
					buttonPressed = true
					pressStartTime = time.Now()
					log.Printf("[GPIO] Button %s pressed", pinName)

					// Queue Pause and skip current song (= start pause)
					player.QueueClipNext(clips.NewPause(10 * time.Minute))
					player.SkipCurrent()

					// Start the release timer
					releaseTimer = time.AfterFunc(pulseTimeout, func() {
						events <- buttonReleased
					})
				}

			case buttonReleased:
				if !buttonPressed {
					// Ignore spurious release events
					continue
				}

				// No pulses received for pulseTimeout milliseconds.
				// We assume this means the button was released.
				buttonPressed = false
				pressDuration := time.Since(pressStartTime)

				log.Printf("[GPIO] Button %s released (held for %v)", pinName, pressDuration)

				// If released quickly (< 1 second), cancel the pause by skipping it
				if pressDuration < longPressThreshold {
					log.Printf("[GPIO] Quick release detected - canceling pause")
					player.SkipCurrent()
				}
			}
		}
	}()
}
