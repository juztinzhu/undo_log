package main

import (
	"fmt"
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

func TestConcurrent(t *testing.T) {
	s := NewSystem()

	const COUNT = 1000000
	var users = [COUNT]*User{nil}
	for idx, user := range users {
		username := fmt.Sprintf("user_%d", idx)
		user = &User{idx, username, 10}
		s.AddUser(user)
	}

	ch := make(chan int)
	defer close(ch)

	tryTrans := func(ch chan int, id *int, from int, to int, cash int) {
		for true {
			if err := s.DoTransaction(&Transcation{*id, from, to, cash}); err != nil {
				fmt.Printf("Transaction failed from %d to %d with %d", from, to, cash)
				*id++
				ch <- 1
			} else {
				break
			}
		}
	}

	go func(ch chan int) {
		id := 0
		for i := 0; i < COUNT/2; i++ {
			tryTrans(ch, &id, i, i+COUNT/2, 5)
			id++
		}
		fmt.Printf("maxid %d\n", id)
		ch <- 0
	}(ch)

	go func(ch chan int) {
		id := COUNT * 8
		for i := COUNT - 1; i >= COUNT/2; i-- {
			tryTrans(ch, &id, i, i-COUNT/2, 5)
			id++
		}
		fmt.Printf("maxid %d\n", id)
		ch <- 0
	}(ch)

	completeCount := 0
	for true {
		var x int
		x = <-ch
		switch x {
		case 0:
			completeCount++
			if completeCount == 2 {
				return
			}
		case 1:
			continue
		}
	}

	for _, user := range users {
		if user.Cash != 10 {
			t.Errorf("%s cash is %d, not 10", user.Name, user.Cash)
		}
	}

	s.UndoTranscation(0)

	for _, user := range users {
		if user.Cash != 10 {
			t.Errorf("%s cash is %d, not 10", user.Name, user.Cash)
		}
	}

}

func TestUndoTransaction(t *testing.T) {
	s := NewSystem()

	users := make(map[int]*User)

	users[3] = &User{3, "u3", 9}
	users[2] = &User{2, "u2", 5}
	s.AddUser(users[3])
	s.AddUser(users[2])

	transcations := []*Transcation{
		{1, 3, 2, 3},
		{2, 3, 2, 3},
		{3, 3, 2, 3},
		{4, 3, 2, 3},
		{5, 3, 2, 3},
	}

	for _, trans := range transcations {
		if err := s.DoTransaction(trans); err != nil {
		}
	}

	s.UndoTranscation(3)
	if users[3].Cash != 3 {
		t.Error("first undo failed")
	}

	s.UndoTranscation(1)
	if users[3].Cash != 9 {
		t.Error("undo twice failed")
	}
	if users[2].Cash != 5 {
		t.Error("after undo, u2 not revert")
	}

}
