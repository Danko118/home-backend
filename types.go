package main

type Message struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

type Board struct {
	ID         int64
	Board_Name string
	Board_type string
}

type Sensor struct {
	ID           int64  `json:"id"`
	Sensor_name  string `json:"name"`
	Data_type    string `json:"type"`
	Sensor_group string `json:"group"`
	Sensor_value string `json:"value"`
	Board_name   string `json:"board_name"`
}
