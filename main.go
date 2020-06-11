package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	datafusion "mainprocess/datafusion"
	ml "mainprocess/ml_regression_tracking"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	// Array that stores the received data from each sensor
	receivedData datafusion.CollectData
	// Copy of ´receivedData´, done when the txFlag is deactivated, so that collecting new data will be possible while doing the prediction
	receivedDataFinal datafusion.CollectData
	// Struct generated from the received data of each sensor
	generatedData datafusion.JoinedData
	// Array of data for each detected user by the camera and/or rfid reader
	predictionDataStruct datafusion.FinalData
	// Struct to store the train data for the Logistic Regression
	trainData ml.TrainData
	// Structrained Logistic Regression Model and other related data
	trainModel ml.ModelData
	// MQTT Client
	mqttClient mqtt.Client
	// txFlag is a global variable that allows or denies the transmission of data from each sensor
	txFlag bool
	// count is used to allow only one thread to activate the txFlag and deactivate it after some time
	count int
)

var sensorDataListener mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	if !txFlag && count == 0 && trainModel.Model != nil {
		count++
		byteData, err := json.Marshal(true)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		token := mqttClient.Publish("Node/Flag", 0, false, byteData)
		if token.Wait() && token.Error() != nil {
			panic(fmt.Sprintf("Error publishing: %v", token.Error()))
		}

		go func() {
			viper.SetDefault("ml.window", 500)
			sleepTime := viper.GetInt("ml.window")
			time.Sleep(time.Duration(sleepTime) * time.Millisecond)
			byteData, err := json.Marshal(false)
			if err != nil {
				log.Errorf(err.Error())
				return
			}

			log.Debugf("[MQTT] Deactivating flag after %dms!", sleepTime)
			token := mqttClient.Publish("Node/Flag", 0, false, byteData)
			if token.Wait() && token.Error() != nil {
				panic(fmt.Sprintf("Error publishing: %v", token.Error()))
			}

			receivedDataFinal = receivedData
			log.Infof("[MQTT] Camera size: %v\nPresence size: %v\nRfid size: %v\nWifi size: %v\n",
				len(receivedDataFinal.Camera), len(receivedDataFinal.Presence), len(receivedDataFinal.Rfid), len(receivedDataFinal.Wifi))
			err = makePredictions()
			if err != nil {
				log.Errorf(err.Error())
			}
			count--
		}()
	} else if txFlag {
		log.Tracef("Received: %v", string(msg.Payload()))
		split := strings.Split(msg.Topic(), "/")
		sensor := split[len(split)-1]
		err := receivedData.AddNewValue(msg.Payload(), strings.ToLower(sensor))
		if err != nil {
			log.Errorf(err.Error())
		}
	}
}

var txFlagListener mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
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
	mqttClient = mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
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
	if token := mqttClient.Subscribe("Node/Sensor/+", 0, sensorDataListener); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	if token := mqttClient.Subscribe("Node/Flag", 0, txFlagListener); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func main() {
	// In order to keep the code running. Provisional
	fmt.Scanln()
}

func makePredictions() error {
	t1 := time.Now()

	// Calculate the AVG result / list of results from the whole data received from each sensor
	generatedData = datafusion.JoinedData{}
	err := generatedData.GetFinalValues(receivedDataFinal)
	if err != nil {
		return fmt.Errorf("Can't make prediction: %v", err.Error())
	}

	log.Debugf("[Prediction] CAMERA -> %#v", generatedData.Camera)
	log.Debugf("[Prediction] PRESENCE -> %#v", generatedData.Presence)
	log.Debugf("[Prediction] RFID -> %#v", generatedData.Rfid)
	log.Debugf("[Prediction] WiFi -> %#v", generatedData.Wifi)

	// Obtain a final list with the data to send to the ML algorithm
	predictionDataStruct = datafusion.FinalData{}
	predictionDataStruct.ObtainFinalData(generatedData)

	result, err := json.MarshalIndent(predictionDataStruct, "", "  ")
	if err != nil {
		return err
	}
	log.Debugf("[Prediction] %v", string(result))

	predictionData := predictionDataStruct.To2DFloatArray()
	log.Debugf("[Prediction] Obtained 2D Array to predict: %v", predictionData)

	prediction, err := trainModel.MakePrediction(predictionData)
	if err != nil {
		return err
	}
	log.Infof("[Prediction] Result of prediction: %v", prediction)

	t2 := time.Now()
	log.Debugf("[Prediction] Time doing join and calculating final data array: %v", t2.Sub(t1))

	return nil
}
