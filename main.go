package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-systemd/login1"
)

// Sensor is a IIO accelerometer struct
type Sensor struct {
	axis  map[string]float64
	scale float64
	debug bool
}

func main() {
	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	sensitivity, err := strconv.ParseInt(os.Getenv("SENSITIVITY"), 0, 16)
	if err != nil || sensitivity < 0 {
		sensitivity = 10
	}
	log.Printf("GyroLock start with sensitivity = %d", sensitivity)
	prev := New(debug)
	cur := New(debug)
	for {
		prev.Get()
		time.Sleep(200 * time.Millisecond)
		cur.Get()
		if CheckShake(sensitivity, cur, prev, debug) {
			LockSessions(debug, "1")
			if !debug {
				time.Sleep(60 * time.Second)
			} else {
				time.Sleep(5 * time.Second)
			}
		}
	}
}

// CheckShake check if sensor was shake
func CheckShake(sensitivity int64, cur *Sensor, prev *Sensor, debug bool) bool {
	diffX := int64(math.Abs(cur.axis["x"] - prev.axis["x"]))
	diffY := int64(math.Abs(cur.axis["y"] - prev.axis["y"]))
	diffZ := int64(math.Abs(cur.axis["z"] - prev.axis["z"]))
	shake := diffX > sensitivity || diffY > sensitivity || diffZ > sensitivity
	if debug {
		log.Printf("diff: diffX:%v, diffY:%v, diffZ:%v", diffX, diffY, diffZ)
	}
	if shake {
		log.Printf("GyroLock, shake detected: x:%v, y:%v, z:%v", diffX, diffY, diffZ)
	}
	return shake
}

// New create a new sensor
func New(debug bool) *Sensor {
	axis := make(map[string]float64)
	s := Sensor{debug: debug, axis: axis}
	s.ReadSensorScale()
	s.Get()
	return &s
}

// Get the values of the sensor
func (s *Sensor) Get() {
	x := s.ReadSensor("x")
	y := s.ReadSensor("y")
	z := s.ReadSensor("z")
	if s.debug {
		log.Printf("current: x:%v y:%v z:%v", x, y, z)
	}
}

// ReadSensorScale read sensor scale value
func (s *Sensor) ReadSensorScale() {
	fp, err := filepath.Glob("/sys/bus/iio/devices/iio:device*/in_accel_scale")
	if err != nil {
		log.Fatal("Can't get file in_accel_scale")
	}
	content, err := ioutil.ReadFile(fp[0])
	if err != nil {
		log.Fatalf("Can't read sensor scale value from in_accel_scale")
	}
	scale, err := strconv.ParseFloat(strings.TrimSpace(string(content)), 64)
	if err != nil {
		scale = 1
	}
	s.scale = scale
}

// ReadSensor read sensor value as absolute value
func (s *Sensor) ReadSensor(axis string) float64 {
	var value int64
	for {
		fp, err := filepath.Glob(
			fmt.Sprintf("/sys/bus/iio/devices/iio:device*/in_accel_%s_raw", axis),
		)
		if err != nil {
			log.Fatalf("Can't get file in_accel_%s_raw", axis)
		}
		content, err := ioutil.ReadFile(fp[0])
		if err != nil {
			log.Fatalf("Can't read sensor value from in_accel_%s_raw", axis)
		}
		value, err = strconv.ParseInt(strings.TrimSpace(string(content)), 0, 64)
		if err == nil && value != 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	result := float64(math.Abs(float64(value)) * s.scale)
	s.axis[axis] = result
	return result
}

// LockSessions lock the current session
func LockSessions(debug bool, seat string) {
	conn, err := login1.New()
	if err != nil {
		os.Exit(1)
	}
	if !debug {
		if isRoot() {
			conn.LockSessions()
			log.Println("GyroLock lock sessions !")
		} else {
			conn.LockSession(seat)
			log.Println("GyroLock lock session %v !", seat)
		}
	} else {
		log.Println("GyroLock would lock sessions !")
	}
}

func isRoot() bool {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("Unable to get current user: %s", err)
	}
	return currentUser.Username == "root"
}
