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
	axisValues     map[string]float64
	prevAxisValues map[string]float64
	axisPaths      map[string]string
	scale          float64
	debug          bool
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
	s := NewSensor(debug)
	for {
		s.Get()
		if debug {
			log.Printf("current: %v", s.axisValues)
		}
		if s.CheckShake(sensitivity, debug) {
			LockSessions(debug, "1")
			if !debug {
				time.Sleep(60 * time.Second)
			} else {
				time.Sleep(5 * time.Second)
			}
		}

		s.savePrevious()
		if debug {
			log.Printf("previous: %v", s.prevAxisValues)
		}

		time.Sleep(200 * time.Millisecond)
	}
}

func (s *Sensor) savePrevious() {
	axis := []string{"x_raw", "y_raw", "z_raw"}
	for _, v := range axis {
		s.prevAxisValues[v] = s.axisValues[v]
	}
}

// CheckShake check if sensor was shake
func (s *Sensor) CheckShake(sensitivity int64, debug bool) bool {
	axis := []string{"x_raw", "y_raw", "z_raw"}
	shake := false

	for _, v := range axis {
		ret := int64(math.Abs(s.prevAxisValues[v] - s.axisValues[v]))
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
	axisValues := make(map[string]float64)
	prevAxisValues := make(map[string]float64)
	axisPaths := make(map[string]string)
	axis := []string{"x_raw", "y_raw", "z_raw", "scale"}
	for _, v := range axis {
		axisPaths[v] = getSensorPath(fmt.Sprintf("in_accel_%s", v))
	}
	s := Sensor{debug: debug, axisValues: axisValues, prevAxisValues: prevAxisValues, axisPaths: axisPaths}
	s.ReadSensorScale()
	return &s
}

// Get the values of the sensor
func (s *Sensor) Get() {
	axis := []string{"x_raw", "y_raw", "z_raw"}
	for _, v := range axis {
		s.ReadSensor(v)
	}
}

// ReadSensorScale read sensor scale value
func (s *Sensor) ReadSensorScale() {
	content, err := os.ReadFile(s.axisPaths["scale"])
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
func (s *Sensor) ReadSensor(axis string) {
	var value int64
	for {
		content, err := os.ReadFile(s.axisPaths[axis])
		if err != nil {
			log.Fatalf("Can't read sensor value from in_accel_%s", axis)
		}
		value, err = strconv.ParseInt(strings.TrimSpace(string(content)), 0, 64)
		if err == nil && value != 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	result := float64(math.Abs(float64(value)) * s.scale)

	s.axisValues[axis] = result
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

func getSensorPath(sensor string) string {
	fp, err := filepath.Glob(fmt.Sprintf("/sys/bus/iio/devices/iio:device*/%s", sensor))
	if err != nil || len(fp) == 0 {
		log.Fatal("Can't find a sensors")
	}
	return fp[0]
}
