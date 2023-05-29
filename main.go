package main

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/rakyll/portmidi"
)

func main() {
	// Terminate PortMidi to try and fix the replugging issue
	portmidi.Terminate()

	err := portmidi.Initialize()
	if err != nil {
		log.Fatal("Failed to initialize PortMidi:", err)
	}
	defer portmidi.Terminate()

	// Find the MIDI device with the name "MIX5R Pro"
	deviceID := findMIDIDevice("MIX5R Pro ")
	if deviceID == -1 {
		log.Fatal("MIDI device not found: MIX5R Pro")
	}

	// trying to fix the replug problem
	// Add a 100 millisecond delay
	time.Sleep(time.Millisecond * 100)

	// Open the MIDI input stream for the specified device
	stream, err := portmidi.NewInputStream(portmidi.DeviceID(deviceID), 1024)
	if err != nil {
		log.Fatalf("Failed to open MIDI input stream: %v", err)
	}
	defer stream.Close()

	fmt.Println("Listening to MIDI input...")
	for {
		// Read available MIDI events from the input stream
		events, err := stream.Read(1024)
		if err != nil {
			log.Printf("Failed to read MIDI events: %v", err)
			continue
		}

		// Process each MIDI event
		for _, event := range events {
			// Print the MIDI event information
			fmt.Println("Received MIDI event")

			// handle the speaker volume changes
			if event.Status == 176 && event.Data1 == 11 {
				fmt.Println(event.Data2)
				percentage := int64((float64(event.Data2) / 127) * 100) // 127 is max midi volume
				fmt.Println(percentage)
				if percentage > 100 {
					setVolume(100)
				} else {
					setVolume(percentage)
				}
			}

			// handle the mic gain changes
			if event.Status == 176 && event.Data1 == 1 {
				percentage := int64((float64(event.Data2) / 127) * 100) // 127 is max midi volume
				if percentage > 100 {
					setGain(100)
				} else {
					setGain(percentage)
				}
			}
		}

		// Sleep for a short duration to reduce CPU usage
		time.Sleep(10 * time.Millisecond)
	}
}

func findMIDIDevice(deviceName string) int {
	numDevices := portmidi.CountDevices()
	for i := 0; i < numDevices; i++ {
		info := portmidi.Info(portmidi.DeviceID(i))

		// escape early
		if info == nil {
			return -1
		}

		// make sure we have the correct midi device
		if info.Name == deviceName && info.IsInputAvailable {
			fmt.Println("Device found:", info.Name)
			return i
		}
	}
	return -1
}

func setVolume(volPercent int64) {
	// Execute the pactl command to set the volume
	cmd := exec.Command("pactl", "set-sink-volume", "@DEFAULT_SINK@", strings.Join([]string{strconv.Itoa(int(volPercent)), "%"}, ""))
	commandErr := cmd.Run()
	if commandErr != nil {
		log.Fatal("Failed to set speaker volume:", commandErr)
	}

	log.Println("Speaker volume set successfully.")
}

func setGain(gainPercent int64) {
	// Execute the pactl command to set the gain
	cmd := exec.Command("pactl", "set-source-volume", "@DEFAULT_SOURCE@", strings.Join([]string{strconv.Itoa(int(gainPercent)), "%"}, ""))
	commandErr := cmd.Run()
	if commandErr != nil {
		log.Fatal("Failed to set mic volume:", commandErr)
	}

	log.Println("Mic volume set successfully.")
}
