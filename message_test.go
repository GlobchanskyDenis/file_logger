package flogger

import (
	"encoding/json"
	"testing"
	"time"
)

/*	Бэнчмарки показывают что кастомная маршаллизация медленнее дефолтной из-за большего количества аллокаций
**	Тем не менее это допустимо так как количество аргументов в мапе всегда будет небольшим.
**	Оптимизация невозможна так как пришлось бы отказаться от сортировки аргументов по алфавиту, а это ОЧЕНЬ
**	полезное свойство. Причина аллокаций - аргументы подаются
 */

/*	go test -bench . -benchmem  */
func BenchmarkMessageMarshalCustom(b *testing.B) {
	var dto = messageType{
		Timestamp: 100500,
		Time:      timeType{Time: time.Now()},
		LogLevel:  "Warning",
		Error: &errorType{
			Code:    42,
			Type:    "Business",
			Message: "cant do something: \"becouse of...,\"",
		},
		Fields: map[string]interface{}{
			"worker":  1,
			"arg1asd": "asddsadasd",
			"arg1":    "asds",
			"arg2":    "fdsjkfhdsfjkh",
		},
		Message: "while something",
	}
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(dto); err != nil {
			b.Errorf("Error: %s", err)
			b.FailNow()
		}
	}
}

/*	go test -bench . -benchmem  */
func BenchmarkMessageMarshalStandart(b *testing.B) {
	type alternateMessage struct {
		Timestamp uint       `json:"stamp"`           // Это поле для обработки программой автоматического чтения логов (чтобы не приходилось парсить время)
		Time      timeType   `json:"time"`            // Это человекочитаемое время (без даты, дата в любом случае отображается в названии файла логгирования)
		LogLevel  string     `json:"level"`           // Error / Info / Debug / Warning...
		Error     *errorType `json:"error,omitempty"` // Код, тип (Internal / External / Request / Business) и тело ошибки
		Fields    map[string]interface{}
		Message   string `json:"message"`
	}
	var dto = alternateMessage{
		Timestamp: 100500,
		Time:      timeType{Time: time.Now()},
		LogLevel:  "Warning",
		Error: &errorType{
			Code:    42,
			Type:    "Business",
			Message: "cant do something: \"becouse of...,\"",
		},
		Fields: map[string]interface{}{
			"worker":  1,
			"arg1asd": "asddsadasd",
			"arg1":    "asds",
			"arg2":    "fdsjkfhdsfjkh",
		},
		Message: "while something",
	}
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(dto); err != nil {
			b.Errorf("Error: %s", err)
			b.FailNow()
		}
	}
}

func TestMessageMarshalLength(t *testing.T) {
	t.Run("custom message", func(t *testing.T) {
		var dto = messageType{
			Timestamp: 100500,
			Time:      timeType{Time: time.Now()},
			LogLevel:  "Warning",
			Error: &errorType{
				Code:    42,
				Type:    "Business",
				Message: "cant do something: \"becouse of...,\"",
			},
			Fields: map[string]interface{}{
				"worker":  1,
				"arg1asd": "asddsadasd",
				"arg1":    "asds",
				"arg2":    "fdsjkfhdsfjkh",
			},
			Message: "while something",
		}

		jsonB, err := dto.MarshalJSON() // json.Marshal(dto)//
		if err != nil {
			t.Errorf("Error: %s", err)
			t.FailNow()
		}
		t.Logf("length %d", len(jsonB))

		timeByte, _ := dto.Time.MarshalJSON()

		etalonJsonB := `{"stamp":100500,"time":` + string(timeByte) + `,"level":"Warning","error":{"code":42,"type":"Business","message":"cant do something: \"becouse of...,\""},` +
			`"arg1":"asds","arg1asd":"asddsadasd","arg2":"fdsjkfhdsfjkh","worker":1,"message":"while something"}` + "\n"
		if string(jsonB) != string(etalonJsonB) {
			t.Errorf("%s", etalonJsonB)
			t.Errorf("%s", jsonB)
			t.Errorf("Fail: missmatch with etalon %d %d", len(jsonB), len(etalonJsonB))
			t.FailNow()
		}
	})

	t.Run("alternate message", func(t *testing.T) {
		type alternateMessage struct {
			Timestamp uint       `json:"stamp"`           // Это поле для обработки программой автоматического чтения логов (чтобы не приходилось парсить время)
			Time      timeType   `json:"time"`            // Это человекочитаемое время (без даты, дата в любом случае отображается в названии файла логгирования)
			LogLevel  string     `json:"level"`           // Error / Info / Debug / Warning...
			Error     *errorType `json:"error,omitempty"` // Код, тип (Internal / External / Request / Business) и тело ошибки
			Fields    map[string]interface{}
			Message   string `json:"message"`
		}
		var dto = alternateMessage{
			Timestamp: 100500,
			Time:      timeType{Time: time.Now()},
			LogLevel:  "Warning",
			Error: &errorType{
				Code:    42,
				Type:    "Business",
				Message: "cant do something: \"becouse of...,\"",
			},
			Fields: map[string]interface{}{
				"worker":  1,
				"arg1asd": "asddsadasd",
				"arg1":    "asds",
				"arg2":    "fdsjkfhdsfjkh",
			},
			Message: "while something",
		}

		jsonB, err := json.Marshal(dto)
		if err != nil {
			t.Errorf("Error: %s", err)
			t.FailNow()
		}
		t.Logf("length %d", len(jsonB))
	})

	t.Run("timestamp", func(t *testing.T) {
		now := time.Now()
		timestamp := now.Unix()
		t.Logf("timestamp %d", timestamp)
		t.Logf("time was %s recreated %s", now.Format("2006-01-02T15:04"), time.Unix(timestamp, 0).Format("2006-01-02T15:04"))
	})

	t.Run("slice", func(t *testing.T) {
		var arr = make([]int, 0, 3)
		if len(arr) != 0 {
			t.Errorf("Fail: len(arr)=%d != 0", len(arr))
		}
		if cap(arr) != 3 {
			t.Errorf("Fail: cap(arr)=%d != 3", len(arr))
		}
	})
}
