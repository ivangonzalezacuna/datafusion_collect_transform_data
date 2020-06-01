package main

import (
	"encoding/json"
	"math/rand"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type (
	// CameraData stores the data received from the camera
	CameraData struct {
		Sensor    string `json:"sensor"`
		Timestamp string `json:"timestamp"`
		Person    int    `json:"person"`
	}
	// PresenceData stores the data received from the presence detection sensor
	PresenceData struct {
		Sensor    string `json:"sensor"`
		Timestamp string `json:"timestamp"`
		Detection bool   `json:"detection"`
	}
	// RfidData stores the data received from the rfid reader
	RfidData struct {
		Sensor    string  `json:"sensor"`
		Timestamp string  `json:"timestamp"`
		Power     float64 `json:"power"`
		Person    int     `json:"person"`
	}
	// WifiData stores the data received from the wifi
	WifiData struct {
		Sensor           string `json:"sensor"`
		Timestamp        string `json:"timestamp"`
		ConnectedDevices int    `json:"connecteddevices"`
	}

	// FinalCameraData stores the final fusion data for the camera
	FinalCameraData struct {
		Sensor      string                   `json:"sensor"`
		Timestamp   string                   `json:"timestamp"`
		PersonCount []FinalCameraPersonCount `json:"personcount"`
	}
	//FinalCameraPersonCount stores the list
	FinalCameraPersonCount struct {
		Count  int `json:"count"`
		Person int `json:"person"`
	}

	// FinalPresenceData stores the final fusion data for the presence detector
	FinalPresenceData struct {
		Sensor    string  `json:"sensor"`
		Timestamp string  `json:"timestamp"`
		Detection float64 `json:"detection"`
	}

	// FinalRfidData stores the final fusion data for the RFID reader
	FinalRfidData struct {
		Sensor      string                 `json:"sensor"`
		Timestamp   string                 `json:"timestamp"`
		PersonCount []FinalRfidPersonCount `json:"personcount"`
	}
	//FinalRfidPersonCount stores the list
	FinalRfidPersonCount struct {
		Count  int     `json:"count"`
		Power  float64 `json:"power"`
		Person int     `json:"person"`
	}

	// FinalWifiData stores the final fusion data for the WiFi
	FinalWifiData struct {
		Sensor           string  `json:"sensor"`
		Timestamp        string  `json:"timestamp"`
		ConnectedDevices float64 `json:"person"`
	}
)

var (
	cameraInfo        []CameraData
	presenceInfo      []PresenceData
	rfidInfo          []RfidData
	wifiInfo          []WifiData
	finalCameraInfo   FinalCameraData
	finalPresenceInfo FinalPresenceData
	finalRfidInfo     FinalRfidData
	finalWifiInfo     FinalWifiData
)

func init() {
	log.SetLevel(log.DebugLevel)
	cameraInfo = nil
	presenceInfo = nil
	rfidInfo = nil
	wifiInfo = nil
}

func main() {
	log.Info("Hello")
	var i int = 0
	for i < 20 {
		timestamp := 123456789 + i
		data := CameraData{
			Sensor:    "camera",
			Timestamp: strconv.Itoa(timestamp),
			Person:    5,
		}
		dataReceiver(data, data.Sensor)
		i++
	}
	for i < 30 {
		timestamp := 123456789 + i
		data := CameraData{
			Sensor:    "camera",
			Timestamp: strconv.Itoa(timestamp),
			Person:    7,
		}
		dataReceiver(data, data.Sensor)
		i++
	}
	i = 0
	for i < 20 {
		timestamp := 123456789 + i
		detection := (i%2 == 0) || (i%3 == 0)
		data := PresenceData{
			Sensor:    "presence",
			Timestamp: strconv.Itoa(timestamp),
			Detection: detection,
		}
		dataReceiver(data, data.Sensor)
		i++
	}
	i = 0
	for i < 20 {
		timestamp := 123456789 + i
		data := RfidData{
			Sensor:    "rfid",
			Timestamp: strconv.Itoa(timestamp),
			Power:     float64(i),
			Person:    5,
		}
		dataReceiver(data, data.Sensor)
		i++
	}
	for i < 30 {
		timestamp := 123456789 + i
		data := RfidData{
			Sensor:    "rfid",
			Timestamp: strconv.Itoa(timestamp),
			Power:     float64(i),
			Person:    7,
		}
		dataReceiver(data, data.Sensor)
		i++
	}
	i = 0
	for i < 20 {
		timestamp := 123456789 + i
		min := 0
		max := 3
		data := WifiData{
			Sensor:           "wifi",
			Timestamp:        strconv.Itoa(timestamp),
			ConnectedDevices: rand.Intn(max-min) + min,
		}
		dataReceiver(data, data.Sensor)
		i++
	}
	// log.Infof("Camera: %v\nData: %#v\n\n", len(cameraInfo), cameraInfo)
	// log.Infof("Prensence: %v\nData: %#v\n\n", len(presenceInfo), presenceInfo)
	// log.Infof("RFID: %v\nData: %#v\n\n", len(rfidInfo), rfidInfo)
	// log.Infof("WiFi: %v\nData: %#v\n\n", len(wifiInfo), wifiInfo)

	log.Infof("FINISHED DATA TRANSMISSION")

	calculateCameraFinalValues()
	calculatePresenceFinalValues()
	calculateRfidFinalValues()
	calculateWifiFinalValues()

	log.Infof("Final camera info: %#v", finalCameraInfo)
	log.Infof("Final presence info: %#v", finalPresenceInfo)
	log.Infof("Final RFID info: %#v", finalRfidInfo)
	log.Infof("Final WiFi info: %#v", finalWifiInfo)
}

func dataReceiver(u interface{}, topic string) {
	byteData, err := json.Marshal(u)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	log.Infof(string(byteData))
	switch topic {
	case "camera":
		var c CameraData
		err = json.Unmarshal(byteData, &c)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		cameraInfo = append(cameraInfo, c)
		log.Infof("Camera detected")
	case "presence":
		var p PresenceData
		err = json.Unmarshal(byteData, &p)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		presenceInfo = append(presenceInfo, p)
		log.Infof("Presence detected")
	case "rfid":
		var r RfidData
		err = json.Unmarshal(byteData, &r)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		rfidInfo = append(rfidInfo, r)
		log.Infof("RFID detected")
	case "wifi":
		var w WifiData
		err = json.Unmarshal(byteData, &w)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		wifiInfo = append(wifiInfo, w)
		log.Infof("WiFi detected")
	}
}

func calculateCameraFinalValues() error {
	finalCameraInfo = FinalCameraData{Sensor: "camera"}
	if len(cameraInfo) == 0 {
		return nil
	}
	finalCameraInfo.Timestamp = cameraInfo[0].Timestamp

	peopleCount := make(map[int]int)
	for _, v := range cameraInfo {
		_, exist := peopleCount[v.Person]
		if exist {
			peopleCount[v.Person]++
		} else {
			peopleCount[v.Person] = 1
		}
	}

	for k, v := range peopleCount {
		currentCount := FinalCameraPersonCount{Person: k, Count: v}
		finalCameraInfo.PersonCount = append(finalCameraInfo.PersonCount, currentCount)
		log.Infof("Person : %d , Count : %d\n", k, v)
	}

	return nil
}

func calculatePresenceFinalValues() error {
	finalPresenceInfo.Sensor = "presence"
	if len(presenceInfo) == 0 {
		return nil
	}
	finalPresenceInfo.Timestamp = presenceInfo[0].Timestamp

	totalEntries := len(presenceInfo)
	var positiveEntries int = 0
	for _, v := range presenceInfo {
		if v.Detection {
			positiveEntries++
		}
	}
	var detectionAvg float64
	detectionAvg = (float64(positiveEntries) / float64(totalEntries)) * 100
	log.Infof("Detection AVG: %f", detectionAvg)
	finalPresenceInfo.Detection = detectionAvg

	return nil
}

type auxRfid struct {
	count int
	total float64
}

func calculateRfidFinalValues() error {
	finalRfidInfo.Sensor = "rfid"
	if len(rfidInfo) == 0 {
		return nil
	}
	finalRfidInfo.Timestamp = rfidInfo[0].Timestamp

	peopleCount := make(map[int]auxRfid)
	for _, v := range rfidInfo {
		_, exist := peopleCount[v.Person]
		var data auxRfid
		if exist {
			data = peopleCount[v.Person]
			data.count++
			data.total += v.Power
			peopleCount[v.Person] = data
		} else {
			data.count = 1
			data.total = v.Power
			peopleCount[v.Person] = data
		}
	}

	for k, v := range peopleCount {
		var powerAvg float64
		powerAvg = v.total / float64(v.count)
		currentCount := FinalRfidPersonCount{Person: k, Count: v.count, Power: powerAvg}
		finalRfidInfo.PersonCount = append(finalRfidInfo.PersonCount, currentCount)
		log.Infof("Person : %d , Data : %v\n", k, v)
	}

	return nil
}

func calculateWifiFinalValues() error {
	finalWifiInfo.Sensor = "wifi"
	if len(wifiInfo) == 0 {
		return nil
	}
	finalWifiInfo.Timestamp = wifiInfo[0].Timestamp

	totalEntries := len(wifiInfo)
	var totalConnDevices int = 0
	for _, v := range wifiInfo {
		totalConnDevices += v.ConnectedDevices
	}
	var devicesAvg float64
	devicesAvg = float64(totalConnDevices) / float64(totalEntries)
	log.Infof("Devices connected AVG: %f", devicesAvg)
	finalWifiInfo.ConnectedDevices = devicesAvg

	return nil
}
