package mainprocess

import (
	log "github.com/sirupsen/logrus"
)

type (
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

	// JoinedData stores the final data fusion from the sensors
	JoinedData struct {
		Camera   cameraStructFinal
		Presence presenceStructFinal
		Rfid     rfidStructFinal
		Wifi     wifiStructFinal
	}
)

//GetFinalCameraValues obtains one struct using all the received data from the camera
func (g *JoinedData) GetFinalCameraValues(data CollectData) error {
	if len(data.Camera) == 0 {
		return nil
	}
	g.Camera.Timestamp = data.Camera[0].Timestamp

	peopleCount := make(map[int]int)
	for _, v := range data.Camera {
		if _, exist := peopleCount[v.Person]; exist {
			peopleCount[v.Person]++
		} else {
			peopleCount[v.Person] = 1
		}
	}

	for k, v := range peopleCount {
		currentCount := cameraStructCountFinal{Person: k, Count: v}
		g.Camera.PersonCount = append(g.Camera.PersonCount, currentCount)
		log.Debugf("CAMERA -> Person : %d , Count : %d\n", k, v)
	}

	return nil
}

//GetFinalPresenceValues obtains one struct using all the received data from the presence detector
func (g *JoinedData) GetFinalPresenceValues(data CollectData) error {
	// generatedData.Presence.Sensor = "presence"
	if len(data.Presence) == 0 {
		return nil
	}
	g.Presence.Timestamp = data.Presence[0].Timestamp

	var positiveEntries int = 0
	for _, v := range data.Presence {
		if v.Detection {
			positiveEntries++
		}
	}
	detectionAvg := (float64(positiveEntries) / float64(len(data.Presence))) * 100
	log.Debugf("PRESENCE AVG: %.2f", detectionAvg)
	g.Presence.Detection = detectionAvg

	return nil
}

//GetFinalRfidValues obtains one struct using all the received data from the rfid reader
func (g *JoinedData) GetFinalRfidValues(data CollectData) error {
	// generatedData.Rfid.Sensor = "rfid"
	if len(data.Rfid) == 0 {
		return nil
	}
	g.Rfid.Timestamp = data.Rfid[0].Timestamp

	peopleCount := make(map[int]struct {
		count int
		total float64
	})
	for _, v := range data.Rfid {
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
		g.Rfid.PersonCount = append(g.Rfid.PersonCount, rfidStructCountFinal{Person: k, Count: v.count, Power: powerAvg})
		log.Debugf("RFID -> Person: %d , Data: %v\n", k, v)
	}

	return nil
}

//GetFinalWifiValues obtains one struct using all the received data from the wifi
func (g *JoinedData) GetFinalWifiValues(data CollectData) error {
	// generatedData.Wifi.Sensor = "wifi"
	if len(data.Wifi) == 0 {
		return nil
	}
	g.Wifi.Timestamp = data.Wifi[0].Timestamp

	var totalConnDevices int = 0
	for _, v := range data.Wifi {
		totalConnDevices += v.ConnectedDevices
	}
	devicesAvg := float64(totalConnDevices) / float64(len(data.Wifi))
	log.Debugf("WiFi AVG: %f", devicesAvg)
	g.Wifi.ConnectedDevices = devicesAvg

	return nil
}