package main

import (
	"log"
)

func main() {
	system := NewSystem()
	defer system.Close()

	users := []*User{
		{
			ID:   1,
			Name: "Tom",
			Cash: 10,
		},
		{
			ID:   2,
			Name: "Jerry",
			Cash: 10,
		},
		{
			ID:   3,
			Name: "Spike",
			Cash: 10,
		},
	}
	transcations := []*Transcation{
		{
			TranscationID: 1,
			FromID:        1,
			ToID:          2,
			Cash:          10,
		},
		{
			TranscationID: 2,
			FromID:        2,
			ToID:          3,
			Cash:          5,
		},
		{
			TranscationID: 3,
			FromID:        3,
			ToID:          1,
			Cash:          20,
		},
		{
			TranscationID: 4,
			FromID:        2,
			ToID:          1,
			Cash:          10,
		},
	}

	for _, user := range users {
		if err := system.AddUser(user); err != nil {
			log.Printf("add user failed %v", err)
		}
	}

	// TODO: do transcation parallel
	for _, transcation := range transcations {
		if err := system.DoTransaction(transcation); err != nil {
			log.Printf("do transcation failed %v", err)
		}
	}

	for _, user := range system.Users {
		log.Printf("after transcation, %s has %d money", user.Name, user.Cash)
	}

	if err := system.UndoTranscation(2); err != nil {
		log.Printf("undo transcation failed %v", err)
	}

	for _, user := range system.Users {
		log.Printf("after undo transcation, %s has %d money", user.Name, user.Cash)
	}
}
