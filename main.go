package main

//Здесь должно быть описано всё, что связано с webScoket

import (
	"encoding/json"
	"log"
	"net/http"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/websocket"
)

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

	// Goroutine которая отправляет все сообщения пользователям ws
	// Скорее всего удалиться
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

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func websocketConnect(w http.ResponseWriter, r *http.Request) {
	// Получаем IP-адрес клиента
	clientIP := r.RemoteAddr
	if _, exists := activeClients[clientIP]; exists {
		log.Printf("Клиент уже подключен: %s", clientIP)
		http.Error(w, "Клиент уже подключен", http.StatusForbidden)
		return
	}
	// Апгрейдим подключение ws
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Ошибка WebSocket апгрейда: %v", err)
		return
	}

	// Отрабатываем подключение новго юзера
	activeClients[clientIP] = ws
	log.Printf("Новое соединение: %s", clientIP)

	defer func() {
		delete(activeClients, clientIP)
		ws.Close()
		log.Printf("Соединение закрыто: %s", clientIP)
	}()

	// Обработка сообщений клиента
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Ошибка чтения WebSocket от %s: %v", clientIP, err)
			break
		}

		var message Message
		err = json.Unmarshal(msg, &message)
		if err != nil {
			log.Println(string(msg))
			log.Println("Ошибка парсинга JSON:", err)
			continue
		}

		switch string(message.Type) {
		case "sensors-refresh":
			if err := ws.WriteMessage(websocket.TextMessage, []byte(marshalData())); err != nil {
				log.Printf("Ошибка при отправке: %v", err)
				ws.Close()
				delete(activeClients, clientIP)
			}
		case "echo":
			log.Println("Получен echo, отправляю обратно:")
		default:
			log.Println("Неизвестный тип сообщения:")
		}
	}
}
