package flogger

import (
	"errors"
)

type ConfigType struct {
	ServiceName              string `conf:"ServiceName"`
	LogFolder                string `conf:"LogFolder" env:"true"`
	Permissions              string `conf:"Permissions"`
	MaxHoursToChangeLogFile  uint   `conf:"MaxHoursToChangeLogFile" min:"1" max:"24"`
	MaxBufSize               uint   `conf:"MaxBufSize" min:"1"`
	FileWriteDurationSeconds uint   `conf:"FileWriteDurationSeconds" min:"1"`
	WriteChanSize            uint   `conf:"WriteChanSize" min:"1"`
	EnableServiceDebug       bool   `conf:"EnableServiceDebug"`
	EnableBusinessDebug      bool   `conf:"EnableBusinessDebug"`
	EnableQuery              bool   `conf:"EnableQuery"`
	EnableImportant          bool   `conf:"EnableImportant"`
	EnableDecision           bool   `conf:"EnableDecision"`
	EnableFileForQuery       bool   `conf:"EnableFileForQuery"`
	EnableFileForImportant   bool   `conf:"EnableFileForImportant"`
}

/*	Глобальная структура конфига  */
var gConf *ConfigType

/*	Возвращает структуру настроек - без заполнения ее полей пакет считается неинициализированным  */
func GetConfig() *ConfigType {
	if gConf == nil {
		gConf = &ConfigType{}
	}
	return gConf
}

/*	*/
func checkConfig() error {
	if gConf == nil {
		return errors.New("Пакет flogger не сконфигурирован")
	}
	if gConf.LogFolder == "" {
		return errors.New("Параметр LogFolder конфигурации модуля flogger не может быть пустым")
	}
	return nil
}
