package mini_orm

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

type LogLevel int

// define lv cua log
const (
	FATAL LogLevel = iota
	ERROR
	WARN
	INFO
	TRACE
)

var currentLevel = INFO
var mutex = &sync.RWMutex{}

// StringToLevel converts string to  LogLevel
func StringToLevel(level string) LogLevel {
	level = strings.ToUpper(level)
	switch level {
	case "FATAL":
		return FATAL
	case "ERROR":
		return ERROR
	case "WARN":
		return WARN
	case "INFO":
		return INFO
	case "TRACE":
		return TRACE
	default:
		Errorf("Lỗi không tìm thấy log level %s, set về mặc định là  TRACE.\n", level)
		return TRACE
	}
}

// SetLogLevel changes the current log level
func SetLogLevel(level LogLevel) {
	mutex.Lock()
	currentLevel = level
	mutex.Unlock()
}

func pad(text string, length int) string {
	strlen := len(text)
	if strlen < length {
		text += strings.Repeat(" ", length-strlen)
	}
	return text
}

// funcName gets the current function name from a pointer
func funcName(ptr uintptr) string {
	fname := runtime.FuncForPC(ptr).Name()
	lastDot := 0
	for i := 0; i < len(fname); i++ {
		if fname[i] == '.' {
			lastDot = i
		}
	}
	if lastDot == 0 {
		return filepath.Base(fname)
	}
	return fname[lastDot+1:] + "()"
}

// goroutineID fetches the current goroutine ID. Used solely for
// debugging which goroutine is doing what in the logs.
func goroutineID() uint64 {
	buf := make([]byte, 64)
	buf = buf[:runtime.Stack(buf, false)]
	// parse out # in the format "goroutine # "
	buf = bytes.TrimPrefix(buf, []byte("goroutine "))
	buf = buf[:bytes.IndexByte(buf, ' ')]
	id, _ := strconv.ParseUint(string(buf), 10, 64)
	return id
}

func Caller(level int) string {
	ptr, file, line, ok := runtime.Caller(level)
	var functionName string
	if ok {
		functionName = funcName(ptr)
	} else {
		functionName = "(unknown source)"
	}

	return fmt.Sprintf("%d:%s:%d:%s",
		goroutineID(), filepath.Base(file), line, functionName)
}

func logger(level LogLevel, format string, args ...interface{}) {
	mutex.RLock()
	defer mutex.RUnlock()
	if level > currentLevel {
		return
	}

	var prefix string
	switch level {
	case FATAL:
		prefix = "FATAL"
	case ERROR:
		prefix = "ERROR"
	case WARN:
		prefix = "WARN"
	case INFO:
		prefix = "INFO"
	case TRACE:
		prefix = "TRACE"
	}

	preformatted := fmt.Sprintf(format, args...)

	log.Printf("- %s - %s - %s",
		pad(prefix, 5), // log level
		Caller(3),      // goroutine + function being logged
		preformatted)   // actual log message
}

func Fatalf(format string, args ...interface{}) {
	logger(FATAL, format, args...)
	os.Exit(1)
}

func Fatal(args ...interface{}) {
	logger(FATAL, "%s", fmt.Sprintln(args...))
	os.Exit(1)
}

func Errorf(format string, args ...interface{}) {
	logger(ERROR, format, args...)
}

func Error(args ...interface{}) {
	logger(ERROR, "%s", fmt.Sprintln(args...))
}

func Warnf(format string, args ...interface{}) {
	logger(WARN, format, args...)
}

func Warn(args ...interface{}) {
	logger(WARN, "%s", fmt.Sprintln(args...))
}

func Infof(format string, args ...interface{}) {
	logger(INFO, format, args...)
}

func Info(args ...interface{}) {
	logger(INFO, "%s", fmt.Sprintln(args...))
}

func Tracef(format string, args ...interface{}) {
	logger(TRACE, format, args...)
}

func Trace(args ...interface{}) {
	logger(TRACE, "%s", fmt.Sprintln(args...))
}
