package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
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
	s.X = ReadSensor("x")
	s.Y = ReadSensor("y")
	s.Z = ReadSensor("z")
	if debug {
		log.Printf("x:%v y:%v z:%v\n", s.X, s.Y, s.Z)
	}
}

// Calibrate the sensor
func (s *Sensor) Calibrate(debug bool) {
	s.InitX = ReadSensor("x")
	s.InitY = ReadSensor("y")
	s.InitZ = ReadSensor("z")
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
func ReadSensor(axis string) int64 {
	var value int64
	var scale float64
	for {
		fp, err := filepath.Glob(fmt.Sprintf("/sys/bus/iio/devices/iio:device*/in_accel_%s_raw", axis))
		if err != nil {
			log.Fatal("Can't get file")
		}
		content, err := ioutil.ReadFile(fp[0])
		if err != nil {
			log.Fatalf("Can't read sensor value of axis %s", axis)
		}
		value, _ = strconv.ParseInt(strings.TrimSpace(string(content)), 0, 64)
		fp, err = filepath.Glob("/sys/bus/iio/devices/iio:device*/in_accel_scale")
		if err != nil {
			log.Fatal("Can't get file")
		}
		content, err = ioutil.ReadFile(fp[0])
		if err != nil {
			log.Fatalf("Can't read sensor scale value")
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
		conn.LockSessions()
	}
	log.Println("Lock !")
}
