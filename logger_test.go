package flogger

import (
	yaml "github.com/GlobchanskyDenis/yaml"
	"errors"
	"sync"
	"testing"
)

const (
	GREEN    = "\033[32m"
	GREEN_BG = "\033[42;30m"
	YELLOW   = "\033[33m"
	BLUE     = "\033[34m"
	RED      = "\033[31m"
	RED_BG   = "\033[41;30m"
	NO_COLOR = "\033[m"
)

func TestInitAndTestGlobalLogger(t *testing.T) {
	configurator := yaml.NewConfigurator()

	if err := configurator.ReadFile("test.yaml"); err != nil {
		t.Errorf("%sError: %s%s", RED_BG, err, NO_COLOR)
		t.FailNow()
	}
	loggerConf := GetConfig()
	if err := configurator.ParseToStruct(loggerConf, "Logger"); err != nil {
		t.Errorf("%sError: %s%s", RED_BG, err, NO_COLOR)
		t.FailNow()
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	logger, err := NewLogger(wg)
	if err != nil {
		t.Errorf("%sError: %s%s", RED_BG, err, NO_COLOR)
		t.FailNow()
	}
	logger.Info(map[string]interface{}{
		"worker_num": 1,
		"field_n":    "value_n",
	}, "тестовый лог")

	logger.Stop()
	wg.Wait()

	t.Logf("%sSuccess%s", GREEN_BG, NO_COLOR)
}

func TestGlobalLoggerMultithread(t *testing.T) {
	configurator := yaml.NewConfigurator()
	if err := configurator.ReadFile("test.yaml"); err != nil {
		t.Errorf("%sError: %s%s", RED_BG, err, NO_COLOR)
		t.FailNow()
	}
	loggerConf := GetConfig()
	if err := configurator.ParseToStruct(loggerConf, "Logger"); err != nil {
		t.Errorf("%sError: %s%s", RED_BG, err, NO_COLOR)
		t.FailNow()
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	logger, err := NewLogger(wg)
	if err != nil {
		t.Errorf("%sError: %s%s", RED_BG, err, NO_COLOR)
		t.FailNow()
	}

	wg2 := &sync.WaitGroup{}
	wg2.Add(50)
	for i := 1; i <= 50; i++ {
		go t.Run("worker", func(t *testing.T) {
			for j := 0; j < 50; j++ {
				logger.Info(map[string]interface{}{
					"worker_num": 1,
					"field_n":    "value_n",
				}, "лог 1")
			}
			wg2.Done()
		})
	}

	wg2.Wait()
	logger.Stop()
	wg.Wait()

	if t.Failed() == false {
		t.Logf("%sSuccess%s", GREEN_BG, NO_COLOR)
	}
}

func TestDifferentLogTypes(t *testing.T) {
	configurator := yaml.NewConfigurator()
	if err := configurator.ReadFile("test.yaml"); err != nil {
		t.Errorf("%sError: %s%s", RED_BG, err, NO_COLOR)
		t.FailNow()
	}
	loggerConf := GetConfig()
	if err := configurator.ParseToStruct(loggerConf, "Logger"); err != nil {
		t.Errorf("%sError: %s%s", RED_BG, err, NO_COLOR)
		t.FailNow()
	}
	loggerConf.EnableFileForImportant = true
	loggerConf.FileWriteDurationSeconds = 1

	wg := &sync.WaitGroup{}
	wg.Add(1)
	logger, err := NewLogger(wg)
	if err != nil {
		t.Errorf("%sError: %s%s", RED_BG, err, NO_COLOR)
		t.FailNow()
	}

	logger.Fatal(map[string]interface{}{
		"key": "value",
	}, errors.New("_error_"), "message")
	logger.Error(map[string]interface{}{
		"key": "value",
	}, errors.New("_error_"), "message")
	logger.Warning(map[string]interface{}{
		"key": "value",
	}, errors.New("_error_"), "message")
	logger.Info(map[string]interface{}{
		"key": "value",
	}, "message")
	logger.ServiceDebug(map[string]interface{}{
		"key": "value",
	}, "message")
	logger.BusinessDebug(map[string]interface{}{
		"key": "value",
	}, "message")
	logger.Query(map[string]interface{}{
		"key": "value",
	}, "message")
	logger.Important(map[string]interface{}{
		"key": "value",
	}, "message")
	logger.Decision(map[string]interface{}{
		"key": "value",
	}, "message")

	logger.Stop()
	wg.Wait()

	if t.Failed() == false {
		t.Logf("%sSuccess%s", GREEN_BG, NO_COLOR)
	}
}
