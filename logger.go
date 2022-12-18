package flogger

import (
	"fmt"
	"os"
	"sync"
	"time"
)

const (
	fatalLevel         = "FATAL"
	errorLevel         = "ERROR"
	warningLevel       = "WARNING"
	infoLevel          = "INFO"
	serviceDebugLevel  = "DEBUG"
	businessDebugLevel = "DEBUG"
	queryLevel         = "QUERY"
	importantLevel     = "IMPORTANT"
	decisionLevel      = "DECISION"
)

var gLogger *LoggerType

type LoggerType struct {
	enableServiceDebug  bool
	enableBusinessDebug bool
	enableQuery         bool
	enableImportant     bool
	enableDecision      bool
	fatalTrigger        IMonitoringTrigger
	errorTrigger        IMonitoringTrigger
	importantTrigger    IMonitoringTrigger
	defaultFile         *fileType
	importantFile       *fileType
	queryFile           *fileType
}

var _ IServiceLogger = (*LoggerType)(nil)
var _ IBusinessLogger = (*LoggerType)(nil)

type IServiceLogger interface {
	Fatal(fields map[string]interface{}, err error, format string, args ...interface{})
	Error(fields map[string]interface{}, err error, format string, args ...interface{})
	Warning(fields map[string]interface{}, err error, format string, args ...interface{})
	Info(fields map[string]interface{}, format string, args ...interface{})
	ServiceDebug(fields map[string]interface{}, format string, args ...interface{})
	SetFatalMonitoringTrigger(IMonitoringTrigger)
	SetErrorMonitoringTrigger(IMonitoringTrigger)
	SetImportantMonitoringTrigger(IMonitoringTrigger)
	SetErrorHandler(errorHandler func(error) (uint, string, string))
	Stop()
}

type IBusinessLogger interface {
	BusinessDebug(fields map[string]interface{}, format string, args ...interface{})
	Query(fields map[string]interface{}, format string, args ...interface{})
	Important(fields map[string]interface{}, format string, args ...interface{}) // Пример - дедлайн запроса или другое важное событие. Логгирование в файл исключений
	Decision(fields map[string]interface{}, format string, args ...interface{})  // Пример -- такая-то проверка не прошла поэтому дальше логика будет двигаться в таком-то направлении. Логгирование в дефолтный файл
}

type IMonitoringTrigger interface {
	Trig()
}

func NewLogger(wg *sync.WaitGroup) (*LoggerType, error) {
	if err := checkConfig(); err != nil {
		return nil, err
	}
	conf := GetConfig()

	logger := &LoggerType{
		enableServiceDebug:  conf.EnableServiceDebug,
		enableBusinessDebug: conf.EnableBusinessDebug,
		enableQuery:         conf.EnableQuery,
		enableImportant:     conf.EnableImportant,
		enableDecision:      conf.EnableDecision,
		defaultFile:         newFile("default", conf),
	}

	if err := logger.defaultFile.setNewLogFile(); err != nil {
		return nil, err
	}

	/*	Тут переопределяется файл сразу для 3-х уровней логгирования - Important Error Fatal */
	if conf.EnableFileForImportant == true {
		logger.importantFile = newFile("important", conf)
		if err := logger.importantFile.setNewLogFile(); err != nil {
			return nil, err
		}
	}

	/*	Тут переопределяется файл для уровня логгирования Query */
	if conf.EnableFileForQuery == true {
		logger.queryFile = newFile("query", conf)
		if err := logger.queryFile.setNewLogFile(); err != nil {
			return nil, err
		}
	}

	gLogger = logger
	go logger.writeLoopAsync(wg, conf.FileWriteDurationSeconds)

	return logger, nil
}

func (this *LoggerType) SetFatalMonitoringTrigger(trigger IMonitoringTrigger) {
	this.fatalTrigger = trigger
}

func (this *LoggerType) SetErrorMonitoringTrigger(trigger IMonitoringTrigger) {
	this.errorTrigger = trigger
}

func (this *LoggerType) SetImportantMonitoringTrigger(trigger IMonitoringTrigger) {
	this.importantTrigger = trigger
}

func (this *LoggerType) SetErrorHandler(errorHandler func(error) (uint, string, string)) {
	this.defaultFile.SetErrorHandler(errorHandler)
	if this.importantFile != nil {
		this.importantFile.SetErrorHandler(errorHandler)
	}
}

func (this *LoggerType) Fatal(fields map[string]interface{}, err error, msg string, args ...interface{}) {
	if this.importantFile != nil {
		this.importantFile.addToBuffer(fatalLevel, err, fields, fmt.Sprintf(msg, args...))
	}
	this.defaultFile.addToBuffer(fatalLevel, err, fields, fmt.Sprintf(msg, args...))
	if this.fatalTrigger != nil {
		go this.fatalTrigger.Trig()
	}
}

func (this *LoggerType) Error(fields map[string]interface{}, err error, msg string, args ...interface{}) {
	if this.importantFile != nil {
		this.importantFile.addToBuffer(errorLevel, err, fields, fmt.Sprintf(msg, args...))
	}
	this.defaultFile.addToBuffer(errorLevel, err, fields, fmt.Sprintf(msg, args...))
	if this.errorTrigger != nil {
		go this.errorTrigger.Trig()
	}
}

func (this *LoggerType) Warning(fields map[string]interface{}, err error, msg string, args ...interface{}) {
	this.defaultFile.addToBuffer(warningLevel, err, fields, fmt.Sprintf(msg, args...))
}

func (this *LoggerType) Info(fields map[string]interface{}, msg string, args ...interface{}) {
	this.defaultFile.addToBuffer(infoLevel, nil, fields, fmt.Sprintf(msg, args...))
}

func (this *LoggerType) ServiceDebug(fields map[string]interface{}, msg string, args ...interface{}) {
	if this.enableServiceDebug == true {
		this.defaultFile.addToBuffer(serviceDebugLevel, nil, fields, fmt.Sprintf(msg, args...))
	}
}

func (this *LoggerType) BusinessDebug(fields map[string]interface{}, msg string, args ...interface{}) {
	if this.enableBusinessDebug == true {
		this.defaultFile.addToBuffer(businessDebugLevel, nil, fields, fmt.Sprintf(msg, args...))
	}
}

func (this *LoggerType) Query(fields map[string]interface{}, msg string, args ...interface{}) {
	if this.enableQuery == true {
		if this.queryFile != nil {
			this.queryFile.addToBuffer(queryLevel, nil, fields, fmt.Sprintf(msg, args...))
		} else {
			this.defaultFile.addToBuffer(queryLevel, nil, fields, fmt.Sprintf(msg, args...))
		}
	}
}

func (this *LoggerType) Important(fields map[string]interface{}, msg string, args ...interface{}) {
	if this.enableImportant == true {
		if this.importantFile != nil {
			this.importantFile.addToBuffer(importantLevel, nil, fields, fmt.Sprintf(msg, args...))
		}
		this.defaultFile.addToBuffer(importantLevel, nil, fields, fmt.Sprintf(msg, args...))
	}
	if this.importantTrigger != nil {
		go this.importantTrigger.Trig()
	}
}

func (this *LoggerType) Decision(fields map[string]interface{}, msg string, args ...interface{}) {
	if this.enableDecision == true {
		this.defaultFile.addToBuffer(decisionLevel, nil, fields, fmt.Sprintf(msg, args...))
	}
}

func (this *LoggerType) Stop() {
	close(this.defaultFile.GetWriteChan())
	if this.importantFile != nil {
		close(this.importantFile.GetWriteChan())
	}
	if this.queryFile != nil {
		close(this.queryFile.GetWriteChan())
	}
}

func (this *LoggerType) writeLoopAsync(wg *sync.WaitGroup, fileWriteDurationSeconds uint) {
	ticker := time.NewTicker(time.Second * time.Duration(fileWriteDurationSeconds))
	if this.importantFile != nil && this.queryFile != nil {
		for {
			select {
			case <-ticker.C:
				this.defaultFile.writeFromBufIfNotEmpty()
				if this.importantFile != nil {
					this.importantFile.writeFromBufIfNotEmpty()
				}
				if this.queryFile != nil {
					this.queryFile.writeFromBufIfNotEmpty()
				}
			case cpyBuf, inWork := <-this.defaultFile.GetWriteChan():
				if inWork == false {
					this.defaultFile.writeFromBufIfNotEmpty()
					if this.importantFile != nil {
						this.importantFile.writeFromBufIfNotEmpty()
					}
					if this.queryFile != nil {
						this.queryFile.writeFromBufIfNotEmpty()
					}
					ticker.Stop()
					if err := this.defaultFile.Close(); err != nil {
						fmt.Fprintf(os.Stderr, "%s", err)
					}
					if err := this.importantFile.Close(); err != nil {
						fmt.Fprintf(os.Stderr, "%s", err)
					}
					if err := this.queryFile.Close(); err != nil {
						fmt.Fprintf(os.Stderr, "%s", err)
					}
					wg.Done()
					return
				}
				if len(cpyBuf) > 0 {
					this.defaultFile.write(convertBufToBite(cpyBuf))
				}
			case cpyBuf := <-this.importantFile.GetWriteChan():
				if len(cpyBuf) > 0 {
					this.importantFile.write(convertBufToBite(cpyBuf))
				}
			case cpyBuf := <-this.queryFile.GetWriteChan():
				if len(cpyBuf) > 0 {
					this.queryFile.write(convertBufToBite(cpyBuf))
				}
			}
		}
	} else if this.importantFile != nil {
		for {
			select {
			case <-ticker.C:
				this.defaultFile.writeFromBufIfNotEmpty()
				if this.importantFile != nil {
					this.importantFile.writeFromBufIfNotEmpty()
				}
			case cpyBuf, inWork := <-this.defaultFile.GetWriteChan():
				if inWork == false {
					this.defaultFile.writeFromBufIfNotEmpty()
					if this.importantFile != nil {
						this.importantFile.writeFromBufIfNotEmpty()
					}
					ticker.Stop()
					if err := this.defaultFile.Close(); err != nil {
						fmt.Fprintf(os.Stderr, "%s", err)
					}
					if err := this.importantFile.Close(); err != nil {
						fmt.Fprintf(os.Stderr, "%s", err)
					}
					wg.Done()
					return
				}
				if len(cpyBuf) > 0 {
					this.defaultFile.write(convertBufToBite(cpyBuf))
				}
			case cpyBuf := <-this.importantFile.GetWriteChan():
				if len(cpyBuf) > 0 {
					this.importantFile.write(convertBufToBite(cpyBuf))
				}
			}
		}
	} else if this.queryFile != nil {
		for {
			select {
			case <-ticker.C:
				this.defaultFile.writeFromBufIfNotEmpty()
				if this.queryFile != nil {
					this.queryFile.writeFromBufIfNotEmpty()
				}
			case cpyBuf, inWork := <-this.defaultFile.GetWriteChan():
				if inWork == false {
					this.defaultFile.writeFromBufIfNotEmpty()
					if this.queryFile != nil {
						this.queryFile.writeFromBufIfNotEmpty()
					}
					ticker.Stop()
					if err := this.defaultFile.Close(); err != nil {
						fmt.Fprintf(os.Stderr, "%s", err)
					}
					if err := this.queryFile.Close(); err != nil {
						fmt.Fprintf(os.Stderr, "%s", err)
					}
					wg.Done()
					return
				}
				if len(cpyBuf) > 0 {
					this.defaultFile.write(convertBufToBite(cpyBuf))
				}
			case cpyBuf := <-this.queryFile.GetWriteChan():
				if len(cpyBuf) > 0 {
					this.queryFile.write(convertBufToBite(cpyBuf))
				}
			}
		}
	} else {
		for {
			select {
			case <-ticker.C:
				this.defaultFile.writeFromBufIfNotEmpty()
			case cpyBuf, inWork := <-this.defaultFile.GetWriteChan():
				if inWork == false {
					this.defaultFile.writeFromBufIfNotEmpty()
					ticker.Stop()
					if err := this.defaultFile.Close(); err != nil {
						fmt.Fprintf(os.Stderr, "%s", err)
					}
					wg.Done()
					return
				}
				if len(cpyBuf) > 0 {
					this.defaultFile.write(convertBufToBite(cpyBuf))
				}
			}
		}
	}
}
