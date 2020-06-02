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
		ConnectedDevices float64 `json:"connecteddevices"`
	}

	// FinalDataFusion is the struct with the data to send to the LogisticRegression Model
	FinalDataFusion struct {
		Timestamp   string  `json:"timestamp"`
		Person      int     `json:"person"`
		Presence    float64 `json:"presence"`
		ConnDevices float64 `json:"conndevices"`
		RfidUser    float64 `json:"rfiduser"`
		RfidPower   float64 `json:"power"`
		CameraUser  float64 `json:"camerauser"`
	}
)

var (
	// Initial array of data received from each sensor
	cameraInfo   []CameraData
	presenceInfo []PresenceData
	rfidInfo     []RfidData
	wifiInfo     []WifiData

	// Struct generated from the received data of each sensor
	finalCameraInfo   FinalCameraData
	finalPresenceInfo FinalPresenceData
	finalRfidInfo     FinalRfidData
	finalWifiInfo     FinalWifiData

	// Array of data for each detected user by the camera and/or rfid reader
	finalDataFusion []FinalDataFusion
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func auxSendCamera() {
	i := 0
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
	timestamp := 123456789 + i
	data := CameraData{
		Sensor:    "camera",
		Timestamp: strconv.Itoa(timestamp),
		Person:    9,
	}
	dataReceiver(data, data.Sensor)
}

func auxSendPresence() {
	i := 0
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
}

func auxSendRfid() {
	i := 0
	for i < 20 {
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
	for i < 30 {
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
	timestamp := 123456789 + i
	data := RfidData{
		Sensor:    "rfid",
		Timestamp: strconv.Itoa(timestamp),
		Power:     float64(i),
		Person:    6,
	}
	dataReceiver(data, data.Sensor)
}

func auxSendWifi() {
	i := 0
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
}

func main() {
	// Send some data for each sensor (random data)
	auxSendCamera()
	auxSendPresence()
	auxSendRfid()
	auxSendWifi()

	log.Infof("FINISHED DATA TRANSMISSION")

	// Calculate the AVG result / list of results from the whole data received from each sensor
	calculateCameraFinalValues()
	calculatePresenceFinalValues()
	calculateRfidFinalValues()
	calculateWifiFinalValues()

	log.Debugf("CAMERA -> %#v", finalCameraInfo)
	log.Debugf("PRESENCE -> %#v", finalPresenceInfo)
	log.Debugf("RFID -> %#v", finalRfidInfo)
	log.Debugf("WiFi -> %#v", finalWifiInfo)

	// Obtain a final list with the data to send to the ML algorithm
	obtainFinalStruct()
	for k, v := range finalDataFusion {
		log.Debugf("DATA FUSION [%d] -> %#v", k, v)
	}
}

func obtainFinalStruct() {
	finalTimestamp := finalCameraInfo.Timestamp
	if finalPresenceInfo.Timestamp < finalTimestamp || finalTimestamp == "" {
		finalTimestamp = finalPresenceInfo.Timestamp
	}
	if finalRfidInfo.Timestamp < finalTimestamp || finalTimestamp == "" {
		finalTimestamp = finalRfidInfo.Timestamp
	}
	if finalWifiInfo.Timestamp < finalTimestamp || finalTimestamp == "" {
		finalTimestamp = finalWifiInfo.Timestamp
	}

	totalCameraCount := 0
	totalRfidCount := 0

	for _, v := range finalCameraInfo.PersonCount {
		totalCameraCount += v.Count
	}
	for _, v := range finalRfidInfo.PersonCount {
		totalRfidCount += v.Count
	}

	avgCameraUserMap := make(map[int]float64, len(finalCameraInfo.PersonCount))
	avgRfidUserMap := make(map[int]float64, len(finalRfidInfo.PersonCount))

	for _, v := range finalCameraInfo.PersonCount {
		avgPerson := float64(v.Count) / float64(totalCameraCount)
		avgCameraUserMap[v.Person] = avgPerson
	}

	for _, v := range finalRfidInfo.PersonCount {
		avgPerson := float64(v.Count) / float64(totalRfidCount)
		avgRfidUserMap[v.Person] = avgPerson
	}

	log.Tracef("%#v", avgCameraUserMap)
	log.Tracef("%#v", avgRfidUserMap)

	for k, v := range avgCameraUserMap {
		if !isPersonEntryCreated(k) {
			var currentData FinalDataFusion
			if val, ok := avgRfidUserMap[k]; ok {
				currentData = generateCurrentData(finalTimestamp, k, val, v)
			} else {
				currentData = generateCurrentData(finalTimestamp, k, 0, v)
			}
			finalDataFusion = append(finalDataFusion, currentData)
		} else {
			log.Infof("Person %d already saved in struct slice", k)
		}
	}

	for k, v := range avgRfidUserMap {
		if !isPersonEntryCreated(k) {
			var currentData FinalDataFusion
			if val, ok := avgCameraUserMap[k]; ok {
				currentData = generateCurrentData(finalTimestamp, k, v, val)
			} else {
				currentData = generateCurrentData(finalTimestamp, k, v, 0)
			}
			finalDataFusion = append(finalDataFusion, currentData)
		} else {
			log.Infof("Person %d already saved in struct slice", k)
		}
	}
}

func generateCurrentData(finalTimestamp string, person int, rfidUser, camUser float64) FinalDataFusion {
	currentData := FinalDataFusion{
		Timestamp:   finalTimestamp,
		Person:      person,
		Presence:    finalPresenceInfo.Detection,
		ConnDevices: finalWifiInfo.ConnectedDevices,
		RfidUser:    rfidUser,
		CameraUser:  camUser,
	}
	for _, data := range finalRfidInfo.PersonCount {
		if data.Person == currentData.Person {
			currentData.RfidPower = data.Power
		}
	}
	log.Tracef("Current Data info: %#v", currentData)
	return currentData
}

func isPersonEntryCreated(person int) bool {
	for _, v := range finalDataFusion {
		if person == v.Person {
			return true
		}
	}
	return false
}

func dataReceiver(u interface{}, topic string) {
	byteData, err := json.Marshal(u)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	log.Tracef(string(byteData))
	switch topic {
	case "camera":
		var c CameraData
		err = json.Unmarshal(byteData, &c)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		cameraInfo = append(cameraInfo, c)
		log.Tracef("Camera detected")
	case "presence":
		var p PresenceData
		err = json.Unmarshal(byteData, &p)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		presenceInfo = append(presenceInfo, p)
		log.Tracef("Presence detected")
	case "rfid":
		var r RfidData
		err = json.Unmarshal(byteData, &r)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		rfidInfo = append(rfidInfo, r)
		log.Tracef("RFID detected")
	case "wifi":
		var w WifiData
		err = json.Unmarshal(byteData, &w)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		wifiInfo = append(wifiInfo, w)
		log.Tracef("WiFi detected")
	}
}

func calculateCameraFinalValues() error {
	finalCameraInfo.Sensor = "camera"
	if len(cameraInfo) == 0 {
		return nil
	}
	finalCameraInfo.Timestamp = cameraInfo[0].Timestamp

	peopleCount := make(map[int]int)
	for _, v := range cameraInfo {
		if _, exist := peopleCount[v.Person]; exist {
			peopleCount[v.Person]++
		} else {
			peopleCount[v.Person] = 1
		}
	}

	for k, v := range peopleCount {
		currentCount := FinalCameraPersonCount{Person: k, Count: v}
		finalCameraInfo.PersonCount = append(finalCameraInfo.PersonCount, currentCount)
		log.Debugf("CAMERA -> Person : %d , Count : %d\n", k, v)
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
	detectionAvg := (float64(positiveEntries) / float64(totalEntries)) * 100
	log.Debugf("PRESENCE AVG: %.2f", detectionAvg)
	finalPresenceInfo.Detection = detectionAvg

	return nil
}

func calculateRfidFinalValues() error {
	finalRfidInfo.Sensor = "rfid"
	if len(rfidInfo) == 0 {
		return nil
	}
	finalRfidInfo.Timestamp = rfidInfo[0].Timestamp

	peopleCount := make(map[int]struct {
		count int
		total float64
	})
	for _, v := range rfidInfo {
		if data, exist := peopleCount[v.Person]; exist {
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
		powerAvg := v.total / float64(v.count)
		finalRfidInfo.PersonCount = append(finalRfidInfo.PersonCount, FinalRfidPersonCount{Person: k, Count: v.count, Power: powerAvg})
		log.Debugf("RFID -> Person: %d , Data: %v\n", k, v)
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
	devicesAvg := float64(totalConnDevices) / float64(totalEntries)
	log.Debugf("WiFi AVG: %f", devicesAvg)
	finalWifiInfo.ConnectedDevices = devicesAvg

	return nil
}
