package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path"
	"strings"
	"time"

	datafusion "mainprocess/datafusion"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	ml "github.com/ivangonzalezacuna/ml_regression_tracking"
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

	// Topic names used in the system
	topicSensor = "/Nodes/Node_ID/Tracking/Sensor/+"
	topicTxFlag = "/Nodes/Node_ID/Tracking/TxFlag"
	toolTopic   = "/Nodes/Node_%v/Tracking/Detection"
)

var sensorDataListener mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	if !txFlag && count == 0 && trainModel.Model != nil {
		count++
		txFlag = true
		byteData, err := json.Marshal(txFlag)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		token := mqttClient.Publish(topicTxFlag, 0, false, byteData)
		if token.Wait() && token.Error() != nil {
			log.Errorf(fmt.Sprintf("Error publishing: %v", token.Error()))
		}

		go func() {
			viper.SetDefault("ml.window", 350)
			sleepTime := viper.GetInt("ml.window")
			viper.Set("ml.window", sleepTime)
			viper.WriteConfig()
			time.Sleep(time.Duration(sleepTime) * time.Millisecond)
			txFlag = false
			byteData, err := json.Marshal(txFlag)
			if err != nil {
				log.Errorf(err.Error())
				return
			}

			log.Debugf("[MQTT] Deactivating flag after %dms!", sleepTime)
			token := mqttClient.Publish(topicTxFlag, 0, false, byteData)
			if token.Wait() && token.Error() != nil {
				log.Errorf(fmt.Sprintf("Error publishing: %v", token.Error()))
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

func init() {
	log.SetLevel(log.DebugLevel)
	txFlag = false
	count = 0

	readConfig()
	viper.SetDefault("mqtt.server", "tcp://127.0.0.1:1883")
	server := viper.GetString("mqtt.server")
	viper.Set("mqtt.server", server)
	viper.SetDefault("mqtt.mainclientid", "main-client")
	clientID := viper.GetString("mqtt.mainclientid")
	viper.Set("mqtt.mainclientid", clientID)
	viper.SetDefault("mqtt.keepAlive", 10)
	keepAlive := viper.GetInt("mqtt.keepAlive")
	viper.Set("mqtt.keepAlive", keepAlive)
	viper.SetDefault("mqtt.pingTimeout", 1)
	pingTimeout := viper.GetInt("mqtt.pingTimeout")
	viper.Set("mqtt.pingTimeout", pingTimeout)
	viper.WriteConfig()

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
	viper.SetDefault("ml.trainFile", "./data/train.csv")
	trainFile := viper.GetString("ml.trainFile")
	viper.Set("ml.trainFile", trainFile)
	viper.SetDefault("ml.testFile", "./data/test.csv")
	testFile := viper.GetString("ml.testFile")
	viper.Set("ml.testFile", testFile)
	viper.WriteConfig()
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
	userDir, err := user.Current()
	if err != nil {
		log.Errorf(err.Error())
	}

	configDir := path.Join(userDir.HomeDir, ".config", "ml-system")
	_, err = os.Stat(configDir)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(configDir, 0755)
		if errDir != nil {
			log.Errorf(err.Error())
		}
	}

	cfgFileDir := path.Join(configDir, "config.toml")
	file, err := os.OpenFile("access.log", os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		log.Errorf(err.Error())
	}
	err = file.Close()
	if err != nil {
		log.Errorf(err.Error())
	}
	viper.SetConfigFile(cfgFileDir)
	if err := viper.ReadInConfig(); err != nil {
		log.Errorf("[Init] Unable to read config from file %s: %s", cfgFileDir, err.Error())
	} else {
		log.Infof("[Init] Read configuration from file %s", cfgFileDir)
	}
}

func subscribeToTopics() error {
	log.Infof("[MQTT] Subscribing to MQTT Topic...")
	if token := mqttClient.Subscribe(topicSensor, 0, sensorDataListener); token.Wait() && token.Error() != nil {
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

	if len(prediction) != len(predictionDataStruct) {
		return fmt.Errorf("Prediction results sizes mismatch")
	}
	for k := range predictionDataStruct {
		if prediction[k] == 1 {
			predictionDataStruct[k].Detection = true
			byteData, err := json.Marshal(predictionDataStruct[k])
			if err != nil {
				return err
			}
			topic := fmt.Sprintf(toolTopic, predictionDataStruct[k].Person)
			token := mqttClient.Publish(topic, 0, false, byteData)
			if token.Wait() && token.Error() != nil {
				log.Errorf(fmt.Sprintf("Error publishing: %v", token.Error()))
			}
		}
	}

	return nil
}
