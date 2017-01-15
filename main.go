package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/rakyll/portmidi"
)

func listDevices() (devs []*portmidi.DeviceInfo) {
	numDevices := portmidi.CountDevices()
	for i := 0; i < numDevices; i++ {
		info := portmidi.Info(portmidi.DeviceID(i))
		devs = append(devs, info)
	}
	return
}

func getInputDevices() (devs []*portmidi.DeviceInfo) {
	for _, d := range listDevices() {
		if d.IsInputAvailable {
			devs = append(devs, d)
		}
	}
	return
}

type midiMessage struct {
	status, data1, data2 int64
}

var commandMap = map[midiMessage]*exec.Cmd{
	midiMessage{
		0xb0,
		0x73,
		0x7f,
	}: exec.Command("osascript", "-e", "tell application \"Spotify\" to playpause"),
}

func handleEvent(e portmidi.Event) {
	if cmd, ok := commandMap[midiMessage{e.Status, e.Data1, e.Data2}]; ok {
		cmd.Run()
	}
}

func main() {
	if err := portmidi.Initialize(); err != nil {
		log.Fatal(err)
	}
	defer portmidi.Terminate()
	devs := getInputDevices()
	if len(devs) == 0 {
		fmt.Fprintln(os.Stderr, "No input devices found")
		os.Exit(1)
	}
	flag.Parse()
	inStream, err := portmidi.NewInputStream(portmidi.DefaultInputDeviceID(), 1024)
	if err != nil {
		log.Fatal(err)
	}
	defer inStream.Close()
	inCh := inStream.Listen()
	for {
		event := <-inCh
		fmt.Fprintf(os.Stderr, "Event! %x %x %x\n", event.Status, event.Data1, event.Data2)
		handleEvent(event)
	}

}
