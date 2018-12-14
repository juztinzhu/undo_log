package main

import "testing"

func TestLogWrite(t *testing.T) {
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

	log := NewUndoLog("./test.bin")
	if err := log.Open(); err != nil {
		t.Error(err)
	}
	defer log.Close()
	origin := UndoItem{start, 0x99, 1, 3, 10}
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
	log := NewUndoLog("./test.bin")
	if err := log.Open(); err != nil {
		t.Error(err)
	}

	defer log.Close()
	origins := []UndoItem{
		UndoItem{start, 0x1, 1, 2, 10},
		UndoItem{start, 0x2, 2, 3, 20},
		UndoItem{start, 0x3, 3, 4, 30},
		UndoItem{start, 0x4, 5, 6, 40},
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
