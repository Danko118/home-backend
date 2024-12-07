package main

import (
	"database/sql"
	"encoding/json"
	"log"

	_ "github.com/lib/pq"
)

// func readBoards() []Board {
// 	connStr := "user=postgres dbname=postgres sslmode=disable"
// 	db, err := sql.Open("postgres", connStr)

// 	var boards []Board

// 	if err != nil {
// 		log.Fatal("Ошибка при подключении: ", err)
// 	}
// 	rows, err := db.Query("SELECT * FROM boards")
// 	if err != nil {
// 		log.Fatal("Ошибка при извлечении данных: ", err)
// 	}
// 	defer rows.Close()
// 	for rows.Next() {
// 		var brd Board
// 		if err := rows.Scan(&brd.ID, &brd.Name, &brd.Board_type); err != nil {
// 			log.Fatal("Ошибка при извлечении данных: ", err)
// 		}
// 		boards = append(boards, brd)
// 	}
// 	return boards
// }

func marshalData() string {
	j, err := json.Marshal(readSensors())
	if err != nil {
		log.Fatal("Ошибка при JSON Parse:", err)
	}
	return string(j)

}

func readSensors() []Sensor {
	connStr := "user=postgres dbname=haven sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	var sensors []Sensor

	if err != nil {
		log.Fatal("Ошибка при подключении: ", err)
	}
	rows, err := db.Query("SELECT s.id AS sensor_id, s.sensor_name AS sensor_name, s.data_type AS data_type, b.board_name AS board_name, s.sensor_group AS sensor_group, s.sensor_value AS sensor_value FROM sensors s JOIN boards b ON s.board_id = b.id;")
	if err != nil {
		log.Fatal("Ошибка при извлечении данных: ", err)
	}
	defer rows.Close()
	for rows.Next() {
		var snsr Sensor
		if err := rows.Scan(&snsr.ID, &snsr.Sensor_name, &snsr.Data_type, &snsr.Board_name, &snsr.Sensor_group, &snsr.Sensor_value); err != nil {
			log.Fatal("Ошибка при извлечении данных: ", err)
		}
		sensors = append(sensors, snsr)
	}
	return sensors
}
