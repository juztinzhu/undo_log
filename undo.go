package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

type fromToBinary interface {
	ToBinary(w io.Writer, currentOffset int64, prevOffset int64) (int64, error)
	FromBinary(r io.Reader) (int64, error)
	NextOffset() int64
	PrevOffset() int64
}

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
	var info os.FileInfo
	if info, err = l.file.Stat(); err != nil {
		return err
	}
	if info.Size() == 0 {
		l.w = bufio.NewWriter(l.file)
		l.r = bufio.NewReader(l.file)
		return l.Write(newFileHeader())
	}

	if err = l.checkIntegrity(info.Size()); err != nil {
		return errors.New("Check integrity failed:" + err.Error())
	}

	return nil
}

func (l *UndoLog) checkIntegrity(size int64) error {
	header, err := l.readHeader()
	if err != nil {
		return err
	}
	l.readOffset = header.EndingItemOffset
	item, err := l.Read()

	return nil
}

// Close close file
func (l *UndoLog) Close() { //TODO: in destructor?
	l.file.Close()
}

// Write write&flush an item to file
func (l *UndoLog) Write(item fromToBinary) error {
	l.seekForWrite()
	length, err := item.ToBinary(l.w, l.writeOffset, l.readOffset)
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

func (l *UndoLog) readHeader() (*fileHeader, error) {
	if _, err := l.file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	header := fileHeader{}
	if _, err := header.FromBinary(l.r); err != nil {
		return nil, err
	}

	if !checkFileHeader(&header) {
		return nil, errors.New("wrong magic no in header")
	}

	return &header, nil
}

// Read read file till we get a whole item, return nil if nothing to read
func (l *UndoLog) Read() (*UndoItem, error) {
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

// Pop pop the prev UndoItem from file
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
	//start cmdType = iota
	write cmdType = iota
	commit
	//abort
)

// magic:4|version:4|next:4|endItem:4|
type fileHeader struct {
	Magic            int
	Version          int
	NextItemOffset   int64
	EndingItemOffset int64
}

//Next offset of next item
func (t *fileHeader) Next() int64 {
	return t.NextItemOffset
}

//Prev offset of previous item
func (t *fileHeader) Prev() int64 {
	return -1
}

const constMAGIC int = 0x006f6475
const constVERSION int = 1

func newFileHeader() *fileHeader {
	return &fileHeader{Magic: constMAGIC, Version: constVERSION, NextItemOffset: 16}
}

func checkFileHeader(header *fileHeader) bool {
	return header.Magic == constMAGIC
}

// UndoItem undo log implementation
// cmd:4|next:4|prev:4|trans:4|from:4|fromcash:4|to:4|tocash:4|cash:4
// prev: writeOffset of prev item. For the first item, it's -1
// TODO: writeOffset limited to 2G so far
type UndoItem struct {
	Cmd           cmdType
	TranscationID int
	FromID        int
	FromCash      int // from-user 's cash when transaction begin
	ToID          int
	ToCash        int // to-user 's cash when transaction begin
	Cash          int
	next          int
	prev          int
}

//Next offset of next item
func (t *UndoItem) Next() int64 {
	return int64(t.next)
}

//Prev offset of previous item
func (t *UndoItem) Prev() int64 {
	return int64(t.prev)
}

// ToBinary write binary to writer, so far limited to 2G. return length of this item.
func (t *UndoItem) ToBinary(w io.Writer, currentOffset int64, prevOffset int64) (int64, error) {
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

	itemLength := 36

	if t.Cmd == commit {
		itemLength = 16 //TODO: how to calc size?
	}

	wint(int32(t.Cmd))                             //cmd
	wint(int32(currentOffset) + int32(itemLength)) //next
	wint(int32(prevOffset))                        //prev For the first item, it's -1
	wint(int32(t.TranscationID))
	if t.Cmd != commit { //commit events do not need those values
		wint(int32(t.FromID))
		wint(int32(t.FromCash))
		wint(int32(t.ToID))
		wint(int32(t.ToCash))
		wint(int32(t.Cash))
	}
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

	var cmd int
	var next int
	var prev int
	rint(&cmd)
	t.Cmd = cmdType(cmd)
	rint(&next)
	rint(&prev)
	rint(&t.TranscationID)
	if t.Cmd != commit {
		rint(&t.FromID)
		rint(&t.FromCash)
		rint(&t.ToID)
		rint(&t.ToCash)
		rint(&t.Cash)
	}
	if pErr != nil {
		return 0, *pErr
	}
	return int64(prev), nil
}

func (h *fileHeader) ToBinary(w io.Writer, currentOffset int64, prevOffset int64) (int64, error) {
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

	itemLength := 16
	wint(int32(h.Magic))            //magic
	wint(int32(h.Version))          //version
	wint(int32(itemLength))         //next
	wint(int32(h.EndingItemOffset)) //ending item offset

	if pErr != nil {
		return int64(length), *pErr
	}
	return int64(length), nil
}

// FromBinary read binary from reader, return writeOffset of the item before the one being read.
func (h *fileHeader) FromBinary(r io.Reader) (int64, error) {
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

	var next int
	var ending int
	rint(&h.Magic)
	rint(&h.Version)
	rint(&next)
	rint(&ending)
	h.NextItemOffset = int64(next)
	h.EndingItemOffset = int64(ending)

	if pErr != nil {
		return -1, *pErr
	}
	return int64(-1), nil
}
