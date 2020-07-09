package tracker

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

type LocationStruct struct {
	Timestamp string
	Person    int
	Location  string
	Alarm     bool
	Rfid      float64
	Wifi      float64
	Counter   int
}

// CheckPermissionsAndStoreEntry checks the permission of a user to be in a room and then decides if the entry is saved in DDBB
func CheckPermissionsAndStoreEntry(info map[string]interface{}, nodeID string, changeCounterRfid, changeCounterWifi float64) {
	var newDetectionData LocationStruct
	if val, ok := info["person"]; !ok {
		newDetectionData.Person = 0
	} else {
		newDetectionData.Person = int(val.(float64))
	}
	if val, ok := info["timestamp"]; ok {
		newDetectionData.Timestamp = val.(string)
	}
	if val, ok := info["rfidpower"]; ok {
		newDetectionData.Rfid = val.(float64)
	}
	if val, ok := info["wifirssi"]; ok {
		newDetectionData.Wifi = val.(float64)
	}
	newDetectionData.Location = nodeID
	newDetectionData.Counter = 0
	log.Infof("Proceeding to check if user %d is allowed to be in the room %s", newDetectionData.Person, newDetectionData.Location)
	// Add check to database. If not allowed, generate alarm
	var allowed bool = true
	newDetectionData.Alarm = allowed
	if !allowed {
		// Generate alarm
	}

	// Get last user detection and compare using counters & roomName. Then decide if store data of not
	err := storeLogInDatabase(newDetectionData)
	if err != nil {
		log.Errorf(err.Error())
	}
}

func storeLogInDatabase(info LocationStruct) error {
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}
	log.Infof("%v", string(data))
	return nil
}
