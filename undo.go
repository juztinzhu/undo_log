package main

import (
	"bufio"
	"encoding/binary"
	"io"
	"os"
)

// UndoLog manage file read\write
// TODO: log file rotate, maybe add an index file
type UndoLog struct {
	fileName    string
	file        *os.File
	writeOffset int64
	readOffset  int64 //only for read
	prevOffset  int64 //offset of previous item to be read.
	w           *bufio.Writer
	r           *bufio.Reader
}

// NewUndoLog create log with filename
func NewUndoLog(name string) *UndoLog {
	u := &UndoLog{fileName: name}
	u.Open()
	return u
}

// Open open file
func (l *UndoLog) Open() error {
	var err error
	l.file, err = os.OpenFile(l.fileName, os.O_RDWR|os.O_CREATE, 0640) //TODO: consider excl
	if err != nil {
		return err
	}

	if l.writeOffset, err = l.file.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	l.w = bufio.NewWriter(l.file)
	l.r = bufio.NewReader(l.file)

	return nil
}

// Close close file
func (l *UndoLog) Close() { //TODO: in destructor?
	l.file.Close()
}

// Write write&flush an item to file
func (l *UndoLog) Write(item *UndoItem) error {
	//defer l.w.Reset(l.w)
	l.seekForWrite()
	length, err := item.ToBinary(l.w, l.writeOffset)
	l.readOffset = l.writeOffset
	l.writeOffset += length
	if err != nil {
		return err
	}
	err = l.w.Flush()
	return err
}

func (l *UndoLog) seekForWrite() {
	l.writeOffset, _ = l.file.Seek(0, io.SeekEnd)
}
func (l *UndoLog) seekForRead() bool {
	if l.readOffset == -1 {
		return false
	}
	if _, err := l.file.Seek(l.readOffset, io.SeekStart); err != nil {
		return false
	}
	return true
}
func (l *UndoLog) trunc(pos int64) error {
	return l.file.Truncate(pos)
}

// Read read file till we get a whole item, return nil if nothing to read
func (l *UndoLog) Read() (*UndoItem, error) {
	defer l.r.Reset(l.file)
	if !l.seekForRead() {
		return nil, nil
	}
	item := UndoItem{}
	var err error
	l.prevOffset, err = item.FromBinary(l.r)

	if err != nil {
		return nil, err
	}
	return &item, nil
}

// Pop pop the last UndoItem from file
func (l *UndoLog) Pop() error {
	if err := l.trunc(l.readOffset); err != nil {
		return err
	}
	l.writeOffset = l.readOffset
	l.readOffset = l.prevOffset
	return nil
}

type cmdType = byte

const (
	start cmdType = iota
	write
	commit
	abort
)

// UndoItem undo log implementation
// magic:4|version:4|next:4|prev:4|trans:4|from:4|to:4|cash:4
// prev: writeOffset of last item. For the first item, it's -1
// TODO: writeOffset limited to 2G
type UndoItem struct {
	Cmd           cmdType
	TranscationID int
	FromID        int
	//FromValue     int
	//OldFromValue  int
	ToID int
	//ToValue       int
	//OldToValue    int
	Cash int
}

// ToBinary write binary to writer, so far limited to 2G. return length of this item.
func (t *UndoItem) ToBinary(w io.Writer, lastoffset int64) (int64, error) {
	var pErr *error
	var length int
	wint := func(v int32) {
		if pErr != nil {
			return
		}

		if err := binary.Write(w, binary.LittleEndian, v); err != nil {
			pErr = &err
			return
		}
		length += 4
	}

	wint(0x006f6475)             //magic
	wint(0x00000001)             //version
	wint(int32(lastoffset) + 32) //next //TODO: how to calc size?
	if lastoffset != 0 {         //prev For the first item, it's -1
		wint(int32(lastoffset) - 32)
	} else {
		wint(int32(-1))
	}
	wint(int32(t.TranscationID))
	wint(int32(t.FromID))
	wint(int32(t.ToID))
	wint(int32(t.Cash))
	if pErr != nil {
		return int64(length), *pErr
	}
	return int64(length), nil
}

// FromBinary read binary from reader, return writeOffset of the item before the one being read.
func (t *UndoItem) FromBinary(r io.Reader) (int64, error) {
	var pErr *error
	//var length int
	rint := func(p *int) {
		if pErr != nil {
			return
		}
		var value int32
		if err := binary.Read(r, binary.LittleEndian, &value); err != nil {
			pErr = &err
			return
		}
		*p = int(value)
	}

	var magic int
	var version int
	var next int
	var last int
	rint(&magic)
	rint(&version)
	rint(&next)
	rint(&last)
	rint(&t.TranscationID)
	rint(&t.FromID)
	rint(&t.ToID)
	rint(&t.Cash)
	if pErr != nil {
		return 0, *pErr
	}
	return int64(last), nil
}
