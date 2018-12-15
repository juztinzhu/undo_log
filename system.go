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

	var cashFrom, cashTo int
	var ok bool
	var userFrom, userTo *User
	if userFrom, ok = s.Users[t.FromID]; ok {
		cashFrom = userFrom.Cash
	}
	if userTo, ok = s.Users[t.ToID]; ok {
		cashTo = userTo.Cash
	}

	s.writeUndoLog(t, cashFrom, cashTo)

	userFrom.Cash = cashFrom - t.Cash
	userTo.Cash = cashTo + t.Cash

	s.commitUndoLog(t)

	if userFrom.Cash < 0 { //could check at the begnning of transaction, unless it's MVCC
		s.undo()
		return fmt.Errorf("Insufficient fund, %s with %d transfering %d", userFrom.Name, userFrom.Cash, t.Cash)
	}

	return nil
}

// writeUndoLog writes undo log to file
func (s *System) writeUndoLog(t *Transcation, fromCash int, toCash int) error {
	return s.undoLog.Write(&UndoItem{Cmd: write,
		TranscationID: t.TranscationID,
		FromID:        t.FromID,
		FromCash:      fromCash,
		ToID:          t.ToID,
		ToCash:        toCash,
		Cash:          t.Cash,
	})
}

// commitUndoLog commit the transaction & write to file
func (s *System) commitUndoLog(t *Transcation) error {
	return s.undoLog.Write(&UndoItem{Cmd: commit, TranscationID: t.TranscationID})
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
	if log.Cmd == commit {
		if err = s.undoLog.Pop(); err != nil {
			return 0, err
		}
		if log, err = s.undoLog.Read(); err != nil {
			return 0, err
		}
	}

	s.Users[log.ToID].Cash = log.ToCash
	s.Users[log.FromID].Cash = log.FromCash
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

// Close cleanup, close opened files
func (s *System) Close() {
	s.undoLog.Close()
}
