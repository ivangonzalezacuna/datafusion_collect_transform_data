package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
)

var (
	txFlagListener mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		err := json.Unmarshal(msg.Payload(), &txFlag)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		log.Infof("[MQTT] TxFlag updated to ´%v´", txFlag)
	}

	c      mqtt.Client
	txFlag bool

	server      string = "tcp://127.0.0.1:1883"
	clientID    string = "sensor-client"
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
	c = mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		log.Errorf(token.Error().Error())
		os.Exit(400)
	}
}

func subscribeToTopics() error {
	log.Infof("[MQTT] Subscribing to MQTT Topic...")
	if token := c.Subscribe("Node/Flag", 0, txFlagListener); token.Wait() && token.Error() != nil {
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
