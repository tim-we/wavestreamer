package gpio

import (
	"log"
	"sync"
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

type ButtonEvent int

const (
	ButtonPressed ButtonEvent = iota
	ButtonReleased
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

	events := make(chan ButtonEvent, 10)

	// Goroutine 1: Convert GPIO pulses to Button events (pressed/released)
	go func() {
		var releaseTimer *time.Timer
		var timerMutex sync.Mutex

		if pin.Read() == gpio.High {
			log.Printf("[GPIO] Button %s already pressed at startup", pinName)
			log.Printf("[GPIO] Button might be stuck - ignoring all button input")
			return
		}

		for {
			// Indefinitely wait for a rising edge
			pin.WaitForEdge(-1)

			// Verify it's actually HIGH (filter out noise/glitches)
			if pin.Read() != gpio.High {
				continue
			}

			timerMutex.Lock()
			if releaseTimer != nil {
				releaseTimer.Reset(pulseTimeout)
				timerMutex.Unlock()

				// Avoid sending multiple pressed events
				continue
			}

			// We assume the button has been released after no high edges after pulseTimeout
			releaseTimer = time.AfterFunc(pulseTimeout, func() {
				events <- ButtonReleased

				timerMutex.Lock()
				releaseTimer = nil
				timerMutex.Unlock()
			})
			timerMutex.Unlock()

			events <- ButtonPressed
		}
	}()

	// Goroutine 2: Process button events with state management
	go func() {
		var pressStartTime time.Time
		var longPressTimer *time.Timer

		for event := range events {
			switch event {
			case ButtonPressed:
				log.Printf("[GPIO] Button %s pressed", pinName)
				pressStartTime = time.Now()

				// Schedule the long pause
				player.QueueClipNext(clips.NewPause(10 * time.Minute))
				player.PlayPriorityClip(clips.NewBeep())
				player.SkipCurrent()

				longPressTimer = time.AfterFunc(longPressThreshold, func() {
					// Indicate long press by playing a beep
					player.PlayPriorityClip(clips.NewBeep())
				})
			case ButtonReleased:
				if longPressTimer == nil {
					// This should never happen but we can handle it trivially
					break
				}
				longPressTimer.Stop()

				pressDuration := time.Since(pressStartTime)
				log.Printf("[GPIO] Button %s released (held for %v)", pinName, pressDuration)

				// Handle short press. Long presses are handled in the longPressTimer callback.
				if pressDuration < longPressThreshold {
					// Skip pause, user just wants to skip the current clip.
					log.Printf("[GPIO] Quick release detected - canceling pause")
				}
			}
		}
	}()
}
