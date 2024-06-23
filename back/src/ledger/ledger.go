package ledger

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"
)

type Ledger struct {
	Opts       *LedgerOptions
	recordChan chan []string
	done       chan struct{}
	file       *os.File
	writer     *csv.Writer
}

type LedgerOptions struct {
	Path           string
	PathTs         bool
	RecordChanSize int64
	FlushCount     int64
	MaxRecords     int64
}

// NewLedger creates a new ledger
func NewLedger(opts *LedgerOptions) *Ledger {
	if opts.RecordChanSize == 0 {
		opts.RecordChanSize = 100
	}
	return &Ledger{
		Opts:       opts,
		recordChan: make(chan []string, opts.RecordChanSize),
		done:       make(chan struct{}),
	}
}

// Put record into the ledger's channel
func (l *Ledger) PutRecord(record []string) {
	l.recordChan <- record
}

// Close closes the ledger
func (l *Ledger) Close() {
	close(l.done)
}

// [re-]create and return writer
func (l *Ledger) createWriter() error {

	var fileName string
	if l.Opts.PathTs {
		fileName = fmt.Sprintf("%s-%d.csv", l.Opts.Path, time.Now().Unix())
	} else {
		fileName = l.Opts.Path
	}

	log.Printf("Openning ledger file: %s\n", fileName)
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer l.file.Close()

	l.writer = csv.NewWriter(file)
	defer l.writer.Flush()

	return nil
}

// read from channel and dump records to file
func (l *Ledger) Start() {

	var lines int64 = 0
	var err error

	for {
		select {
		case record := <-l.recordChan:
			if lines == 0 && lines%l.Opts.MaxRecords == 0 {
				err = l.createWriter()
				if err != nil {
					log.Printf("Error creating writer: %v\n", err)
					return
				}
			}
			err = l.writer.Write(record)
			if err != nil {
				log.Printf("Error writing record to file: %v\n", err)
				return
			}
			lines++
			if l.Opts.FlushCount > 0 && lines%l.Opts.FlushCount == 0 {
				l.writer.Flush()
				l.file.Sync()
			}
		case <-l.done:
			l.writer.Flush()
			l.file.Sync()
			l.file.Close()
			return
		}
	}
}
