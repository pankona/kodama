package main

import (
	"log"

	"github.com/pankona/kodama"
)

func main() {
	cfg := &kodama.Configuration{
		Port:      50058,
		QueueLen:  10,
		WorkerNum: 3,
		RetryNum:  5,
		Validator: &myValidator{},
		Worker:    &myWorker{},
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

type myWorker struct{}

func (w *myWorker) Work(desc string) error {
	// TODO: do task that takes long time to complete
	return nil
}
