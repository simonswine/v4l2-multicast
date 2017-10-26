// Example program that uses blakjack/webcam library
// for working with V4L2 devices.
// The application reads frames from device and writes them to stdout
// If your device supports motion formats (e.g. H264 or MJPEG) you can
// use it's output as a video stream.
// Example usage: go run stdout_streamer.go | vlc -
package main

import (
	"fmt"
	"os"

	"github.com/blackjack/webcam"
	log "github.com/sirupsen/logrus"
)

func readChoice(s string) int {
	var i int
	for true {
		print(s)
		_, err := fmt.Scanf("%d\n", &i)
		if err != nil || i < 1 {
			println("Invalid input. Try again")
		} else {
			break
		}
	}
	return i
}

type FrameSizes []webcam.FrameSize

func (slice FrameSizes) Len() int {
	return len(slice)
}

//For sorting purposes
func (slice FrameSizes) Less(i, j int) bool {
	ls := slice[i].MaxWidth * slice[i].MaxHeight
	rs := slice[j].MaxWidth * slice[j].MaxHeight
	return ls < rs
}

//For sorting purposes
func (slice FrameSizes) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func main() {
	log.SetLevel(log.DebugLevel)

	cam, err := webcam.Open("/dev/video0")
	if err != nil {
		log.Fatalf("error opening camera: %s", err.Error())
	}
	defer cam.Close()

	format_desc := cam.GetSupportedFormats()
	var format webcam.PixelFormat
	for key, val := range format_desc {
		log.Debugf("supported format: %s", val)
		if val == "H.264" {
			format = key
			break
		}
	}

	if format == 0 {
		log.Fatalf("H.264 not supported")
	}

	f, w, h, err := cam.SetImageFormat(format, uint32(1280), uint32(720))
	if err != nil {
		log.Fatalf("error setting image format: %s", err.Error())
	} else {
		log.Infof("resulting image format: %s (%dx%d)\n", format_desc[f], w, h)
	}

	println("Press Enter to start streaming")
	fmt.Scanf("\n")
	err = cam.StartStreaming()
	if err != nil {
		panic(err.Error())
	}

	timeout := uint32(5) //5 seconds
	for {
		err = cam.WaitForFrame(timeout)

		switch err.(type) {
		case nil:
		case *webcam.Timeout:
			fmt.Fprint(os.Stderr, err.Error())
			continue
		default:
			panic(err.Error())
		}

		frame, err := cam.ReadFrame()
		if len(frame) != 0 {
			print(".")
			os.Stdout.Write(frame)
			os.Stdout.Sync()
		} else if err != nil {
			panic(err.Error())
		}
	}
}
