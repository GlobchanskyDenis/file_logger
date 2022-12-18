package utils

import (
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func strFilePemToUint(parseString string) (uint32, error) {
	ui32, err := strconv.ParseUint(parseString, 8, 32)
	if err != nil {
		return 0, err
	}
	return uint32(ui32), nil
}

func OpenOrCreateNewFile(pathToFile, fileName, logFileCreatePem string) (*os.File, error) {
	FilePem, err := strFilePemToUint(logFileCreatePem)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(pathToFile); err != nil {
		err := os.MkdirAll(pathToFile, os.FileMode(FilePem))
		if err != nil {
			return nil, err
		}
	}
	f, err := os.OpenFile(filepath.Join(pathToFile, fileName), os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.FileMode(FilePem))
	if err != nil {
		return nil, err
	}
	return f, nil
}

/*	Дата должна совпасть полностью, а час - с учетом настроек конфигурации  */
func IsSameDateAndHour(oldDate, now time.Time, maxHours uint) bool {
	if oldDate.Year() == now.Year() && oldDate.Month() == now.Month() && oldDate.Day() == now.Day() {
		if CalcHour(oldDate.Hour(), int(maxHours)) != CalcHour(now.Hour(), int(maxHours)) {
			return false
		}
		return true
	}
	return false
}

func CalcHour(currentHour, maxHoursToChangeLogFile int) int {
	var hour int
	for i := 1; i*maxHoursToChangeLogFile <= currentHour; i++ {
		hour = i * maxHoursToChangeLogFile
	}
	return hour
}
