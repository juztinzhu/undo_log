package main

import (
	"errors"
	"fmt"
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

// System keeps the user and transcation information
type System struct {
	sync.RWMutex
	Users        map[int]*User
	Transcations []*Transcation
	undoLog      *UndoLog
}

// NewSystem returns a System
func NewSystem() *System {
	return &System{
		Users:        make(map[int]*User),
		Transcations: make([]*Transcation, 0, 10),
		undoLog:      NewUndoLog("./undo.bin"),
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

	s.Transcations = append(s.Transcations, t)

	s.writeUndoLog(t)

	var cashFrom, cashTo int
	var ok bool
	var userFrom, userTo *User
	if userFrom, ok = s.Users[t.FromID]; ok {
		cashFrom = userFrom.Cash
		userFrom.Cash = cashFrom - t.Cash

	}
	if userTo, ok = s.Users[t.ToID]; ok {
		cashTo = userTo.Cash
		userTo.Cash = cashTo + t.Cash
	}

	if userFrom.Cash < 0 {
		s.undo()
		return fmt.Errorf("Insufficient fund, %s with %d transfering %d", userFrom.Name, userFrom.Cash, t.Cash)
	}

	return nil
}

// writeUndoLog writes undo log to file
func (s *System) writeUndoLog(t *Transcation) error {
	return s.undoLog.Write(&UndoItem{write,
		t.TranscationID,
		t.FromID,
		t.ToID,
		t.Cash,
	})
}

// gcUndoLog the old undo log
func (s *System) gcUndoLog() {
	// TODO: implement gcUndoLog
}

func (s *System) undo() (int, error) {
	log, err := s.undoLog.Read()
	if err != nil {
		return 0, err
	}
	s.Users[log.ToID].Cash -= log.Cash
	s.Users[log.FromID].Cash += log.Cash
	if err = s.undoLog.Pop(); err != nil {
		return 0, err
	}
	return log.TranscationID, nil

}

// UndoTranscation roll back some transcations
func (s *System) UndoTranscation(fromID int) error {
	// TODO: implement UndoTranscation
	// undo transcation from fromID to the last transcation

	s.Lock()
	defer s.Unlock()

	for true {
		tid, err := s.undo()
		if err != nil {
			return err
		}
		if tid == fromID {
			//TODO: what if fromID does not exist
			break
		}
	}
	return nil
}

// Close cleanup
func (s *System) Close() {
	s.undoLog.Close()
}
