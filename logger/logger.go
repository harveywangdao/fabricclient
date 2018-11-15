package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

type Level int32

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

/*
===================
 utils functions
===================
*/
func fileSize(file string) int64 {
	fmt.Println("fileSize", file)
	f, e := os.Stat(file)
	if e != nil {
		fmt.Println(e.Error())
		return 0
	}
	return f.Size()
}

func isExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

/*
===================
 log handlers
===================
*/
type Handler interface {
	SetOutput(w io.Writer)
	Output(calldepth int, s string) error
	Printf(format string, v ...interface{})
	Print(v ...interface{})
	Println(v ...interface{})
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Fatalln(v ...interface{})

	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})

	ErrorD(calldepth int, v ...interface{})

	Flags() int
	SetFlags(flag int)
	Prefix() string
	SetPrefix(prefix string)
	close()
}

type LogHandler struct {
	lg      *log.Logger
	lgDebug *log.Logger
	lgInfo  *log.Logger
	lgWarn  *log.Logger
	lgError *log.Logger
}

type ConsoleHander struct {
	LogHandler
}

type FileHandler struct {
	LogHandler
	logfile *os.File
}

type RotatingHandler struct {
	LogHandler
	dir      string
	filename string
	maxNum   int
	maxSize  int64
	suffix   int
	logfile  *os.File
	mu       sync.Mutex
}

var Console = NewConsoleHandler()

func NewConsoleHandler() *ConsoleHander {
	lg := log.New(os.Stderr, "", log.LstdFlags)
	lgDebug := log.New(os.Stderr, "[DEBUG] ", log.LstdFlags)
	lgInfo := log.New(os.Stderr, "[INFO] ", log.LstdFlags)
	lgWarn := log.New(os.Stderr, "[WARN] ", log.LstdFlags)
	lgError := log.New(os.Stderr, "[ERROR] ", log.LstdFlags)
	return &ConsoleHander{
		LogHandler: LogHandler{
			lg:      lg,
			lgDebug: lgDebug,
			lgInfo:  lgInfo,
			lgWarn:  lgWarn,
			lgError: lgError,
		}}
}

func NewFileHandler(filepath string) *FileHandler {
	logfile, _ := os.OpenFile(filepath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	lg := log.New(logfile, "", log.LstdFlags)
	lgDebug := log.New(logfile, "[DEBUG] ", log.LstdFlags)
	lgInfo := log.New(logfile, "[INFO] ", log.LstdFlags)
	lgWarn := log.New(logfile, "[WARN] ", log.LstdFlags)
	lgError := log.New(logfile, "[ERROR] ", log.LstdFlags)
	return &FileHandler{
		LogHandler: LogHandler{
			lg:      lg,
			lgDebug: lgDebug,
			lgInfo:  lgInfo,
			lgWarn:  lgWarn,
			lgError: lgError,
		},
		logfile: logfile,
	}
}

func NewRotatingHandler(dir string, filename string, maxNum int, maxSize int64) *RotatingHandler {
	logfile, _ := os.OpenFile(dir+"/"+filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	lg := log.New(logfile, "", log.LstdFlags)
	lgDebug := log.New(logfile, "[DEBUG] ", log.LstdFlags)
	lgInfo := log.New(logfile, "[INFO] ", log.LstdFlags)
	lgWarn := log.New(logfile, "[WARN] ", log.LstdFlags)
	lgError := log.New(logfile, "[ERROR] ", log.LstdFlags)

	h := &RotatingHandler{
		LogHandler: LogHandler{
			lg:      lg,
			lgDebug: lgDebug,
			lgInfo:  lgInfo,
			lgWarn:  lgWarn,
			lgError: lgError,
		},
		dir:      dir,
		filename: filename,
		maxNum:   maxNum,
		maxSize:  maxSize,
		suffix:   0,
	}

	if h.isMustRename() {
		h.rename()
	} else {
		h.mu.Lock()
		defer h.mu.Unlock()
		h.lg.SetOutput(logfile)
	}

	// monitor filesize per second
	go func() {
		timer := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-timer.C:
				h.fileCheck()
			}
		}
	}()

	return h
}

func (l *LogHandler) SetOutput(w io.Writer) {
	l.lg.SetOutput(w)
}

func (l *LogHandler) Output(calldepth int, s string) error {
	return l.lg.Output(calldepth, s)
}

func (l *LogHandler) Print(v ...interface{}) {
	//l.lg.Print(v...)
	l.lg.Output(3, fmt.Sprint(v...))
}

func (l *LogHandler) Printf(format string, v ...interface{}) {
	//l.lg.Printf(format, v...)
	l.lg.Output(3, fmt.Sprintf(format, v...))
}

func (l *LogHandler) Println(v ...interface{}) {
	//l.lg.Println(v...)
	l.lg.Output(3, fmt.Sprintln(v...))
}

func (l *LogHandler) Fatal(v ...interface{}) {
	l.lg.Output(3, fmt.Sprint(v...))
}

func (l *LogHandler) Fatalf(format string, v ...interface{}) {
	l.lg.Output(3, fmt.Sprintf(format, v...))
}

func (l *LogHandler) Fatalln(v ...interface{}) {
	l.lg.Output(3, fmt.Sprintln(v...))
}

func (l *LogHandler) Flags() int {
	return l.lg.Flags()
}

func (l *LogHandler) SetFlags(flag int) {
	l.lg.SetFlags(flag)
	l.lgDebug.SetFlags(flag)
	l.lgInfo.SetFlags(flag)
	l.lgWarn.SetFlags(flag)
	l.lgError.SetFlags(flag)
}

func (l *LogHandler) Prefix() string {
	return l.lg.Prefix()
}

func (l *LogHandler) SetPrefix(prefix string) {
	l.lg.SetPrefix(prefix)
}

func (l *LogHandler) Debug(v ...interface{}) {
	//l.lg.Output(3, fmt.Sprintln("[DEBUG]", v))
	l.lgDebug.Output(3, fmt.Sprintln(v))
}

func (l *LogHandler) Info(v ...interface{}) {
	l.lgInfo.Output(3, fmt.Sprintln(v))
}

func (l *LogHandler) Warn(v ...interface{}) {
	l.lgWarn.Output(3, fmt.Sprintln(v))
}

func (l *LogHandler) Error(v ...interface{}) {
	l.lgError.Output(3, fmt.Sprintln(v))
}

func (l *LogHandler) close() {

}

func (h *FileHandler) close() {
	if h.logfile != nil {
		h.logfile.Close()
	}
}

func (h *RotatingHandler) close() {
	if h.logfile != nil {
		h.logfile.Close()
	}
}

func (h *RotatingHandler) isMustRename() bool {
	if h.maxNum > 1 {
		if fileSize(h.dir+"/"+h.filename) >= h.maxSize {
			return true
		}
	}
	return false
}

func (h *RotatingHandler) rename() {
	h.suffix = h.suffix%h.maxNum + 1

	if h.logfile != nil {
		h.logfile.Close()
	}

	newpath := fmt.Sprintf("%s/%s.%d.log", h.dir, h.filename, h.suffix)
	if isExist(newpath) {
		os.Remove(newpath)
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	filepath := h.dir + "/" + h.filename
	os.Rename(filepath, newpath)
	h.logfile, _ = os.Create(filepath)
	h.lg.SetOutput(h.logfile)
}

func (h *RotatingHandler) fileCheck() {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	if h.isMustRename() {
		h.rename()
	}
}

/*
===================
 logger
===================
*/
type _Logger struct {
	handlers []Handler
	level    Level
	mu       sync.Mutex
}

var logger = &_Logger{
	handlers: []Handler{
		Console,
	},
	level: DEBUG,
}

func SetHandlers(handlers ...Handler) {
	logger.handlers = handlers
}

func SetFlags(flag int) {
	for i := range logger.handlers {
		logger.handlers[i].SetFlags(flag)
	}
}

func SetLevel(level Level) {
	logger.level = level
}

func Print(v ...interface{}) {
	for i := range logger.handlers {
		logger.handlers[i].Print(v...)
	}
}

func Printf(format string, v ...interface{}) {
	for i := range logger.handlers {
		logger.handlers[i].Printf(format, v...)
	}
}

func Println(v ...interface{}) {
	for i := range logger.handlers {
		logger.handlers[i].Println(v...)
	}
}

func Fatal(v ...interface{}) {
	for i := range logger.handlers {
		logger.handlers[i].Fatal(v...)
	}
	os.Exit(1)
}

func Fatalf(format string, v ...interface{}) {
	for i := range logger.handlers {
		logger.handlers[i].Fatalf(format, v...)
	}
	os.Exit(1)
}

func Fatalln(v ...interface{}) {
	for i := range logger.handlers {
		logger.handlers[i].Fatalln(v...)
	}
	os.Exit(1)
}

func Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	for i := range logger.handlers {
		logger.handlers[i].Output(3, s)
	}
	panic(s)
}

func Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	for i := range logger.handlers {
		logger.handlers[i].Output(3, s)
	}
	panic(s)
}

func Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	for i := range logger.handlers {
		logger.handlers[i].Output(3, s)
	}
	panic(s)
}

func Debug(v ...interface{}) {
	if logger.level <= DEBUG {
		for i := range logger.handlers {
			logger.handlers[i].Debug(v...)
		}
	}
}

func Info(v ...interface{}) {
	if logger.level <= INFO {
		for i := range logger.handlers {
			logger.handlers[i].Info(v...)
		}
	}
}

func Warn(v ...interface{}) {
	if logger.level <= WARN {
		for i := range logger.handlers {
			logger.handlers[i].Warn(v...)
		}
	}
}

func Error(v ...interface{}) {
	if logger.level <= ERROR {
		for i := range logger.handlers {
			logger.handlers[i].Error(v...)
		}
	}
}

func Close() {
	for i := range logger.handlers {
		logger.handlers[i].close()
	}
}

func (l *LogHandler) ErrorD(calldepth int, v ...interface{}) {
	l.lgError.Output(calldepth, fmt.Sprintln(v))
}

func ErrorD(calldepth int, v ...interface{}) {
	if logger.level <= ERROR {
		for i := range logger.handlers {
			logger.handlers[i].ErrorD(calldepth, v...)
		}
	}
}

func CheckError(err error) {
	if err != nil {
		ErrorD(4, err)
		os.Exit(1)
	}
}
