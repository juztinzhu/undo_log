package main

import (
	"bufio"
	"encoding/binary"
	"io"
	"os"
)

//UndoLog manage file read\write
type UndoLog struct {
	fileName   string
	file       *os.File
	offset     int64
	lastOffset int64 //only for read
	w          *bufio.Writer
	r          *bufio.Reader
}

//Open open file
func (l *UndoLog) Open() error {
	var err error
	l.file, err = os.OpenFile("./undo.bin", os.O_RDWR|os.O_CREATE, 0640) //TODO: consider excl
	if err != nil {
		return err
	}

	if l.offset, err = l.file.Seek(0, io.SeekEnd); err != nil {
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
	defer l.w.Reset(l.w)
	var err error
	currentOffset := l.offset
	l.offset, err = item.ToBinary(l.w, l.offset)
	if err != nil {
		return err
	}
	l.lastOffset = currentOffset
	return l.w.Flush()
}

//func (l *UndoLog) seekForRead()

// Read read file till we get a whole item
func (l *UndoLog) Read() (*UndoItem, error) {
	//l.seekForRead()
	item := UndoItem{}
	var err error
	l.lastOffset, err = item.FromBinary(l.r)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// Pop pop the last UndoItem from file
func (l *UndoLog) Pop() error {

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
// magic:4|version:4|next:4|last:4|trans:4|from:4|to:4|cash:4
// last: offset of last item. For the first item, it's 0 too.
// TODO: offset limited to 2G
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

//ToBinary write binary to writer, so far limited to 2G
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

	wint(0x006f6475)
	wint(0x00000001)
	wint(int32(lastoffset) + 32) //TODO: how to calc size?
	if lastoffset != 0 {
		wint(int32(lastoffset) - 32)
	} else {
		wint(int32(lastoffset))
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

//FromBinary read binary from reader
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

//MarshalBinary interface BinaryMarshaler,
//func (l *UndoLog) MarshalBinary() (data []byte, err error) {}

//UnmarshalBinary interface BinaryUnmarshaler
//func (l *UndoLog) UnmarshalBinary(data []byte) error {}
