package main

import (
	"database/sql"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var db *sql.DB
var mqttClient mqtt.Client

func MQTTConnect() {
	// MQTT connect
	opts := mqtt.NewClientOptions().AddBroker("tcp://localhost:1883").SetClientID("mqtt-to-ws")
	mqttc := mqtt.NewClient(opts)

	if token := mqttc.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("[Ошибка] {Этап: Подключение} MQTT: %v", token.Error())
	}
	mqttClient = mqttc
	log.Println("[Подключено] MQTT")
}

func DBConnect() {
	// DB Connect
	connStr := "user=postgres dbname=haven sslmode=disable"
	dbc, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Fatalln("[Ошибка] {Этап: Подключение} PostgreSQL DB:", err)
	}
	db = dbc
	log.Println("[Подключено] PostgreSQL DB")

}
