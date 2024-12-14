package main

//Здесь должно быть описано всё, что связано с webScoket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/websocket"
)

var activeClients = make(map[string]*websocket.Conn)
var broadcast = make(chan string)

func main() {

	MQTTConnect()
	DBConnect()

	defer mqttClient.Disconnect(250)
	defer db.Close()

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
					log.Printf("[Ошибка] {Этап: Отправка сообщения} Web Socket:%v", err)
					client.Close()
					delete(activeClients, clientIP)
				}
			}
		}
	}()

	// service, err := zeroconf.Register(
	// 	"haven",           // Имя устройства
	// 	"_http.webScoket", // Тип сервиса
	// 	".local",          // Домейн mDNS
	// 	8080,              // Порт
	// 	nil,               // Нет текстовых записей
	// 	nil,               // Интерфейс выбирается автоматически
	// )
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer service.Shutdown()

	// // Держим процесс живым
	// select {}

	log.Println("[Подключено] Сервер запущен на порту 1884")
	log.Fatal(http.ListenAndServe(":1884", nil))
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func webScoketSendMessage(data interface{}, ws *websocket.Conn, clientIP string, msgType string) error {

	jsonData, err := interfaceToData(data)
	if err != nil {
		return fmt.Errorf("[Ошибка] {Этап: Маршалинг (1/2) JSON}: %v", err)
	}

	wsmsg, err := json.Marshal(map[string]interface{}{
		"type": msgType,
		"data": jsonData,
	})

	if err != nil {
		return fmt.Errorf("[Ошибка] {Этап: Маршалинг (2/2) JSON}: %v", err)
	}

	if err := ws.WriteMessage(websocket.TextMessage, []byte(wsmsg)); err != nil {
		ws.Close()
		delete(activeClients, clientIP)
		return fmt.Errorf("[Ошибка] {Этап: Отправка сообщения} Web Socket: %w", err)
	}

	return nil
}

func websocketConnect(w http.ResponseWriter, r *http.Request) {
	// Получаем IP-адрес клиента
	clientIP := r.RemoteAddr
	if _, exists := activeClients[clientIP]; exists {
		log.Printf("[Ошибка] {Клиент уже подключен} Web Socket:%s", clientIP)
		http.Error(w, "Клиент уже подключен", http.StatusForbidden)
		return
	}
	// Апгрейдим подключение ws
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[Ошибка] {Этап: Обновление подключения} Web Socket:%v", err)
		return
	}

	// Отрабатываем подключение новго юзера
	activeClients[clientIP] = ws
	log.Printf("[Подключено] Новый клиент:%s", clientIP)

	defer func() {
		delete(activeClients, clientIP)
		ws.Close()
		log.Printf("[Отключено] Клиент %s отключен.", clientIP)
	}()

	// Обработка сообщений клиента
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Printf("[Ошибка] {Этап: Чтение сообщения от %s} Web Socket:%v", clientIP, err)
			break
		}

		var message Message
		err = json.Unmarshal(msg, &message)
		if err != nil {
			log.Println(string(msg))
			log.Println("[Ошибка] {Этап: Парсинг JSON} JSON:", err)
			continue
		}

		switch string(message.Type) {
		case "refresh-all":
			sensors, err := FetchSensors()
			if err != nil {
				log.Fatalf("[Ошибка] {Этап: Получение данных из DB}: %v", err)
			}

			if err := webScoketSendMessage(sensors, ws, clientIP, message.Type); err != nil {
				log.Fatalf(err.Error())
			}

		case "refresh-sensor":
			sensor, err := FetchSensor(message.Data)
			if err != nil {
				log.Fatalf("[Ошибка] {Этап: Получение данных из DB}: %v", err)
			}

			if err := webScoketSendMessage(sensor, ws, clientIP, message.Type); err != nil {
				log.Fatalf(err.Error())
			}

		case "echo":
			log.Println("Получен echo, отправляю обратно:")
		default:
			log.Println("Неизвестный тип сообщения:")
		}
	}
}
