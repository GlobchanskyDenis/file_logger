package flogger

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type messageType struct {
	Timestamp int64      `json:"stamp"`           // Это поле для обработки программой автоматического чтения логов (чтобы не приходилось парсить время)
	Time      timeType   `json:"time"`            // Это человекочитаемое время (без даты, дата в любом случае отображается в названии файла логгирования)
	LogLevel  string     `json:"level"`           // Error / Info / Debug / Warning...
	Error     *errorType `json:"error,omitempty"` // Код, тип (Internal / External / Request / Business) и тело ошибки
	Fields    map[string]interface{}
	Message   string `json:"message"`
}

/*	Сериализуем сообщение компактнее чем позволяет стандартный маршаллер
**	(убираем лишние пробелы между полями (как и в jsonb), параметр Args без рефлексии
**	превращается в поля текущей структуры и сортируется по алфавиту)
**	Соблюдается совместимость с форматом json (но из-за особенностей сериализации параметра Args
**	данная dto недействительна при Unmarshal (нужно будет анмаршаллить в мапу))  */
func (this messageType) MarshalJSON() ([]byte, error) {

	timepart := this.Time.marshalString()
	var errPart string
	if this.Error != nil {
		jsonB, _ := json.Marshal(this.Error.Message)
		if this.Error.Code != 0 && this.Error.Type != "" {
			errPart = "{\"code\":" + strconv.FormatUint(uint64(this.Error.Code), 10) + ",\"type\":\"" + this.Error.Type + "\",\"message\":" + string(jsonB) + "}"
		} else {
			errPart = "{\"message\":" + string(jsonB) + "}"
		}
	}

	var dst []byte
	dst = append(dst, []byte("{\"stamp\":"+strconv.FormatUint(uint64(this.Timestamp), 10)+
		",\"time\":"+timepart+",\"level\":\""+this.LogLevel+"\",")...)
	if this.Error != nil {
		dst = append(dst, []byte("\"error\":"+errPart+",")...)
	}

	dst = append(dst, convertFields(this.Fields)...)
	dst = append(dst, []byte("\"message\":\""+this.Message+"\"}\n")...)
	return dst, nil
}

func convertFields(fieldsMap map[string]interface{}) []byte {
	if len(fieldsMap) == 0 {
		return nil
	}

	/*	Заполняем слайс ключами  */
	var keyList = make([]string, len(fieldsMap))
	var i = 0
	for key, _ := range fieldsMap {
		keyList[i] = key
		i++
	}

	/*	Сортируем ключи в слайсе по алфавиту улучшенным баблсортом (количество аргументов все равно небольшое)  */
	for {
		var wasSwapped bool = false
		for i := 0; i < len(keyList)-1; i++ {
			if isNeedToSwap(keyList[i], keyList[i+1]) {
				var temp = keyList[i]
				keyList[i] = keyList[i+1]
				keyList[i+1] = temp
				wasSwapped = true
			}
		}
		if wasSwapped == false {
			break
		}
	}

	/*	Конвертируем содержимое мапы в порядке который мы получили после сортировки ключей  */
	var dst []byte
	for _, key := range keyList {
		dst = append(dst, []byte("\""+key+"\":")...)
		dst = append(dst, []byte(stringify(fieldsMap[key]))...)
		dst = append(dst, byte(','))
	}
	return dst
}

/*	Сравнивает два ключа по алфавиту и говорит нужно ли свапнуть ключи при заполнении из мапы  */
func isNeedToSwap(key1, key2 string) bool {
	/*	Защита от дурака  */
	if len(key1) == 0 {
		return false
	}
	if len(key2) == 0 {
		return true
	}

	var minLen = len(key1)
	if minLen > len(key2) {
		minLen = len(key2)
	}
	for i := 0; i < minLen; i++ {
		if key1[i] > key2[i] {
			return true
		} else if key1[i] < key2[i] {
			return false
		}
	}
	if len(key1) > len(key2) {
		return true
	}
	return false
}

func stringify(src interface{}) string {
	switch typed := src.(type) {
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case int32:
		return strconv.FormatInt(int64(typed), 10)
	case uint:
		return strconv.FormatUint(uint64(typed), 10)
	case uint32:
		return strconv.FormatUint(uint64(typed), 10)
	case uint64:
		return strconv.FormatUint(typed, 10)
	case float64:
		return strconv.FormatFloat(typed, 'E', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(typed), 'E', -1, 32)
	case bool:
		return strconv.FormatBool(typed)
	case map[string]interface{}:
		if typed == nil {
			return "\"map[string]interface{} nil\""
		}
		var dst string = "{"
		var i int = 0
		for key, value := range typed {
			i++
			dst += "\"" + key + "\":" + stringify(value)
			if i != len(typed) {
				dst += ","
			}
		}
		return dst + "}"
	case []interface{}:
		var dst string = "["
		for i, value := range typed {
			dst += stringify(value)
			if i < len(typed)-1 {
				dst += ","
			}
		}
		return dst + "]"
	case string:
		return "\"" + typed + "\""
	default:
		return fmt.Sprintf("\"%T\"", src)
	}
}

func convertBufToBite(cpyBuf []messageType) []byte {
	var dst []byte
	for _, message := range cpyBuf {
		jsonB, err := message.MarshalJSON()
		if err != nil {
			println("не смог сериализовать лог " + err.Error())
		} else {
			dst = append(dst, jsonB...)
		}
	}
	return dst
}
