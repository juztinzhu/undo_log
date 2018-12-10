package main

import (
	"errors"
	"sync"
)

// User saves user's information
type User struct {
	ID   int
	Name string
	Cash int
}

// Transcation record a transcation.
type Transcation struct {
	TranscationID int
	FromID        int
	ToID          int
	Cash          int
}

type cmdType = byte

const (
	start cmdType = iota
	write
	commit
	abort
)

type undoLog struct {
	Cmd           cmdType
	TranscationID int
	FromUserID    int
	FromValue     int
	OldFromValue  int
	ToUserID      int
	ToValue       int
	OldToValue    int
}

// System keeps the user and transcation information
type System struct {
	sync.RWMutex

	Users map[int]*User

	Transcations []*Transcation

	// TODO: add some variables about undo log
	UndoLogs []*undoLog
}

// NewSystem returns a System
func NewSystem() *System {
	return &System{
		Users:        make(map[int]*User),
		Transcations: make([]*Transcation, 0, 10),
		UndoLogs:     make([]*undoLog, 0, 10),
	}
}

// AddUser adds a new user to the system
func (s *System) AddUser(u *User) error {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.Users[u.ID]; ok {
		return errors.New("user id is already exists")
	}

	s.Users[u.ID] = u

	return nil
}

// DoTransaction applys a transaction
func (s *System) DoTransaction(t *Transcation) error {
	// TODO: implement DoTransaction
	// if after this transcation, user's cash is less than zero,
	// rollback this transcation according to undo log.
	s.Lock()
	defer s.Unlock()

	_ = append(s.UndoLogs, &undoLog{start, 0, 0, 0, 0})
	var cashFrom, cashTo int
	//var userFrom,userTo *User
	if userFrom, ok := s.Users[t.FromID]; ok {
		cashFrom = userFrom.Cash

		userFrom.Cash = cashFrom - t.Cash

	}
	if userTo, ok := s.Users[t.ToID]; ok {
		cashTo = userTo.Cash

		//TODO: write file
		userTo.Cash = cashTo + t.Cash
	}

	return nil
}

// writeUndoLog writes undo log to file
func (s *System) writeUndoLog(t *Transcation) error {
	// TODO: implement writeUndoLog
	_ = append(s.UndoLogs, &undoLog{write,
		t.TranscationID,
		t.FromID,
		cashFrom - t.Cash,
		cashFrom,
	})

	_ = append(s.UndoLogs, &undoLog{write,
		t.TranscationID,
		t.ToID,
		cashTo + t.Cash,
		cashTo,
	})
	return nil
}

// gcUndoLog the old undo log
func (s *System) gcUndoLog() {
	// TODO: implement gcUndoLog
}

// UndoTranscation roll back some transcations
func (s *System) UndoTranscation(fromID int) error {
	// TODO: implement UndoTranscation
	// undo transcation from fromID to the last transcation

	return nil
}
