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
	ml "mainprocess/ml_regression_tracking"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	// Initial array of data received from each sensor
	receivedData datafusion.CollectData
	// Auxiliar array of data received from each sensor (to avoid re-write)
	receivedDataFinal datafusion.CollectData
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
var txFlag bool
var count int

var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	if !txFlag && count == 0 {
		count++
		value := true
		byteData, err := json.Marshal(value)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		token := c.Publish("Node/Flag", 0, false, byteData)
		if token.Wait() && token.Error() != nil {
			panic(fmt.Sprintf("Error publishing: %v", token.Error()))
		}

		go func() {
			viper.SetDefault("ml.window", 200)
			sleepTime := viper.GetInt("ml.window")
			time.Sleep(time.Duration(sleepTime) * time.Millisecond)
			value := false
			byteData, err := json.Marshal(value)
			if err != nil {
				log.Errorf(err.Error())
				return
			}
			log.Infof("[MQTT] Deactivating flag after 200ms!")
			token := c.Publish("Node/Flag", 0, false, byteData)
			if token.Wait() && token.Error() != nil {
				panic(fmt.Sprintf("Error publishing: %v", token.Error()))
			}
			receivedDataFinal = receivedData
			log.Infof("[MQTT] Camera size: %v\nPresence size: %v\nRfid size: %v\nWifi size: %v\n",
				len(receivedDataFinal.Camera), len(receivedDataFinal.Presence), len(receivedDataFinal.Rfid), len(receivedDataFinal.Wifi))
			makePredictions()
			count--
		}()
	} else if txFlag {
		log.Tracef("Received: %v", string(msg.Payload()))
		split := strings.Split(msg.Topic(), "/")
		sensor := split[len(split)-1]
		receivedData.AddNewValue(msg.Payload(), strings.ToLower(sensor))
	}
}

var ff mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	err := json.Unmarshal(msg.Payload(), &txFlag)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	log.Infof("[MQTT] TxFlag updated to ´%v´", txFlag)
}

func init() {
	log.SetLevel(log.DebugLevel)
	txFlag = false
	count = 0

	readConfig()
	viper.SetDefault("mqtt.server", "tcp://127.0.0.1:1883")
	server := viper.GetString("mqtt.server")
	viper.SetDefault("mqtt.clientId", "test-client")
	clientID := viper.GetString("mqtt.clientId")
	viper.SetDefault("mqtt.keepAlive", 2)
	keepAlive := viper.GetInt("mqtt.keepAlive")
	viper.SetDefault("mqtt.pingTimeout", 1)
	pingTimeout := viper.GetInt("mqtt.pingTimeout")

	opts := mqtt.NewClientOptions().AddBroker(server).SetClientID(clientID)
	opts.SetKeepAlive(time.Duration(keepAlive) * time.Second)
	opts.SetPingTimeout(time.Duration(pingTimeout) * time.Second)
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		err := subscribeToTopics()
		if err != nil {
			log.Errorf(err.Error())
			os.Exit(400)
		}
	})

	log.Infof("[MQTT] Connecting to MQTT broker...")
	c = mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		log.Errorf(token.Error().Error())
		os.Exit(400)
	}
	err := generateTrainData()
	if err != nil {
		log.Errorf(err.Error())
		os.Exit(400)
	}
	receivedData = datafusion.CollectData{}
}

func generateTrainData() error {
	viper.SetDefault("ml.trainFile", "")
	trainFile := viper.GetString("ml.trainFile")
	viper.SetDefault("ml.testFile", "")
	testFile := viper.GetString("ml.testFile")
	trainData, err := ml.LoadTrainDataFromCSV(trainFile, testFile)
	if err != nil {
		return err
	}

	trainModel, err = trainData.CreateBestModel()
	if err != nil {
		return err
	}

	log.Debugf("[Init] ModelData: %#v", trainModel)
	return nil
}

func readConfig() {
	cfgFile := "config.toml"
	viper.SetConfigFile(cfgFile)
	if err := viper.ReadInConfig(); err != nil {
		log.Errorf("[Init] Unable to read config from file %s: %s", cfgFile, err.Error())
	} else {
		log.Infof("[Init] Read configuration from file %s", cfgFile)
	}
}

func subscribeToTopics() error {
	log.Infof("[MQTT] Subscribing to MQTT Topic...")
	if token := c.Subscribe("Node/Sensor/+", 0, f); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	if token := c.Subscribe("Node/Flag", 0, ff); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func main() {
	k := 0
	for k < 30 {
		k++
		// Send data for each sensor (random data)
		go auxSendCamera()
		go auxSendPresence()
		go auxSendRfid()
		go auxSendWifi()
		time.Sleep(5 * time.Second)
	}

	// In order to keep the code running. Provisional
	fmt.Scanln()
}

func makePredictions() {
	t1 := time.Now()

	// Calculate the AVG result / list of results from the whole data received from each sensor
	generatedData = datafusion.JoinedData{}
	generatedData.Camera.Sensor = "camera"
	generatedData.GetFinalCameraValues(receivedDataFinal)
	generatedData.Presence.Sensor = "presence"
	generatedData.GetFinalPresenceValues(receivedDataFinal)
	generatedData.Rfid.Sensor = "rfid"
	generatedData.GetFinalRfidValues(receivedDataFinal)
	generatedData.Wifi.Sensor = "camera"
	generatedData.GetFinalWifiValues(receivedDataFinal)

	log.Debugf("[Prediction] CAMERA -> %#v", generatedData.Camera)
	log.Debugf("[Prediction] PRESENCE -> %#v", generatedData.Presence)
	log.Debugf("[Prediction] RFID -> %#v", generatedData.Rfid)
	log.Debugf("[Prediction] WiFi -> %#v", generatedData.Wifi)

	// Obtain a final list with the data to send to the ML algorithm
	predictionDataStruct = datafusion.FinalData{}
	predictionDataStruct.ObtainFinalData(generatedData)

	t2 := time.Now()

	result, _ := json.MarshalIndent(predictionDataStruct, "", "  ")
	log.Infof("[Prediction] %v", string(result))

	log.Infof("[Prediction] Time doing join and calculating final data array: %v", t2.Sub(t1))

	predictionData := predictionDataStruct.To2DFloatArray()
	// var predictionData [][]float64
	// data1 := []float64{76.32, 1.43, 1.43, 21.65, 12.98}
	// data2 := []float64{76.32, 1.43, 71.43, 75.65, 12.98}
	// data3 := []float64{25.32, 0.43, 1.43, 21.65, 35.98}
	// data4 := []float64{28.32, 1.43, 1.43, 21.65, 89.98}
	// predictionData = append(predictionData, data1, data2, data3, data4)
	log.Infof("[Prediction] Obtained 2D Array to predict: %v", predictionData)

	prediction, err := trainModel.MakePrediction(predictionData)
	if err != nil {
		log.Errorf(err.Error())
	}

	log.Infof("[Prediction] Result of prediction: %v", prediction)
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
		if txFlag {
			token := c.Publish("Node/Sensor/Camera", 0, false, byteData)
			if token.Wait() && token.Error() != nil {
				panic(fmt.Sprintf("Error publishing: %v", token.Error()))
			}
		} else {
			log.Warnf("Unable to send Camera data")
		}
		time.Sleep(1 * time.Millisecond)
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
		if txFlag {
			token := c.Publish("Node/Sensor/Camera", 0, false, byteData)
			if token.Wait() && token.Error() != nil {
				panic(fmt.Sprintf("Error publishing: %v", token.Error()))
			}
		} else {
			log.Warnf("Unable to send Camera data")
		}
		time.Sleep(1 * time.Millisecond)
		i++
	}
	for i < 60 {
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
		if txFlag {
			token := c.Publish("Node/Sensor/Camera", 0, false, byteData)
			if token.Wait() && token.Error() != nil {
				panic(fmt.Sprintf("Error publishing: %v", token.Error()))
			}
		} else {
			log.Warnf("Unable to send Camera data")
		}
		time.Sleep(1 * time.Millisecond)
		i++
	}
}

func auxSendPresence() {
	i := 0
	for i < 60 {
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
		if txFlag || (!txFlag && detection) {
			token := c.Publish("Node/Sensor/Presence", 0, false, byteData)
			if token.Wait() && token.Error() != nil {
				panic(fmt.Sprintf("Error publishing: %v", token.Error()))
			}
		} else {
			log.Warnf("Unable to send Presence data")
		}
		time.Sleep(1 * time.Millisecond)
		i++
	}
}

func auxSendRfid() {
	i := 0
	min := 0
	max := 100
	for i < 8 {
		timestamp := 123456789 + i
		power := float64(rand.Intn(max-min) + min)
		data := map[string]interface{}{
			"sensor":    "rfid",
			"timestamp": strconv.Itoa(timestamp),
			"power":     power,
			"person":    7,
		}
		byteData, err := json.Marshal(data)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		if txFlag || (!txFlag && (power >= 60)) {
			token := c.Publish("Node/Sensor/Rfid", 0, false, byteData)
			if token.Wait() && token.Error() != nil {
				panic(fmt.Sprintf("Error publishing: %v", token.Error()))
			}
		} else {
			log.Warnf("Unable to send Rfid data")
		}
		time.Sleep(1 * time.Millisecond)
		i++
	}
	for i < 20 {
		timestamp := 123456789 + i
		power := float64(rand.Intn(max-min) + min)
		data := map[string]interface{}{
			"sensor":    "rfid",
			"timestamp": strconv.Itoa(timestamp),
			"power":     power,
			"person":    5,
		}
		byteData, err := json.Marshal(data)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		if txFlag || (!txFlag && (power >= 60)) {
			token := c.Publish("Node/Sensor/Rfid", 0, false, byteData)
			if token.Wait() && token.Error() != nil {
				panic(fmt.Sprintf("Error publishing: %v", token.Error()))
			}
		} else {
			log.Warnf("Unable to send Rfid data")
		}
		time.Sleep(1 * time.Millisecond)
		i++
	}
	for i < 60 {
		timestamp := 123456789 + i
		power := float64(rand.Intn(max-min) + min)
		data := map[string]interface{}{
			"sensor":    "rfid",
			"timestamp": strconv.Itoa(timestamp),
			"power":     power,
			"person":    6,
		}
		byteData, err := json.Marshal(data)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		if txFlag || (!txFlag && (power >= 60)) {
			token := c.Publish("Node/Sensor/Rfid", 0, false, byteData)
			if token.Wait() && token.Error() != nil {
				panic(fmt.Sprintf("Error publishing: %v", token.Error()))
			}
		} else {
			log.Warnf("Unable to send Rfid data")
		}
		time.Sleep(1 * time.Millisecond)
		i++
	}
}

func auxSendWifi() {
	i := 0
	for i < 20 {
		timestamp := 123456789 + i
		min := 0
		max := 3
		devices := rand.Intn(max-min) + min
		data := map[string]interface{}{
			"sensor":           "wifi",
			"timestamp":        strconv.Itoa(timestamp),
			"connecteddevices": devices,
		}
		byteData, err := json.Marshal(data)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		if txFlag || (!txFlag && (devices >= 2)) {
			token := c.Publish("Node/Sensor/Wifi", 0, false, byteData)
			if token.Wait() && token.Error() != nil {
				panic(fmt.Sprintf("Error publishing: %v", token.Error()))
			}
		} else {
			log.Warnf("Unable to send Wifi data")
		}
		time.Sleep(1 * time.Millisecond)
		i++
	}
}
