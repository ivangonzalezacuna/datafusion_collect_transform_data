package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
)

var (
	newDetectionListener mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		var data map[string]interface{}
		err := json.Unmarshal(msg.Payload(), &data)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		topicList := strings.Split(msg.Topic(), "/")
		if len(topicList) == 4 {
			go checkPermissions(data, topicList[1])
		}
	}

	mqttClient mqtt.Client
	txFlag     bool

	server      string = "tcp://127.0.0.1:1883"
	clientID    string = "database-client"
	keepAlive   int    = 10
	pingTimeout int    = 1
)

func init() {
	log.SetLevel(log.DebugLevel)
	txFlag = false

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
}

func subscribeToTopics() error {
	log.Infof("[MQTT] Subscribing to MQTT Topic...")
	if token := mqttClient.Subscribe("Node/+/Tracking/Detection", 0, newDetectionListener); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func main() {
	// In order to keep the code running. Provisional
	fmt.Scanln()
}

func checkPermissions(info map[string]interface{}, nodeID string) {
	var person int
	if val, ok := info["person"]; !ok {
		person = 0
	} else {
		person = int(val.(float64))
	}
	info["location"] = nodeID
	log.Infof("Proceeding to check if user %d is allowed to be in the room %s", person, info["location"])
	// Add check to database. If not allowed, generate alarm  and then save the log. If allowed, save the log
	var allowed bool = true
	info["access"] = allowed
	if !allowed {
		// Generate alarm
	}
	// Store in logging database
	err := storeLogInDatabase(info)
	if err != nil {
		log.Errorf(err.Error())
	}
}

func storeLogInDatabase(info map[string]interface{}) error {
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}
	log.Infof("%v", string(data))
	return nil
}
