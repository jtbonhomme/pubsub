package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/websocket"

	"github.com/jtbonhomme/pubsub"
)

type RideMessage struct {
	RideID         string  `json:"ride_id"`
	PointIdx       int     `json:"point_idx"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	Timestamp      string  `json:"timestamp"`
	MeterReading   float64 `json:"meter_reading"`
	MeterIncrement float64 `json:"meter_increment"`
	RideStatus     string  `json:"ride_status"`
	PassengerCount int     `json:"passenger_count"`
}

const tsFormat string = "2006-01-02T15:04:05.0000-07:00"

var rideStatus = []string{"STARTED", "WAITING", "FINISHED", "BLOCKED"}

func main() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	origin := "http://localhost/"
	url := "ws://localhost:12345/ws"
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Fatal(err)
	}

	payload, err := json.Marshal(RideMessage{
		RideID:         uuid.NewString(),
		PointIdx:       r.Intn(10),
		Latitude:       r.Float64() + float64(48),
		Longitude:      r.Float64() + float64(2),
		Timestamp:      time.Now().Format(tsFormat),
		MeterReading:   r.Float64() * 50,
		MeterIncrement: r.Float64() * 2,
		PassengerCount: r.Intn(5),
		RideStatus:     rideStatus[r.Intn(3)],
	})
	if err != nil {
		log.Fatal(err)
	}

	m := pubsub.Message{
		Action:  "publish",
		Topic:   "com.jtbonhomme.pubsub.rides",
		Payload: payload,
	}

	msg, err := json.Marshal(m)
	if err != nil {
		log.Fatal(err)
	}

	_, err = ws.Write(msg)
	if err != nil {
		log.Fatal(err)
	}
}
