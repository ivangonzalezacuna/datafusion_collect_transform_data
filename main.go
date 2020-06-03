package main

import (
	"encoding/json"
	"math/rand"
	"strconv"

	ml "mainProcess/logisticregression"

	log "github.com/sirupsen/logrus"
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

	cameraStructFinal struct {
		Sensor      string                   `json:"sensor"`
		Timestamp   string                   `json:"timestamp"`
		PersonCount []cameraStructCountFinal `json:"personcount"`
	}
	cameraStructCountFinal struct {
		Count  int `json:"count"`
		Person int `json:"person"`
	}

	presenceStructFinal struct {
		Sensor    string  `json:"sensor"`
		Timestamp string  `json:"timestamp"`
		Detection float64 `json:"detection"`
	}

	rfidStructFinal struct {
		Sensor      string                 `json:"sensor"`
		Timestamp   string                 `json:"timestamp"`
		PersonCount []rfidStructCountFinal `json:"personcount"`
	}
	rfidStructCountFinal struct {
		Count  int     `json:"count"`
		Power  float64 `json:"power"`
		Person int     `json:"person"`
	}

	wifiStructFinal struct {
		Sensor           string  `json:"sensor"`
		Timestamp        string  `json:"timestamp"`
		ConnectedDevices float64 `json:"connecteddevices"`
	}

	// CollectData stores all the structs with received data
	CollectData struct {
		Camera   []cameraStruct
		Presence []presenceStruct
		Rfid     []rfidStruct
		Wifi     []wifiStruct
	}

	// JoinedData stores the final data fusion from the sensors
	JoinedData struct {
		Camera   cameraStructFinal
		Presence presenceStructFinal
		Rfid     rfidStructFinal
		Wifi     wifiStructFinal
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
	receivedData CollectData

	// Struct generated from the received data of each sensor
	generatedData JoinedData

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
		data := cameraStruct{
			Sensor:    "camera",
			Timestamp: strconv.Itoa(timestamp),
			Person:    5,
		}
		dataReceiver(data, data.Sensor)
		i++
	}
	for i < 30 {
		timestamp := 123456789 + i
		data := cameraStruct{
			Sensor:    "camera",
			Timestamp: strconv.Itoa(timestamp),
			Person:    7,
		}
		dataReceiver(data, data.Sensor)
		i++
	}
	timestamp := 123456789 + i
	data := cameraStruct{
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
		data := presenceStruct{
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
		data := rfidStruct{
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
		data := rfidStruct{
			Sensor:    "rfid",
			Timestamp: strconv.Itoa(timestamp),
			Power:     float64(i),
			Person:    5,
		}
		dataReceiver(data, data.Sensor)
		i++
	}
	timestamp := 123456789 + i
	data := rfidStruct{
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
		data := wifiStruct{
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

	log.Debugf("CAMERA -> %#v", generatedData.Camera)
	log.Debugf("PRESENCE -> %#v", generatedData.Presence)
	log.Debugf("RFID -> %#v", generatedData.Rfid)
	log.Debugf("WiFi -> %#v", generatedData.Wifi)

	// Obtain a final list with the data to send to the ML algorithm
	obtainFinalStruct()
	for k, v := range finalDataFusion {
		log.Debugf("DATA FUSION [%d] -> %#v", k, v)
	}

	err := ml.LoadTrainData("./data/trackDataTrain.csv", "./data/trackDataTrain.csv")
	if err != nil {
		log.Errorf(err.Error())
	}
	log.Infof("Trained ML algorithm")
}

func obtainFinalStruct() {
	finalTimestamp := generatedData.Camera.Timestamp
	if generatedData.Presence.Timestamp < finalTimestamp || finalTimestamp == "" {
		finalTimestamp = generatedData.Presence.Timestamp
	}
	if generatedData.Rfid.Timestamp < finalTimestamp || finalTimestamp == "" {
		finalTimestamp = generatedData.Rfid.Timestamp
	}
	if generatedData.Wifi.Timestamp < finalTimestamp || finalTimestamp == "" {
		finalTimestamp = generatedData.Wifi.Timestamp
	}

	totalCameraCount := 0
	totalRfidCount := 0

	for _, v := range generatedData.Camera.PersonCount {
		totalCameraCount += v.Count
	}
	for _, v := range generatedData.Rfid.PersonCount {
		totalRfidCount += v.Count
	}

	avgCameraUserMap := make(map[int]float64, len(generatedData.Camera.PersonCount))
	avgRfidUserMap := make(map[int]float64, len(generatedData.Rfid.PersonCount))

	for _, v := range generatedData.Camera.PersonCount {
		avgPerson := float64(v.Count) / float64(totalCameraCount)
		avgCameraUserMap[v.Person] = avgPerson
	}

	for _, v := range generatedData.Rfid.PersonCount {
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
		Presence:    generatedData.Presence.Detection,
		ConnDevices: generatedData.Wifi.ConnectedDevices,
		RfidUser:    rfidUser,
		CameraUser:  camUser,
	}
	for _, data := range generatedData.Rfid.PersonCount {
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
		var c cameraStruct
		err = json.Unmarshal(byteData, &c)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		receivedData.Camera = append(receivedData.Camera, c)
		log.Tracef("Camera detected")
	case "presence":
		var p presenceStruct
		err = json.Unmarshal(byteData, &p)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		receivedData.Presence = append(receivedData.Presence, p)
		log.Tracef("Presence detected")
	case "rfid":
		var r rfidStruct
		err = json.Unmarshal(byteData, &r)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		receivedData.Rfid = append(receivedData.Rfid, r)
		log.Tracef("RFID detected")
	case "wifi":
		var w wifiStruct
		err = json.Unmarshal(byteData, &w)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		receivedData.Wifi = append(receivedData.Wifi, w)
		log.Tracef("WiFi detected")
	}
}

func calculateCameraFinalValues() error {
	generatedData.Camera.Sensor = "camera"
	if len(receivedData.Camera) == 0 {
		return nil
	}
	generatedData.Camera.Timestamp = receivedData.Camera[0].Timestamp

	peopleCount := make(map[int]int)
	for _, v := range receivedData.Camera {
		if _, exist := peopleCount[v.Person]; exist {
			peopleCount[v.Person]++
		} else {
			peopleCount[v.Person] = 1
		}
	}

	for k, v := range peopleCount {
		currentCount := cameraStructCountFinal{Person: k, Count: v}
		generatedData.Camera.PersonCount = append(generatedData.Camera.PersonCount, currentCount)
		log.Debugf("CAMERA -> Person : %d , Count : %d\n", k, v)
	}

	return nil
}

func calculatePresenceFinalValues() error {
	generatedData.Presence.Sensor = "presence"
	if len(receivedData.Presence) == 0 {
		return nil
	}
	generatedData.Presence.Timestamp = receivedData.Presence[0].Timestamp

	totalEntries := len(receivedData.Presence)
	var positiveEntries int = 0
	for _, v := range receivedData.Presence {
		if v.Detection {
			positiveEntries++
		}
	}
	detectionAvg := (float64(positiveEntries) / float64(totalEntries)) * 100
	log.Debugf("PRESENCE AVG: %.2f", detectionAvg)
	generatedData.Presence.Detection = detectionAvg

	return nil
}

func calculateRfidFinalValues() error {
	generatedData.Rfid.Sensor = "rfid"
	if len(receivedData.Rfid) == 0 {
		return nil
	}
	generatedData.Rfid.Timestamp = receivedData.Rfid[0].Timestamp

	peopleCount := make(map[int]struct {
		count int
		total float64
	})
	for _, v := range receivedData.Rfid {
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
		generatedData.Rfid.PersonCount = append(generatedData.Rfid.PersonCount, rfidStructCountFinal{Person: k, Count: v.count, Power: powerAvg})
		log.Debugf("RFID -> Person: %d , Data: %v\n", k, v)
	}

	return nil
}

func calculateWifiFinalValues() error {
	generatedData.Wifi.Sensor = "wifi"
	if len(receivedData.Wifi) == 0 {
		return nil
	}
	generatedData.Wifi.Timestamp = receivedData.Wifi[0].Timestamp

	totalEntries := len(receivedData.Wifi)
	var totalConnDevices int = 0
	for _, v := range receivedData.Wifi {
		totalConnDevices += v.ConnectedDevices
	}
	devicesAvg := float64(totalConnDevices) / float64(totalEntries)
	log.Debugf("WiFi AVG: %f", devicesAvg)
	generatedData.Wifi.ConnectedDevices = devicesAvg

	return nil
}
