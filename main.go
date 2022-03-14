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
	X     float64
	Y     float64
	Z     float64
	prevX float64
	prevY float64
	prevZ float64
	debug bool
}

func main() {
	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	sensitivity, err := strconv.ParseInt(os.Getenv("SENSITIVITY"), 0, 16)
	if err != nil || sensitivity < 0 {
		sensitivity = 5
	}
	log.Printf("GyroLock start with sensitivy = %d", sensitivity)
	s := New(debug)
	for {
		s.Calibrate()
		time.Sleep(200 * time.Millisecond)
		s.Get()
		if s.CheckShake(sensitivity) {
			LockSessions(debug)
			if !debug {
				time.Sleep(60 * time.Second)
			} else {
				time.Sleep(5 * time.Second)
			}
		}
	}
}

// New create a new sensor
func New(debug bool) *Sensor {
	s := Sensor{debug: debug}
	return &s
}

// Get the values of the sensor
func (s *Sensor) Get() {
	s.X = ReadSensor("x")
	s.Y = ReadSensor("y")
	s.Z = ReadSensor("z")
	if s.debug {
		log.Printf("current: x:%v y:%v z:%v\n", s.X, s.Y, s.Z)
	}
}

// Calibrate the sensor
func (s *Sensor) Calibrate() {
	s.prevX = ReadSensor("x")
	s.prevY = ReadSensor("y")
	s.prevZ = ReadSensor("z")
	if s.debug {
		log.Printf("previous: x:%v y:%v z:%v\n", s.prevX, s.prevY, s.prevZ)
	}
}

// CheckShake check if sensor was shake
func (s *Sensor) CheckShake(sensitivity int64) bool {
	diffX := int64(math.Abs(s.X - s.prevX))
	diffY := int64(math.Abs(s.Y - s.prevY))
	diffZ := int64(math.Abs(s.Z - s.prevZ))
	shake := diffX > sensitivity || diffY > sensitivity || diffZ > sensitivity
	if s.debug {
		log.Printf("diff: diffX:%v, diffY:%v, diffZ:%v", diffX, diffY, diffZ)
	}
	if shake {
		log.Printf("GyroLock, shake detected: x:%v, y:%v, z:%v", diffX, diffY, diffZ)
	}
	return shake
}

// ReadSensor read sensor value as absolute value
func ReadSensor(axis string) float64 {
	var value int64
	var scale float64
	for {
		fp, err := filepath.Glob(
			fmt.Sprintf("/sys/bus/iio/devices/iio:device*/in_accel_%s_raw", axis),
		)
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
		time.Sleep(15 * time.Millisecond)
	}
	return float64(math.Abs(float64(value)) * scale)
}

// LockSessions lock the current session
func LockSessions(debug bool) {
	conn, err := login1.New()
	if err != nil {
		os.Exit(1)
	}
	if !debug {
		conn.LockSessions()
		log.Println("GyroLock lock seesions !")
	} else {
		log.Println("GyroLock would lock seesions !")
	}
}
