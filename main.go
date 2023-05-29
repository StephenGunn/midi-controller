package main

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"github.com/rakyll/portmidi"
)

func main() {
	// MIDI initialization
	err := portmidi.Initialize()
	if err != nil {
		log.Fatal("Error initializing PortMidi: ", err)
	}
	defer portmidi.Terminate()

	// list the devices with "aplaymidi -l" to find the correct device ID
	deviceID := findInputDevice("MIX5R Pro ") // leave the space
	if deviceID == -1 {
		log.Fatal("MIDI device not found")
	}

	// connect to our target device
	input, err := portmidi.NewInputStream(portmidi.DeviceID(deviceID), 1024)
	if err != nil {
		log.Fatal("Error connecting to MIDI input device:", err)
	}
	defer input.Close()

	log.Println("Speaker volume set successfully.")

	// HANDLE MIDI INPUT
	fmt.Println("Reading MIDI input from device:", deviceID)
	ch := input.Listen()
	for {
		event := <-ch
		// Process MIDI event
		fmt.Println("Received MIDI event:", event)
		status := event.Status // check against value 176 for proper midi function
		channel := event.Data1
		value := event.Data2

		// handle the speaker volume changes
		if status == 176 && channel == 11 {
			if value > 100 {
				setVolume(100)
			} else {
				setVolume(value)
			}
		}

		// handle the mic gain changes
		if status == 176 && channel == 1 {
			if value > 100 {
				setGain(100)
			} else {
				setGain(value)
			}
		}
	}
}

func findInputDevice(deviceName string) int {
	numDevices := portmidi.CountDevices()
	fmt.Println("-- all available MIDI devices --")
	for i := 0; i < numDevices; i++ {
		// let's print out each midi device for fun
		info := portmidi.Info(portmidi.DeviceID(i))

		// newb logging

		fmt.Println("DEVICE:", info.Name, ", available:", info.IsInputAvailable)

		if info.IsInputAvailable && info.Name == deviceName {
			return i
		}
	}
	return -1
}

func setVolume(volPercent int64) {
	// execute the pactl command
	// Execute the pactl command to set the volume
	cmd := exec.Command("pactl", "set-sink-volume", "@DEFAULT_SINK@", strings.Join([]string{strconv.Itoa(int(volPercent)), "%"}, ""))
	commandErr := cmd.Run()
	if commandErr != nil {
		log.Fatal("Failed to set speaker volume:", commandErr)
	}

	log.Println("Speaker volume set successfully.")
}

func setGain(gainPercent int64) {
	// execute the pactl command
	// Execute the pactl command to set the volume
	cmd := exec.Command("pactl", "set-source-volume", "@DEFAULT_SOURCE@", strings.Join([]string{strconv.Itoa(int(gainPercent)), "%"}, ""))
	commandErr := cmd.Run()
	if commandErr != nil {
		log.Fatal("Failed to set mic volume:", commandErr)
	}

	log.Println("Mic volume set successfully.")
}
