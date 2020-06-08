package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	datafusion "mainprocess/datafusion"
	ml "mainprocess/logisticregression"

	mqtt "github.com/eclipse/paho.mqtt.golang"
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

var c mqtt.Client
var txFlag bool = false

var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	if !txFlag {
		txFlag = true
		byteData, err := json.Marshal(txFlag)
		if err != nil {
			log.Infof(err.Error())
			return
		}
		token := c.Publish("Node/Flag", 0, false, byteData)
		if token.Wait() && token.Error() != nil {
			panic(fmt.Sprintf("Error publishing: %v", token.Error()))
		}
		go func() {
			time.Sleep(200 * time.Millisecond)
			txFlag = false
		}()
		log.Infof("We need to activate the flag!")
	}
	log.Debugf("Received: %v", string(msg.Payload()))
	split := strings.Split(msg.Topic(), "/")
	sensor := split[len(split)-1]
	receivedData.AddNewValue(msg.Payload(), sensor)
}

var ff mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	err := json.Unmarshal(msg.Payload(), &txFlag)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	log.Infof("MSG: %v, TOPIC: %v", txFlag, msg.Topic())
}

func init() {
	log.SetLevel(log.DebugLevel)

	// receivedData = datafusion.CollectData{}

	// trainData, err := ml.LoadTrainDataFromCSV("./data/trackDataTrain.csv", "./data/trackDataTrain.csv")
	// if err != nil {
	// 	log.Errorf(err.Error())
	// 	os.Exit(400)
	// }

	// trainModel, err = trainData.CreateBestModel()
	// if err != nil {
	// 	log.Errorf(err.Error())
	// 	os.Exit(400)
	// }

	// log.Debugf("ModelData: %#v", trainModel)
}

func main() {
	opts := mqtt.NewClientOptions().AddBroker("tcp://127.0.0.1:1883").SetClientID("test-client")
	opts.SetKeepAlive(2 * time.Second)
	opts.SetPingTimeout(1 * time.Second)

	log.Infof("Connecting to MQTT broker...")
	c = mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		log.Errorf(token.Error().Error())
	}

	log.Infof("Subscribing to MQTT Topic...")
	if token := c.Subscribe("Node/Sensor/+", 0, f); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	if token := c.Subscribe("Node/Flag", 0, ff); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	// // Send some data for each sensor (random data)
	go auxSendCamera()
	go auxSendPresence()
	go auxSendRfid()
	go auxSendWifi()

	// time.Sleep(4 * time.Second)
	log.Infof("FINISHED DATA TRANSMISSION")

	// t1 := time.Now()

	// // Calculate the AVG result / list of results from the whole data received from each sensor
	// generatedData = datafusion.JoinedData{}
	// generatedData.Camera.Sensor = "camera"
	// generatedData.GetFinalCameraValues(receivedData)
	// generatedData.Presence.Sensor = "presence"
	// generatedData.GetFinalPresenceValues(receivedData)
	// generatedData.Rfid.Sensor = "rfid"
	// generatedData.GetFinalRfidValues(receivedData)
	// generatedData.Wifi.Sensor = "camera"
	// generatedData.GetFinalWifiValues(receivedData)

	// log.Debugf("CAMERA -> %#v", generatedData.Camera)
	// log.Debugf("PRESENCE -> %#v", generatedData.Presence)
	// log.Debugf("RFID -> %#v", generatedData.Rfid)
	// log.Debugf("WiFi -> %#v", generatedData.Wifi)

	// // Obtain a final list with the data to send to the ML algorithm
	// predictionDataStruct = datafusion.FinalData{}
	// predictionDataStruct.ObtainFinalData(generatedData)

	// t2 := time.Now()

	// result, _ := json.MarshalIndent(predictionDataStruct, "", "  ")
	// log.Infof(string(result))

	// log.Infof("Time doing join and calculating final data array: %v", t2.Sub(t1))

	// // predictionData := predictionDataStruct.To2DFloatArray()
	// var predictionData [][]float64
	// data1 := []float64{76.32, 1.43, 1.43, 21.65, 12.98}
	// data2 := []float64{76.32, 1.43, 71.43, 75.65, 12.98}
	// data3 := []float64{25.32, 0.43, 1.43, 21.65, 35.98}
	// data4 := []float64{28.32, 1.43, 1.43, 21.65, 89.98}
	// predictionData = append(predictionData, data1, data2, data3, data4)
	// log.Infof("Obtained 2D Array to predict: %v", predictionData)

	// prediction, err := trainModel.MakePrediction(predictionData)
	// if err != nil {
	// 	log.Errorf(err.Error())
	// }

	// log.Infof("Result of prediction: %v", prediction)

	// In order to keep the code running. Provisional
	fmt.Scanln()
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
		byteData, err := json.Marshal(data)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		token := c.Publish("Node/Sensor/Camera", 0, false, byteData)
		if token.Wait() && token.Error() != nil {
			panic(fmt.Sprintf("Error publishing: %v", token.Error()))
		}
		// dataReceiver(data, data["sensor"].(string))
		i++
	}
	for i < 40 {
		timestamp := 123456789 + i
		data := map[string]interface{}{
			"sensor":    "camera",
			"timestamp": strconv.Itoa(timestamp),
			"person":    7,
		}
		byteData, err := json.Marshal(data)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		token := c.Publish("Node/Sensor/Camera", 0, false, byteData)
		if token.Wait() && token.Error() != nil {
			panic(fmt.Sprintf("Error publishing: %v", token.Error()))
		}
		i++
	}
	for i < 100 {
		timestamp := 123456789 + i
		data := map[string]interface{}{
			"sensor":    "camera",
			"timestamp": strconv.Itoa(timestamp),
			"person":    9,
		}
		byteData, err := json.Marshal(data)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		token := c.Publish("Node/Sensor/Camera", 0, false, byteData)
		if token.Wait() && token.Error() != nil {
			panic(fmt.Sprintf("Error publishing: %v", token.Error()))
		}
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
		byteData, err := json.Marshal(data)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		token := c.Publish("Node/Sensor/Presence", 0, false, byteData)
		if token.Wait() && token.Error() != nil {
			panic(fmt.Sprintf("Error publishing: %v", token.Error()))
		}
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
		byteData, err := json.Marshal(data)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		token := c.Publish("Node/Sensor/Rfid", 0, false, byteData)
		if token.Wait() && token.Error() != nil {
			panic(fmt.Sprintf("Error publishing: %v", token.Error()))
		}
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
		byteData, err := json.Marshal(data)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		token := c.Publish("Node/Sensor/Rfid", 0, false, byteData)
		if token.Wait() && token.Error() != nil {
			panic(fmt.Sprintf("Error publishing: %v", token.Error()))
		}
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
		byteData, err := json.Marshal(data)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		token := c.Publish("Node/Sensor/Rfid", 0, false, byteData)
		if token.Wait() && token.Error() != nil {
			panic(fmt.Sprintf("Error publishing: %v", token.Error()))
		}
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
		byteData, err := json.Marshal(data)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		token := c.Publish("Node/Sensor/Wifi", 0, false, byteData)
		if token.Wait() && token.Error() != nil {
			panic(fmt.Sprintf("Error publishing: %v", token.Error()))
		}
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
