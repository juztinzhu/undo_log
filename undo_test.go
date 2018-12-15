package main

import (
	"os"
	"testing"
)

func TestLogWrite(t *testing.T) {
	os.Remove("./test.bin")
	log := NewUndoLog("./test.bin")
	if err := log.Open(); err != nil {
		t.Error(err)
	}
	defer log.Close()
	if err := log.Write(&UndoItem{}); err != nil {
		t.Error(err)
	}
}

func TestLogRead(t *testing.T) {
	os.Remove("./test.bin")
	log := NewUndoLog("./test.bin")
	if err := log.Open(); err != nil {
		t.Error(err)
	}
	defer log.Close()
	origin := UndoItem{write, 0x99, 1, 100, 3, 0, 10}
	if err := log.Write(&origin); err != nil {
		t.Error(err)
	}

	log.Close()

	log.Open()
	item, err := log.Read()
	if err != nil {
		t.Error(err)
		return
	}
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
	if err := log.Open(); err != nil {
		t.Error(err)
	}

	defer log.Close()
	origins := []UndoItem{
		UndoItem{write, 0x1, 1, 100, 2, 0, 10},
		UndoItem{write, 0x2, 2, 100, 3, 0, 20},
		UndoItem{write, 0x3, 3, 100, 4, 0, 30},
		UndoItem{write, 0x4, 5, 100, 6, 0, 40},
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

	if *item != origins[3] {
		t.Errorf("item read does not match origin")
	}

	log.Pop()

	if err := log.Write(&origins[3]); err != nil {
		t.Error(err)
	}

	if item, err := log.Read(); err != nil {
		t.Error(err)
	} else if *item != origins[3] {
		t.Errorf("item read does not match origin")
	}

}
