# Data Fusion Process

Code used to obtain data from different sensors, combine them using simple data fusion methods and predict if a person is neither detected or not using a ML Algorithm called Logistic Regression. If a detection is predicted, the process will analyze if the person has permission rights to be in that room and generate an alarm if not. Finally, if will store in a database the logs of each detection.

## Directories

* **`data`**. Contains the CSV files (*currently in progress*) to train the Logistic Regression Model.
* **`datafusion`**. Contains functions and data structures for different data fusion steps:
  * `collect_data.go`. All the related structures and functions to collect data from the different sensors.
  * `joined_data.go`. All the related structures and functions to join the array of data collected from each sensor. Obtaining a single entry for each sensor
  * `fusion_data.go`. All the related structures and functions to join the data of each sensor. Obtaining an array of entries (one for each different person detected). 
* **`sensor`**. Auxiliar code to generate random data from each sensor.
* **`tracker`**. Contains a function that will check the permission rights of one person to be in a defined room, generate alarms if needed and store logs in a database.

## Dependencies

* **Machine Learning Algorithm**. Using my personal repo: [ml_regression_tracking](https://github.com/ivangonzalezacuna/ml_regression_tracking)
* **MQTT Broker**. In this project I use [Mosquitto](https://mosquitto.org/). You can follow the instructions in [here](https://medium.com/@aegkaluk/install-mqtt-broker-on-ubuntu-18-04-15232ab0ee42) to download and secure it. If an error occurs when running the code saying that it's not possible to connect to the broker. Execute this in the terminal:
```bash
sudo systemctl stop mosquitto.service  // To stop the mosquitto service which is not working properly
mosquitto -p 1883 -v // To execute the broker manually on port 1883 with the verbose flag activated
```

## Execute code

Terminal 1
```bash
git clone https://github.com/ivangonzalezacuna/datafusion_collect_transform_data.git
cd datafusion_collect_transform_data
go build -v
./mainprocess
```

Now, open other 2 terminals in the same directory. And execute these lines:

Terminal 2
```bash
cd ~/datafusion_collect_transform_data/sensor
go build sensor.go
./sensor
```

Terminal 3
```bash
cd ~/datafusion_collect_transform_data/tracker
go build tracker.go
./tracker
```
