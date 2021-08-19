package slog

import (
	"fmt"
	"os"
	"path"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

type FileWriter struct {
	// The opened file
	filename string
	file     *os.File
}

type LogWriter struct {
	rec       chan *LogRecord
	init      chan string
	init_path string
	b_init    int32
	buf       []byte
	filter    map[string]*FileWriter
	end_wait  sync.WaitGroup
	path      string
	dayTimer  *time.Timer
}

func (lw *LogWriter) LogWrite(rec *LogRecord) {
	lw.rec <- rec
}

var _log_writer = &LogWriter{
	rec:    make(chan *LogRecord, LogBufferLength),
	init:   make(chan string),
	filter: make(map[string]*FileWriter),
	buf:    make([]byte, 2048),
	path:   "",
	b_init: 0,
}

func (lw *LogWriter) writeFile(log *LogRecord) {
	file_writer, ok := _log_writer.filter[log.Name]
	if !ok {
		path := path.Join(lw.path, log.Name+".log")
		fd, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
		if err != nil {
			return
		}
		file_writer = &FileWriter{log.Name, fd}
		_log_writer.filter[log.Name] = file_writer
	}

	file_writer.file.Write(_log_writer.buf)
}

func (lw *LogWriter) writeStd(log *LogRecord) {
	os.Stdout.Write(_log_writer.buf)
}

func (lw *LogWriter) InitPath(p string) {
	if atomic.CompareAndSwapInt32(&lw.b_init, 0, 1) {
		lw.init <- p
	}
}

func (lw *LogWriter) MakePath() {
	ok := false
	if file, err := os.Stat(lw.path); err != nil {
		ok = os.IsExist(err)
	} else {
		ok = file.IsDir()
	}

	if !ok {
		os.MkdirAll(lw.path, 0x777)
	}
}

func (lw *LogWriter) MakeLogTimePath() {
	now := time.Now()
	dirName := fmt.Sprintf("%d-%02d-%02d",
		now.Year(),
		now.Month(),
		now.Day())

	lw.path = path.Join(lw.init_path, dirName)

	lw.MakePath()
	zero_time := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	delayTime := 24*60*60 - (now.Unix() - zero_time.Unix())

	if lw.dayTimer == nil {
		lw.dayTimer = time.NewTimer(time.Duration(delayTime) * time.Second)
	} else {
		lw.dayTimer.Reset(time.Duration(delayTime) * time.Second)
	}
	lw.syncLog()
	lw.filter = make(map[string]*FileWriter)
}

func (lw *LogWriter) Run(b bool) {
	defer lw.Recover()
	lw.end_wait.Add(1)
	defer lw.end_wait.Done()
	if b {
		lw.init_path = <-lw.init
	}
	lw.MakeLogTimePath()
	lw.doRun()
}

func (lw *LogWriter) doRun() {
	var logRecord *LogRecord
	var ok bool
	for {
		select {
		case logRecord, ok = <-lw.rec:
		case <-lw.dayTimer.C:
			lw.MakeLogTimePath()
			continue
		}
		if !ok || logRecord == nil {
			goto wait_close
		}
		lw.buf = lw.buf[:0]
		//_log_writer.buf = append(_log_writer.buf, "\x1b[031m"...)
		formatHeader(&lw.buf, logRecord.Level, logRecord.Created)
		lw.buf = append(lw.buf, logRecord.Message...)
		//_log_writer.buf = append(_log_writer.buf, "\x1b[0m"...)
		lw.buf = append(lw.buf, "\r\n"...)

		if logRecord.logfile {
			lw.writeFile(logRecord)
		}

		if logRecord.Level == INFO || DebugLevel() == DEBUG {
			lw.writeStd(logRecord)
		}

		logRecord.Reset()
		log_pool.Put(logRecord)
	}
wait_close:
	lw.DoClose()
}

func (lw *LogWriter) Recover() {
	if r := recover(); r != nil {
		LogError("log_manager", "log worker recover[%v]", r)
		debug.PrintStack()
		go lw.Run(false)
	}
}

func (lw *LogWriter) syncLog() {
	for _, filter := range lw.filter {
		filter.file.Sync()
		filter.file.Close()
	}
}

func (lw *LogWriter) DoClose() {
	lw.syncLog()
}
