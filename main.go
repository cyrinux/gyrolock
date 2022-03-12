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

// Sensor is a iio accelerometer struct
type Sensor struct {
	InitX int64
	InitY int64
	InitZ int64
	X     int64
	Y     int64
	Z     int64
}

func main() {
	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	sensitivity, err := strconv.ParseInt(os.Getenv("SENSITIVITY"), 0, 16)
	if err != nil || sensitivity < 0 || sensitivity > 200 {
		sensitivity = 100
	}
	if debug {
		log.Printf("SENSITIVITY: %d", sensitivity)
	}
	s := New()
	s.Calibrate(debug)
	for {
		s.Get(debug)
		if s.CheckShake(sensitivity) {
			LockSession(debug)
			time.Sleep(15 * time.Second)
			s.Calibrate(debug)
		} else {
			time.Sleep(1 * time.Second)
		}
	}
}

// New create a new sensor
func New() *Sensor {
	s := Sensor{}
	return &s
}

// Get the values of the sensor
func (s *Sensor) Get(debug bool) {
	s.X = ReadSensor(1, "x")
	s.Y = ReadSensor(1, "y")
	s.Z = ReadSensor(1, "z")
	if debug {
		log.Printf("x:%v y:%v z:%v\n", s.X, s.Y, s.Z)
	}
}

// Calibrate the sensor
func (s *Sensor) Calibrate(debug bool) {
	s.InitX = ReadSensor(1, "x")
	s.InitY = ReadSensor(1, "y")
	s.InitZ = ReadSensor(1, "z")
	if debug {
		log.Printf("init: x:%v y:%v z:%v\n", s.InitX, s.InitY, s.InitZ)
	}
}

// CheckShake check if sensor was shake
func (s *Sensor) CheckShake(sensitivity int64) bool {
	return (s.InitX+(s.InitX*sensitivity)/100) < s.X ||
		(s.InitY+(s.InitY*sensitivity)/100) < s.Y || (s.InitZ+(s.InitZ*sensitivity)/100) < s.Z
}

// ReadSensor read sensor value as absolute value
func ReadSensor(device int, axis string) int64 {
	var value int64
	var scale float64
	for {
		content, err := ioutil.ReadFile(fmt.Sprintf("/sys/bus/iio/devices/iio:device%d/in_accel_%s_raw", device, axis))
		if err != nil {
			log.Fatalf("Can't read sensor %s", axis)
		}
		value, _ = strconv.ParseInt(strings.TrimSpace(string(content)), 0, 64)
		content, err = ioutil.ReadFile(fmt.Sprintf("/sys/bus/iio/devices/iio:device%d/in_accel_scale", device))
		if err != nil {
			log.Fatalf("Can't read sensor")
		}
		scale, _ = strconv.ParseFloat(strings.TrimSpace(string(content)), 64)
		if value != 0 {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	return int64(math.Abs(float64(value)) * scale)
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
