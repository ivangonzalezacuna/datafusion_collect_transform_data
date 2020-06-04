package mainprocess

import (
	"encoding/json"

	"github.com/micro/go-micro/util/log"
)

type (
	cameraStruct struct {
		Sensor    string `json:"sensor"`
		Timestamp string `json:"timestamp"`
		Person    int    `json:"person"`
	}

	presenceStruct struct {
		Sensor    string `json:"sensor"`
		Timestamp string `json:"timestamp"`
		Detection bool   `json:"detection"`
	}

	rfidStruct struct {
		Sensor    string  `json:"sensor"`
		Timestamp string  `json:"timestamp"`
		Power     float64 `json:"power"`
		Person    int     `json:"person"`
	}

	wifiStruct struct {
		Sensor           string `json:"sensor"`
		Timestamp        string `json:"timestamp"`
		ConnectedDevices int    `json:"connecteddevices"`
	}

	// CollectData stores all the structs with received data
	CollectData struct {
		Camera   []cameraStruct
		Presence []presenceStruct
		Rfid     []rfidStruct
		Wifi     []wifiStruct
	}
)

// AddNewValue adds a new entry in the sensor's received data slice
func (c *CollectData) AddNewValue(payload []byte, topic string) {
	switch topic {
	case "camera":
		var data cameraStruct
		err := json.Unmarshal(payload, &data)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		c.Camera = append(c.Camera, data)
		log.Tracef("Camera detected")
	case "presence":
		var data presenceStruct
		err := json.Unmarshal(payload, &data)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		c.Presence = append(c.Presence, data)
		log.Tracef("Presence detected")
	case "rfid":
		var data rfidStruct
		err := json.Unmarshal(payload, &data)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		c.Rfid = append(c.Rfid, data)
		log.Tracef("RFID detected")
	case "wifi":
		var data wifiStruct
		err := json.Unmarshal(payload, &data)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		c.Wifi = append(c.Wifi, data)
		log.Tracef("WiFi detected")
	}
}
