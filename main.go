// main GyroLock app
package main

import (
	"fmt"
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
	if !isRoot() {
		log.Println("It's recommanded to run it as root, if you run it as user will be easy to disable !")
	}
	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	sensitivity, err := strconv.ParseInt(os.Getenv("SENSITIVITY"), 0, 16)
	if err != nil || sensitivity < 0 {
		sensitivity = 10
	}
	log.Printf("GyroLock start with sensitivity = %d", sensitivity)
	prev := NewSensor(debug)
	cur := NewSensor(debug)
	for {
		cur.Get()
		if CheckShake(sensitivity, cur, prev, debug) {
			LockSessions(debug, "1")
			if !debug {
				time.Sleep(60 * time.Second)
			} else {
				time.Sleep(5 * time.Second)
			}
		}
		prev.Get()
		time.Sleep(200 * time.Millisecond)
	}
}

// CheckShake check if sensor was shake
func CheckShake(sensitivity int64, cur *Sensor, prev *Sensor, debug bool) bool {
	axis := []string{"x", "y", "z"}
	shake := false

	for _, v := range axis {
		ret := int64(math.Abs(cur.axis[v] - prev.axis[v]))
		if debug {
			log.Printf("diff: %v:%v", v, ret)
		}
		shake = ret > sensitivity
		if shake {
			log.Printf("GyroLock, shake detected: %v:%v", v, ret)
			break
		}
	}
	return shake
}

// NewSensor create a new sensor
func NewSensor(debug bool) *Sensor {
	axis := make(map[string]float64)
	s := Sensor{debug: debug, axis: axis}
	s.ReadSensorScale()
	s.Get()
	return &s
}

// Get the values of the sensor
func (s *Sensor) Get() {
	axis := []string{"x", "y", "z"}
	for _, v := range axis {
		ret := s.ReadSensor(v)
		if s.debug {
			log.Printf("current: %s:%v", v, ret)
		}
	}
}

// ReadSensorScale read sensor scale value
func (s *Sensor) ReadSensorScale() {
	fp, err := filepath.Glob("/sys/bus/iio/devices/iio:device*/in_accel_scale")
	if err != nil || len(fp) == 0 {
		log.Fatal("Can't get file in_accel_scale")
	}
	content, err := os.ReadFile(fp[0])
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
		if err != nil || len(fp) == 0 {
			log.Fatalf("Can't get file in_accel_%s_raw", axis)
		}
		content, err := os.ReadFile(fp[0])
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
			log.Printf("GyroLock lock session %v !", seat)
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
