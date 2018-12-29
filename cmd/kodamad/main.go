package main

import (
	"log"

	"github.com/pankona/kodama"
)

func main() {
	cfg := &kodama.Configuration{
		Port:      50058,
		QueueLen:  10,
		Validator: &myValidator{},
	}

	s := kodama.NewServer(cfg)
	if err := s.Run(); err != nil {
		log.Printf("Kodama terminated: %v", err)
	}
}

type myValidator struct{}

func (v *myValidator) Validate(desc string) error {
	// TODO: check whether description is valid or not
	return nil
}
