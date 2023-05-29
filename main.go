package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/rakyll/portmidi"
)

func main() {
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

	// Open the MIDI input stream for the specified device
	stream, err := portmidi.NewInputStream(portmidi.DeviceID(deviceID), 1024)
	if err != nil {
		log.Fatalf("Failed to open MIDI input stream: %v", err)
	}
	defer stream.Close()

	// Set up signal handling to gracefully exit the program
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("Exiting...")
		os.Exit(0)
	}()

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
			fmt.Printf("Received MIDI event: Status=%d, Data1=%d, Data2=%d\n",
				event.Status, event.Data1, event.Data2)

			// Process MIDI event
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
