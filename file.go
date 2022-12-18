package flogger

import (
	utils "github.com/GlobchanskyDenis/file_logger/pkg/utils"
	"fmt"
	"os"
	"sync"
	"time"
)

/*	мьютексы необходимы чтобы логгер мог использовать буффер потокобезопасно  */
type fileType struct {
	serviceName   string
	logFolder     string
	permissions   string
	maxHours      uint
	maxBufSize    uint
	writeChanSize uint
	fileTypeName  string // default / additional
	fileName      string
	currentDate   time.Time
	osFile        *os.File
	bmu           *sync.Mutex                        // буфферный мьютекс
	buf           []messageType                      // в этом буффере хранятся сформированные сообщения. Буффер отправляется на запись в файл либо при его заполнении либо по таймауту (тикер)
	writeChan     chan []messageType                 // буфферизированный канал передачи между объектом логгера который наполняет буффер и горутиной записи в файл
	errorHandler  func(error) (uint, string, string) // функция извлечения из ошибки ее кода, типа и сообщения
}

func newFile(fileTypeName string, conf *ConfigType) *fileType {
	return &fileType{
		serviceName:   conf.ServiceName,
		logFolder:     conf.LogFolder,
		permissions:   conf.Permissions,
		maxHours:      conf.MaxHoursToChangeLogFile,
		maxBufSize:    conf.MaxBufSize,
		writeChanSize: conf.WriteChanSize,
		fileTypeName:  fileTypeName,
		bmu:           &sync.Mutex{},
		buf:           make([]messageType, 0, int(conf.MaxBufSize)+1),
		writeChan:     make(chan []messageType, int(conf.WriteChanSize)),
		errorHandler:  defaultErrorHandler,
	}
}

func (this *fileType) SetErrorHandler(errorHandler func(error) (uint, string, string)) {
	this.errorHandler = errorHandler
}

func (this *fileType) addToBuffer(level string, err error, fields map[string]interface{}, message string) {
	now := time.Now()
	var cerr *errorType
	if err != nil {
		if this.errorHandler != nil {
			code, errType, errMessage := this.errorHandler(err)
			cerr = &errorType{
				Code:    code,
				Type:    errType,
				Message: errMessage,
			}
		} else {
			println("Warning: file logger found case errorHandler == nil")
			cerr = &errorType{
				Code:    0,
				Type:    "",
				Message: err.Error(),
			}
		}
	}

	this.bmu.Lock()
	this.buf = append(this.buf, messageType{
		Timestamp: now.Unix(),
		Time: timeType{
			Time: now,
		},
		LogLevel: level,
		Error:    cerr,
		Fields:   fields,
		Message:  message,
	})

	/*	Проверяю заполненность буффера - возможно его пора отправить в файл  */
	if len(this.buf) >= int(this.maxBufSize) {
		/*	Копирую буффер чтобы можно было снять буфферный мьютекс  */
		cpyBuf := this.buf
		this.buf = make([]messageType, 0, int(this.maxBufSize)+1)

		/*	Отправляю буффер в горутину для записи в файл  */
		this.writeChan <- cpyBuf
	}
	this.bmu.Unlock()
}

func (this *fileType) GetWriteChan() chan []messageType {
	return this.writeChan
}

func (this *fileType) writeFromBufIfNotEmpty() {
	var cpyBuf []messageType
	this.bmu.Lock()
	if len(this.buf) > 0 {
		cpyBuf = this.buf
		this.buf = make([]messageType, 0, int(this.writeChanSize))
	}
	this.bmu.Unlock()
	if len(cpyBuf) > 0 {
		this.write(convertBufToBite(cpyBuf))
	}
}

func (this *fileType) write(message []byte) {
	if err := this.changeLogFileIfItNeeded(); err != nil {
		fmt.Fprintf(os.Stderr, "Не смог сменить файл логгирования %s", err)
	}
	if amountWrited, err := this.osFile.Write(message); err != nil {
		/*	Повторная попытка нужна для случая когда дата сменилась и какой-то воркер
		**	в другом потоке начал менять дескриптор файла для логгирования. Считаю 500 миллисекунд
		**	достаточно. Не блокировать же мьютексами запись в файл :)  */
		time.Sleep(500 * time.Millisecond)
		if amountWrited, err := this.osFile.Write(message); err != nil {
			fmt.Fprintf(os.Stderr, "Не смог залоггировать в файл %s %s", err, string(message))
		} else if amountWrited != len(message) {
			fmt.Fprintf(os.Stderr, "Логи не записались в файл - ожидалось %d байт записалось %d байт", len(message), amountWrited)
		}
	} else if amountWrited != len(message) {
		fmt.Fprintf(os.Stderr, "Логи не записались в файл - ожидалось %d байт записалось %d байт", len(message), amountWrited)
	}
}

/*	Меняет файл в который записывается логгирование в случае если уже сменилась дата
**	Использует мьютекс, поэтому выполняется горутинобезопасно  */
func (this *fileType) changeLogFileIfItNeeded() error {
	if utils.IsSameDateAndHour(this.currentDate, time.Now(), this.maxHours) == false {
		if err := this.setNewLogFile(); err != nil {
			return err
		}
	}
	return nil
}

/*	Данную функцию в многопоточном режиме нужно запускать в   */
func (this *fileType) setNewLogFile() error {
	/*	Если ранее уже был открыт файл - закрываю его  */
	if err := this.Close(); err != nil {
		return nil
	}

	/*	Открываю новый файл  */
	this.currentDate = time.Now()
	var hour int
	if this.maxHours < 24 {
		hour = utils.CalcHour(this.currentDate.Hour(), int(this.maxHours))
	}
	this.fileName = fmt.Sprintf("%s_%s_%d-%02d-%02d_%02d.log", this.serviceName, this.fileTypeName, this.currentDate.Year(), this.currentDate.Month(), this.currentDate.Day(), hour)
	file, err := utils.OpenOrCreateNewFile(this.logFolder, this.fileName, this.permissions)
	if err != nil {
		return err
	}
	this.osFile = file
	return nil
}

func (this *fileType) Close() error {
	if this.osFile != nil {
		if err := this.osFile.Close(); err != nil {
			return fmt.Errorf("Не смог закрыть лог файл %w", err)
		}
	}
	return nil
}
