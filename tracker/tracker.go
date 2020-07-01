package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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
		if len(topicList) == 5 {
			go checkPermissions(data, topicList[2])
		}
	}

	mqttClient mqtt.Client
	txFlag     bool

	toolTopic = "/Nodes/+/Tracking/Detection"
)

func init() {
	log.SetLevel(log.DebugLevel)
	txFlag = false

	readConfig()
	viper.SetDefault("mqtt.server", "tcp://127.0.0.1:1883")
	server := viper.GetString("mqtt.server")
	viper.SetDefault("mqtt.accessclientid", "tool-control-client")
	clientID := viper.GetString("mqtt.accessclientid")
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
	_, err = os.OpenFile("access.log", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
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
	if token := mqttClient.Subscribe(toolTopic, 0, newDetectionListener); token.Wait() && token.Error() != nil {
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
