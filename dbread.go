package main

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/lib/pq"
)

func interfaceToData(data interface{}) (string, error) {
	j, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(j), nil
}

func QueryToStruct(query string, args []interface{}, destFunc func(rows *sql.Rows) error) error {
	rows, err := db.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return destFunc(rows)
}

func FetchSensors() ([]Sensor, error) {
	var sensors []Sensor
	query := `SELECT s.id, s.sensor_name, s.data_type, b.board_name, s.sensor_group, s.sensor_value 
	          FROM sensors s 
	          JOIN boards b ON s.board_id = b.id;`

	err := QueryToStruct(query, nil, func(rows *sql.Rows) error {
		for rows.Next() {
			var snsr Sensor
			if err := rows.Scan(&snsr.ID, &snsr.Sensor_name, &snsr.Data_type, &snsr.Board_name, &snsr.Sensor_group, &snsr.Sensor_value); err != nil {
				return err
			}
			sensors = append(sensors, snsr)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return sensors, nil
}

func FetchSensor(ID string) (*Sensor, error) {
	query := fmt.Sprintf(`SELECT s.id, s.sensor_name, s.data_type, b.board_name, s.sensor_group, s.sensor_value 
				FROM sensors s 
				JOIN boards b ON s.board_id = b.id
				WHERE s.id = %s;`, ID)

	var snsr Sensor

	err := QueryToStruct(query, nil, func(rows *sql.Rows) error {
		for rows.Next() {
			if err := rows.Scan(&snsr.ID, &snsr.Sensor_name, &snsr.Data_type, &snsr.Board_name, &snsr.Sensor_group, &snsr.Sensor_value); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &snsr, nil
}
