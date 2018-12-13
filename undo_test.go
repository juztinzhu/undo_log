package main

import "testing"

func TestWrite(t *testing.T) {
	log := UndoLog{}
	if err := log.Open(); err != nil {
		t.Error(err)
	}
	if err := log.Write(&UndoItem{}); err != nil {
		t.Error(err)
	}
	log.Close()
}

func TestWriteRead(t *testing.T) {
	log := UndoLog{}
	if err := log.Open(); err != nil {
		t.Error(err)
	}
	origin := UndoItem{write, 0x99, 1, 3, 10}
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

}
