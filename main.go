package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-systemd/login1"
)

func main() {

	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	sensitivity, err := strconv.ParseInt(os.Getenv("SENSITIVITY"), 0, 16)
	if err != nil {
		sensitivity = 100
	}
	if debug {
		log.Printf("SENSITIVITY: %d", sensitivity)
	}

	initX := ReadSensor("x")
	initY := ReadSensor("y")
	initZ := ReadSensor("z")

	if debug {
		log.Printf("x:%v y:%v z:%v\n", initX, initY, initZ)
	}

	for {

		x := ReadSensor("x")
		y := ReadSensor("y")
		z := ReadSensor("z")

		if debug {
			log.Printf("x:%v y:%v z:%v\n", x, y, z)
		}

		if (initX+(initX*sensitivity)/100) < x && (initY+(initY*sensitivity)/100) < y || (initZ+(initZ*sensitivity)/100) < z {
			LockSession(debug)
			initX = x
			initY = y
			initZ = z
			time.Sleep(15 * time.Second)
		} else {
			time.Sleep(2 * time.Second)
		}
	}

}

// ReadSensor read sensor value as absolute value
func ReadSensor(axis string) int64 {
	var value int64
	for {
		content, err := ioutil.ReadFile(fmt.Sprintf("/sys/bus/iio/devices/iio:device1/in_accel_%s_raw", axis))
		if err != nil {
			log.Fatalf("Can't read sensor %s\n", axis)
		}
		value, _ = strconv.ParseInt(strings.TrimSpace(string(content)), 0, 64)
		if value != 0 {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	return int64(math.Abs(float64(value)))
}

// LockSession lock the current session
func LockSession(debug bool) {
	conn, err := login1.New()
	if err != nil {
		os.Exit(1)
	}
	if !debug {
		conn.LockSession("1")
	}
	log.Println("Lock !")
}
