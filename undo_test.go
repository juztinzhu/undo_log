package main

import (
	"os"
	"testing"
)

func TestLogWrite(t *testing.T) {
	os.Remove("./test.bin")
	log := NewUndoLog("./test.bin")
	defer log.Close()
	if err := log.Write(&UndoItem{}); err != nil {
		t.Error(err)
	}
}

func TestLogRead(t *testing.T) {
	os.Remove("./test.bin")
	log := NewUndoLog("./test.bin")
	//defer log.Close()
	origin := UndoItem{write, 0x99, 1, 100, 3, 0, 10, 0, 0}
	if err := log.Write(&origin); err != nil {
		t.Error(err)
	}

	log.Close()

	//log.Open()
	log = NewUndoLog("./test.bin")
	defer log.Close()
	item, err := log.Read()
	if err != nil {
		t.Error(err)
		return
	}
	item.next = 0
	item.prev = 0
	if *item != origin {
		t.Errorf("item read does not match origin")
	}

	if err := log.Pop(); err != nil {
		t.Error(err)
	}
}

func TestLogWriteRead(t *testing.T) {
	os.Remove("./test.bin")
	log := NewUndoLog("./test.bin")
	defer log.Close()
	origins := []UndoItem{
		UndoItem{write, 0x1, 1, 100, 2, 0, 10, 0, 0},
		UndoItem{write, 0x2, 2, 100, 3, 0, 20, 0, 0},
		UndoItem{write, 0x3, 3, 100, 4, 0, 30, 0, 0},
		UndoItem{write, 0x4, 5, 100, 6, 0, 40, 0, 0},
	}

	for _, item := range origins {
		if err := log.Write(&item); err != nil {
			t.Error(err)
		}
	}

	for idx := range origins {
		item, err := log.Read()
		if err != nil {
			t.Error(err)
			return
		}

		//take care of internal field, they dont need to be the same
		item.prev = 0
		item.next = 0
		if *item != origins[len(origins)-idx-1] {
			t.Errorf("item read does not match origin")
		}

		if err := log.Pop(); err != nil {
			t.Error(err)
		}
	}

	for _, item := range origins {
		if err := log.Write(&item); err != nil {
			t.Error(err)
		}
	}

	item, err := log.Read()
	if err != nil {
		t.Error(err)
		return
	}

	item.next = 0
	item.prev = 0
	if *item != origins[3] {
		t.Errorf("item read does not match origin")
	}

	log.Pop()

	if err := log.Write(&origins[3]); err != nil {
		t.Error(err)
	}

	if item, err := log.Read(); err != nil {
		t.Error(err)
	} else {
		item.next = 0
		item.prev = 0
		if *item != origins[3] {
			t.Errorf("item read does not match origin")
		}
	}

}

func TestLogFileHeader(t *testing.T) {
	os.Remove("./test.bin")
	log := NewUndoLog("./test.bin")
	origins := []UndoItem{
		UndoItem{write, 0x1, 1, 100, 2, 0, 10, 0, 0},
		UndoItem{write, 0x2, 2, 100, 3, 0, 20, 0, 0},
		UndoItem{write, 0x3, 3, 100, 4, 0, 30, 0, 0},
		UndoItem{write, 0x4, 5, 100, 6, 0, 40, 0, 0},
	}

	for _, item := range origins {
		if err := log.Write(&item); err != nil {
			t.Error(err)
		}
	}

	log.Close()
	log = NewUndoLog("./test.bin")
	if info, _ := log.file.Stat(); log.header.Size != info.Size() {
		t.Error("endingItemOffset does not match size")
	}

	another := &UndoItem{write, 0x5, 6, 100, 7, 0, 40, 0, 0}
	log.Write(another)
	log.file.Close()

	log = NewUndoLog("./test.bin")
	if info, _ := log.file.Stat(); log.header.Size != info.Size() {
		t.Error("endingItemOffset does not match size after recover")
	}
	log.Close()
}
