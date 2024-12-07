package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var activeClients = make(map[string]*websocket.Conn)
var broadcast = make(chan string)

func main() {
	mqttClient := connectMQTT()
	defer mqttClient.Disconnect(250)

	// Подписка на все темы MQTT
	mqttClient.Subscribe("#", 0, func(client mqtt.Client, msg mqtt.Message) {
		_msg, err := json.Marshal(map[string]interface{}{
			"topic":   msg.Topic(),
			"message": string(msg.Payload()),
		})
		if err == nil {
			broadcast <- string(_msg)
		}
	})

	// WebSocket обработчик
	http.HandleFunc("/", websocketConnect)

	connStr := "user=postgres dbname=postgres sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	var boards []Board

	if err != nil {
		log.Fatal("Ошибка при подключении: ", err)
	}
	rows, err := db.Query("SELECT * FROM boards")
	if err != nil {
		log.Fatal("Ошибка при извлечении данных: ", err)
	}
	defer rows.Close()
	for rows.Next() {
		var brd Board
		if err := rows.Scan(&brd.ID, &brd.Name, &brd.Board_type); err != nil {
			return
		}
		boards = append(boards, brd)
	}
	log.Print(boards)

	go func() {
		for {
			msg := <-broadcast
			for clientIP, client := range activeClients {
				if err := client.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
					log.Printf("Ошибка при отправке: %v", err)
					client.Close()
					delete(activeClients, clientIP)
				}
			}
		}
	}()

	log.Println("Сервер запущен на :1884")
	log.Fatal(http.ListenAndServe(":1884", nil))
}

func websocketConnect(w http.ResponseWriter, r *http.Request) {
	// Получаем IP-адрес клиента
	clientIP := r.RemoteAddr
	if _, exists := activeClients[clientIP]; exists {
		log.Printf("Клиент уже подключен: %s", clientIP)
		http.Error(w, "Клиент уже подключен", http.StatusForbidden)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Ошибка WebSocket апгрейда: %v", err)
		return
	}

	activeClients[clientIP] = ws
	log.Printf("Новое соединение: %s", clientIP)

	defer func() {
		delete(activeClients, clientIP)
		ws.Close()
		log.Printf("Соединение закрыто: %s", clientIP)
	}()

	// Обработка сообщений клиента
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Ошибка чтения WebSocket от %s: %v", clientIP, err)
			break
		}
	}
}

func connectMQTT() mqtt.Client {
	opts := mqtt.NewClientOptions().AddBroker("tcp://localhost:1883").SetClientID("mqtt-to-ws")
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Ошибка подключения к MQTT: %v", token.Error())
	}
	log.Println("Подключено к MQTT брокеру")
	return client
}
