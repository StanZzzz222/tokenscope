package common

import "sync"

/*
   Created by zyx
   Date Time: 2025/9/19
   File: wg.go
*/

type Task func()

type WaitGroup struct {
	wg  *sync.WaitGroup
	sem chan struct{}
}

type IWaitGroup interface {
	Go(task Task)
	Wait()
}

func Wg(sem int) IWaitGroup {
	if sem <= 0 {
		sem = 1
	}
	return &WaitGroup{
		wg:  &sync.WaitGroup{},
		sem: make(chan struct{}, sem),
	}
}

func (w *WaitGroup) Go(task Task) {
	w.wg.Add(1)
	go func() {
		w.sem <- struct{}{}
		defer func() {
			<-w.sem
			w.wg.Done()
		}()
		task()
	}()
}

func (w *WaitGroup) Wait() {
	w.wg.Wait()
}
