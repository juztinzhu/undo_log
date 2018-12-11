package main

import (
	"testing"
)

func TestDoTransaction(t *testing.T) {
	s := NewSystem()

	users := make(map[int]*User)

	users[3] = &User{3, "u3", 3}
	users[2] = &User{2, "u2", 5}
	s.AddUser(users[3])
	s.AddUser(users[2])

	if err := s.DoTransaction(&Transcation{1, 3, 2, 3}); err != nil {
		t.Error("DoTransaction failed")
	}
	if users[3].Cash != 0 && users[2].Cash != 8 {
		t.Error("DoTransaction cash wrong")
	}

	if err := s.DoTransaction(&Transcation{1, 3, 2, 3}); err == nil {
		t.Error("DoTransaction failed")
	}

	if err := s.DoTransaction(&Transcation{1, 3, 2, 3}); err == nil {
		t.Error("DoTransaction failed")
	}

	if users[3].Cash != 0 || users[2].Cash != 8 {
		t.Error("DoTransaction cash wrong")
	}

	//users[1] = &User{1, "u1", 5}
	//s.AddUser(users[3])

}
