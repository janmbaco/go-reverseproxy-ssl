package cross

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type LogLevel int

const (
	Trace = iota
	Info
	Warning
	Error
	Fatal
)

type mylog struct {
	loggers             map[LogLevel]*log.Logger
	activeConsoleLogger map[LogLevel]bool
	activeFileLogger    map[LogLevel]bool
	ErrorLogger         *log.Logger
	logsDir             string
}

var Log mylog

func init() {

	Log = mylog{
		loggers:             make(map[LogLevel]*log.Logger),
		activeConsoleLogger: setLevel(Trace),
		activeFileLogger:    setLevel(Trace),
	}

	createLog := func(level LogLevel, writer io.Writer) *log.Logger {
		levels := [...]string{
			"TRACE: ",
			"INFO: ",
			"WARNING: ",
			"ERROR: ",
			"FATAL: "}

		return log.New(writer,
			levels[level],
			log.Ldate|log.Ltime)

	}

	registerLogger := func(consoleWriter io.Writer, levels ...LogLevel) {
		for _, level := range levels {
			Log.loggers[level] = createLog(level, consoleWriter)
		}
	}

	registerLogger(os.Stdout, Trace, Info, Warning)
	registerLogger(os.Stderr, Error, Fatal)
	Log.ErrorLogger = Log.loggers[Error]
}

func(l *mylog) SetDir(dir string){
	l.logsDir, _ = filepath.Abs(dir)
}

func (l *mylog) Println(level LogLevel, message string) {
	var writers []io.Writer
	if l.activeConsoleLogger[level] {
		if level < Error {
			writers = append(writers, os.Stdout)
		} else {
			writers = append(writers, os.Stderr)
		}
	}
	if len(l.logsDir) > 0 && l.activeFileLogger[level]  {
		year, month, day := time.Now().Date()
		execFile := filepath.Base(os.Args[0])
	
		logFile := l.logsDir + "/" + execFile + "-" + strconv.Itoa(year) + strconv.Itoa(int(month)) + strconv.Itoa(day) + ".log"
		_ = os.MkdirAll(filepath.Dir(logFile), 0666)
		osFile, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Println("impossible to log in file:", err)
		}
		defer osFile.Close()
		writers = append(writers, osFile)
	}

	if writers != nil {
		multiWriter := io.MultiWriter(writers...)
		l.loggers[level].SetOutput(multiWriter)

		if level == Fatal {
			l.loggers[level].Fatalln(message)
		} else {
			l.loggers[level].Println(message)
		}
	}
}

func (l *mylog) Trace(message string) {
	l.Println(Trace, message)
}

func (l *mylog) Info(message string) {
	l.Println(Info, message)
}

func (l *mylog) Warning(message string) {
	l.Println(Warning, message)
}

func (l *mylog) Error(message string) {
	l.Println(Error, message)
}

func (l *mylog) Fatal(message string) {
	l.Println(Fatal, message)
}

func setLevel(level LogLevel) map[LogLevel]bool {
	loggersActives := map[LogLevel]bool{Trace: true, Info: true, Warning: true, Error: true, Fatal: true}
	if level > Trace {
		start := level - 1
		for i := start; i > -1; i-- {
			loggersActives[i] = false
		}
	}
	return loggersActives
}

func (l *mylog) SetConsoleLevel(level LogLevel) {
	l.activeConsoleLogger = setLevel(level)
}

func (l *mylog) SetFileLogLevel(level LogLevel) {
	l.activeFileLogger = setLevel(level)
}
