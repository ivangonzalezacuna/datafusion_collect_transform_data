package mainprocess

import (
	log "github.com/sirupsen/logrus"
)

type (
	// PredictionDataStruct is the struct with the data to send to the LogisticRegression Model
	PredictionDataStruct struct {
		Timestamp   string  `json:"timestamp"`
		Person      int     `json:"person"`
		Presence    float64 `json:"presence"`
		ConnDevices float64 `json:"conndevices"`
		RfidUser    float64 `json:"rfiduser"`
		RfidPower   float64 `json:"power"`
		CameraUser  float64 `json:"camerauser"`
	}

	//FinalData to test
	FinalData []PredictionDataStruct
)

//ObtainFinalData returns the final list of struct to predict
func (f *FinalData) ObtainFinalData(data JoinedData) {
	finalTimestamp := data.Camera.Timestamp
	if data.Presence.Timestamp < finalTimestamp || finalTimestamp == "" {
		finalTimestamp = data.Presence.Timestamp
	}
	if data.Rfid.Timestamp < finalTimestamp || finalTimestamp == "" {
		finalTimestamp = data.Rfid.Timestamp
	}
	if data.Wifi.Timestamp < finalTimestamp || finalTimestamp == "" {
		finalTimestamp = data.Wifi.Timestamp
	}

	totalCameraCount := 0
	totalRfidCount := 0

	for _, v := range data.Camera.PersonCount {
		totalCameraCount += v.Count
	}
	for _, v := range data.Rfid.PersonCount {
		totalRfidCount += v.Count
	}

	avgCameraUserMap := make(map[int]float64, len(data.Camera.PersonCount))
	avgRfidUserMap := make(map[int]float64, len(data.Rfid.PersonCount))

	for _, v := range data.Camera.PersonCount {
		if v.Count == 0 || totalCameraCount == 0 {
			avgCameraUserMap[v.Person] = float64(0)
		} else {
			avgCameraUserMap[v.Person] = float64(v.Count) / float64(totalCameraCount)
		}
	}

	for _, v := range data.Rfid.PersonCount {
		if v.Count == 0 || totalRfidCount == 0 {
			avgRfidUserMap[v.Person] = float64(0)
		} else {
			avgRfidUserMap[v.Person] = float64(v.Count) / float64(totalRfidCount)
		}
	}

	for k, v := range avgCameraUserMap {
		if !f.isPersonEntryCreated(k) {
			currentData := PredictionDataStruct{
				Timestamp:   finalTimestamp,
				Person:      k,
				Presence:    data.Presence.Detection,
				ConnDevices: data.Wifi.ConnectedDevices,
			}
			if val, ok := avgRfidUserMap[k]; ok {
				currentData.setPersonData(val, v, data)
			} else {
				currentData.setPersonData(0, v, data)
			}
			*f = append(*f, currentData)
		} else {
			log.Debugf("Person %d already saved in struct slice", k)
		}
	}

	for k, v := range avgRfidUserMap {
		if !f.isPersonEntryCreated(k) {
			currentData := PredictionDataStruct{
				Timestamp:   finalTimestamp,
				Person:      k,
				Presence:    data.Presence.Detection,
				ConnDevices: data.Wifi.ConnectedDevices,
			}
			if val, ok := avgCameraUserMap[k]; ok {
				currentData.setPersonData(v, val, data)
			} else {
				currentData.setPersonData(v, 0, data)
			}
			*f = append(*f, currentData)
		} else {
			log.Debugf("Person %d already saved in struct slice", k)
		}
	}
}

func (f *PredictionDataStruct) setPersonData(rfidUser, camUser float64, data JoinedData) {
	f.RfidUser = rfidUser * 100
	f.CameraUser = camUser * 100

	for _, data := range data.Rfid.PersonCount {
		if data.Person == f.Person {
			f.RfidPower = data.Power
		}
	}
	log.Tracef("Current Data info: %#v", f)
}

func (f *FinalData) isPersonEntryCreated(person int) bool {
	for _, v := range *f {
		if person == v.Person {
			return true
		}
	}
	return false
}

// To2DFloatArray converts the array of PredictionDataStruct in a 2D Array of float64
func (f *FinalData) To2DFloatArray() [][]float64 {
	var res [][]float64
	for _, v := range *f {
		// var d1 []float64
		d1 := []float64{v.Presence, v.ConnDevices, v.RfidUser, v.RfidPower, v.CameraUser}
		res = append(res, d1)
	}
	return res
}
