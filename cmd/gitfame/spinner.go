package main

import (
	"fmt"
	"time"
)

type jobResult[T any] struct {
	val T
	err error
}

func callAndSpin[T any](job func() (T, error)) (T, error) {
	stop := make(chan struct{}, 1)
	stopped := make(chan struct{}, 1)
	go spinner(100*time.Millisecond, stop, stopped)
	result := make(chan jobResult[T], 1)
	go func() {
		val, err := job()
		result <- jobResult[T]{val, err}
	}()
	res := <-result
	stop <- struct{}{}
	<-stopped
	return res.val, res.err
}

func spinner(delay time.Duration, stop <-chan struct{}, stopped chan<- struct{}) {
	for {
		for _, r := range `-\|/` {
			select {
			case <-stop:
				fmt.Printf("\r")
				stopped <- struct{}{}
				return
			default:
			}
			fmt.Printf("\r%c", r)
			time.Sleep(delay)
		}
	}
}
