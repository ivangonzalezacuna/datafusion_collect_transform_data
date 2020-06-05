package main

import (
	"encoding/json"
	"math/rand"
	"os"
	"strconv"
	"time"

	datafusion "mainprocess/datafusion"
	ml "mainprocess/logisticregression"

	log "github.com/sirupsen/logrus"
)

var (
	// Initial array of data received from each sensor
	receivedData datafusion.CollectData
	// Struct generated from the received data of each sensor
	generatedData datafusion.JoinedData
	// Array of data for each detected user by the camera and/or rfid reader
	predictionDataStruct datafusion.FinalData
	// Struct to store the train data for the Logistic Regression
	trainData ml.TrainData
	// Structrained Logistic Regression Model and other related data
	trainModel ml.ModelData
)

func init() {
	log.SetLevel(log.DebugLevel)
	receivedData = datafusion.CollectData{}

	trainData, err := ml.LoadTrainDataFromCSV("./data/trackDataTrain.csv", "./data/trackDataTrain.csv")
	if err != nil {
		log.Errorf(err.Error())
		os.Exit(400)
	}

	trainModel, err = trainData.CreateBestModel()
	if err != nil {
		log.Errorf(err.Error())
		os.Exit(400)
	}

	log.Debugf("ModelData: %#v", trainModel)
}

func main() {
	// Send some data for each sensor (random data)
	auxSendCamera()
	auxSendPresence()
	auxSendRfid()
	auxSendWifi()

	log.Infof("FINISHED DATA TRANSMISSION")

	t1 := time.Now()

	// Calculate the AVG result / list of results from the whole data received from each sensor
	generatedData = datafusion.JoinedData{}
	generatedData.Camera.Sensor = "camera"
	generatedData.GetFinalCameraValues(receivedData)
	generatedData.Presence.Sensor = "presence"
	generatedData.GetFinalPresenceValues(receivedData)
	generatedData.Rfid.Sensor = "rfid"
	generatedData.GetFinalRfidValues(receivedData)
	generatedData.Wifi.Sensor = "camera"
	generatedData.GetFinalWifiValues(receivedData)

	log.Debugf("CAMERA -> %#v", generatedData.Camera)
	log.Debugf("PRESENCE -> %#v", generatedData.Presence)
	log.Debugf("RFID -> %#v", generatedData.Rfid)
	log.Debugf("WiFi -> %#v", generatedData.Wifi)

	// Obtain a final list with the data to send to the ML algorithm
	predictionDataStruct = datafusion.FinalData{}
	predictionDataStruct.ObtainFinalData(generatedData)

	t2 := time.Now()

	result, _ := json.MarshalIndent(predictionDataStruct, "", "  ")
	log.Infof(string(result))

	log.Infof("Time doing join and calculating final data array: %v", t2.Sub(t1))

	predictionData := predictionDataStruct.To2DFloatArray()
	log.Infof("Obtained 2D Array to predict: %v", predictionData)

	prediction, err := trainModel.MakePrediction(predictionData)
	if err != nil {
		log.Errorf(err.Error())
	}

	log.Infof("Result of prediction: %v", prediction)
}

func auxSendCamera() {
	i := 0
	for i < 20 {
		timestamp := 123456789 + i
		data := map[string]interface{}{
			"sensor":    "camera",
			"timestamp": strconv.Itoa(timestamp),
			"person":    5,
		}
		dataReceiver(data, data["sensor"].(string))
		i++
	}
	for i < 40 {
		timestamp := 123456789 + i
		data := map[string]interface{}{
			"sensor":    "camera",
			"timestamp": strconv.Itoa(timestamp),
			"person":    7,
		}
		dataReceiver(data, data["sensor"].(string))
		i++
	}
	for i < 100 {
		timestamp := 123456789 + i
		data := map[string]interface{}{
			"sensor":    "camera",
			"timestamp": strconv.Itoa(timestamp),
			"person":    9,
		}
		dataReceiver(data, data["sensor"].(string))
		i++
	}
}

func auxSendPresence() {
	i := 0
	for i < 10 {
		timestamp := 123456789 + i
		detection := (i%7 == 0) // || (i%3 == 0)
		data := map[string]interface{}{
			"sensor":    "presence",
			"timestamp": strconv.Itoa(timestamp),
			"detection": detection,
		}
		dataReceiver(data, data["sensor"].(string))
		i++
	}
}

func auxSendRfid() {
	i := 0
	min := 0
	max := 100
	for i < 2 {
		timestamp := 123456789 + i
		data := map[string]interface{}{
			"sensor":    "rfid",
			"timestamp": strconv.Itoa(timestamp),
			"power":     float64(rand.Intn(max-min) + min),
			"person":    7,
		}
		dataReceiver(data, data["sensor"].(string))
		i++
	}
	for i < 3 {
		timestamp := 123456789 + i
		data := map[string]interface{}{
			"sensor":    "rfid",
			"timestamp": strconv.Itoa(timestamp),
			"power":     float64(rand.Intn(max-min) + min),
			"person":    5,
		}
		dataReceiver(data, data["sensor"].(string))
		i++
	}
	for i < 70 {
		timestamp := 123456789 + i
		data := map[string]interface{}{
			"sensor":    "rfid",
			"timestamp": strconv.Itoa(timestamp),
			"power":     float64(rand.Intn(max-min) + min),
			"person":    6,
		}
		dataReceiver(data, data["sensor"].(string))
		i++
	}
}

func auxSendWifi() {
	i := 0
	for i < 20 {
		timestamp := 123456789 + i
		min := 0
		max := 3
		data := map[string]interface{}{
			"sensor":           "wifi",
			"timestamp":        strconv.Itoa(timestamp),
			"connecteddevices": rand.Intn(max-min) + min,
		}
		dataReceiver(data, data["sensor"].(string))
		i++
	}
}

func dataReceiver(u interface{}, topic string) {
	byteData, err := json.Marshal(u)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	log.Trace("Received: %v", string(byteData))
	receivedData.AddNewValue(byteData, topic)
}
