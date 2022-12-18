package flogger

import (
	"time"
)

type timeType struct {
	Time time.Time
}

/*	Сериализуем время (часы минуты секунды) в строку  */
func (this timeType) MarshalJSON() ([]byte, error) {
	return []byte(this.marshalString()), nil
}

func (this timeType) marshalString() string {
	layout := "15:04:05"
	return "\"" + this.Time.Format(layout) + "\""
}
