package log

import (
	"bufio"
	"context"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	maxBufSize  = 50 << 20 // 50M
	maxFileSize = 500      // 500M, too large affect the buff/cache of os memory.
)

type WriteSyncer struct {
	sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc

	buff *bufio.Writer
	file *lumberjack.Logger
	life chan int
}

func newAsyncWriter(filename string) *WriteSyncer {
	ctx, cancel := context.WithCancel(context.Background())
	aw := &WriteSyncer{
		ctx:    ctx,
		cancel: cancel,
		buff:   bufio.NewWriterSize(ioutil.Discard, maxBufSize),
		file: &lumberjack.Logger{
			Filename:   filename,
			MaxSize:    maxFileSize, // megabytes
			MaxAge:     1,           // days
			MaxBackups: 1,
			LocalTime:  true,
		},
		life: make(chan int),
	}

	aw.buff.Reset(aw.file)
	go aw.autoFlushToFile()
	return aw
}

// write into the buffer, auto flush when buf is full.
func (w *WriteSyncer) Write(p []byte) (n int, err error) {
	w.Lock()
	defer w.Unlock()

	return w.buff.Write(p)
}

// Sync exec when the log Level > ErrorLevel
func (w *WriteSyncer) Sync() error {
	return w.flush()
}

func (w *WriteSyncer) Close() error {
	w.cancel()
	<-w.life
	return w.file.Close()
}

func (w *WriteSyncer) flush() error {
	w.Lock()
	defer w.Unlock()

	return w.buff.Flush()
}

// A timer to flush the buf into the file.
func (w *WriteSyncer) autoFlushToFile() {
	ticker := time.NewTicker(time.Second * 2)
	log.Printf("[log flusher for %s startup.]", w.file.Filename)
	for {
		select {
		case <-w.ctx.Done():
			ticker.Stop()
			close(w.life)
			log.Printf("[log flusher for %s exited.]", w.file.Filename)
			return
		case <-ticker.C:
			if err := w.flush(); err != nil {
				log.Printf("%s", err)
			}
		}
	}
}
