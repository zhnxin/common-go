package common

import (
	"context"
	"time"

	"github.com/emirpasic/gods/lists/singlylinkedlist"
)

type timeSpot struct {
	t time.Time
	v interface{}
}

func NewSchedule() *Schedule {
	ctx, cannel := context.WithCancel(context.Background())
	s := &Schedule{
		list:      singlylinkedlist.New(),
		spotChan:  make(chan *timeSpot),
		valueChan: make(chan interface{}),
		ctx:       ctx, cannel: cannel,
	}
	go s.run()
	return s
}

type Schedule struct {
	list      *singlylinkedlist.List
	spotChan  chan *timeSpot
	valueChan chan interface{}
	ctx       context.Context
	cannel    context.CancelFunc
	runCtx    context.Context
}

func (s *Schedule) Add(t time.Time, v interface{}) {
	s.spotChan <- &timeSpot{t: t, v: v}
}

func (s *Schedule) Stop() {
	s.cannel()
}

func (s *Schedule) Clear() {
	s.list.Clear()
}

func (s *Schedule) Done() <-chan struct{} {
	return s.runCtx.Done()
}

func (s *Schedule) Chan() <-chan interface{} {
	return s.valueChan
}

func (s *Schedule) GetNextTime() (time.Time, bool) {
	return s.first()
}

func (s *Schedule) GetSchedule() []time.Time {
	times := make([]time.Time, s.list.Size())
	for i, t := range s.list.Values() {
		times[i] = t.(*timeSpot).t
	}
	return times
}

func (s *Schedule) add(spot *timeSpot) {
	for i := 0; i < s.list.Size(); i++ {
		item, _ := s.list.Get(i)
		if !item.(*timeSpot).t.Before(spot.t) {
			s.list.Insert(i, spot)
			return
		}
	}
	s.list.Add(spot)
}

func (s *Schedule) first() (time.Time, bool) {
	item, ok := s.list.Get(0)
	if !ok {
		return time.Time{}, ok
	}
	return item.(*timeSpot).t, ok
}

func (s *Schedule) FirstSpot() (*timeSpot, bool) {
	item, ok := s.list.Get(0)
	if !ok {
		return nil, ok
	}
	return item.(*timeSpot), ok
}

func (s *Schedule) Remove(v interface{}) {
	s.spotChan <- &timeSpot{v: v}
}

func (s *Schedule) remove(v interface{}) {
	it := s.list.Iterator()
	for it.Next() {
		index, value := it.Index(), it.Value()
		if value.(*timeSpot).v == v {
			s.list.Remove(index)
		}
	}
}

func (s *Schedule) run() {
	defer s.cannel()
	defer close(s.valueChan)
	var nextSpot time.Time
	var cannel context.CancelFunc
	s.runCtx, cannel = context.WithCancel(context.Background())
	defer cannel()
	for {
		if nextSpot.IsZero() {
			select {
			case <-s.ctx.Done():
				return
			case spot := <-s.spotChan:
				if spot.t.IsZero() {
					s.remove(spot.v)
				} else {
					s.add(spot)
				}
			}
		} else {
			select {
			case <-s.ctx.Done():
				return
			case spot := <-s.spotChan:
				if spot.t.IsZero() {
					s.remove(spot.v)
				} else {
					s.add(spot)
				}
			case <-time.After(time.Until(nextSpot)):
				for {
					spot, ok := s.FirstSpot()
					if !ok || spot.t.After(nextSpot) {
						break
					} else {
						s.list.Remove(0)
						go func() { s.valueChan <- spot.v }()
					}
				}
			}
		}
		nextSpot, _ = s.first()
	}
}
