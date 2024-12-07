package main

import (
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func connectMQTT() mqtt.Client {
	opts := mqtt.NewClientOptions().AddBroker("tcp://localhost:1883").SetClientID("mqtt-to-ws")
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Ошибка подключения к MQTT: %v", token.Error())
	}
	log.Println("Подключено к MQTT брокеру")
	return client
}
