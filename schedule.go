package common

import (
	"context"
	"time"

	"github.com/emirpasic/gods/lists/singlylinkedlist"
)

type TimeSpot struct {
	T    time.Time
	Item interface{}
}

func NewSchedule() *Schedule {
	ctx, cannel := context.WithCancel(context.Background())
	s := &Schedule{
		list:      singlylinkedlist.New(),
		spotChan:  make(chan *TimeSpot),
		valueChan: make(chan interface{}),
		ctx:       ctx, cannel: cannel,
	}
	go s.run()
	return s
}

type Schedule struct {
	list      *singlylinkedlist.List
	spotChan  chan *TimeSpot
	valueChan chan interface{}
	ctx       context.Context
	cannel    context.CancelFunc
	runCtx    context.Context
}

func (s *Schedule) Add(t time.Time, v interface{}) {
	s.spotChan <- &TimeSpot{T: t, Item: v}
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

func (s *Schedule) GetSchedule() []TimeSpot {
	times := make([]TimeSpot, s.list.Size())
	for i, t := range s.list.Values() {
		times[i] = TimeSpot{
			T:    t.(*TimeSpot).T,
			Item: t.(*TimeSpot).Item,
		}
	}
	return times
}

func (s *Schedule) add(spot *TimeSpot) {
	for i := 0; i < s.list.Size(); i++ {
		item, _ := s.list.Get(i)
		if !item.(*TimeSpot).T.Before(spot.T) {
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
	return item.(*TimeSpot).T, ok
}

func (s *Schedule) FirstSpot() (*TimeSpot, bool) {
	item, ok := s.list.Get(0)
	if !ok {
		return nil, ok
	}
	return item.(*TimeSpot), ok
}

func (s *Schedule) Remove(v interface{}) {
	s.spotChan <- &TimeSpot{Item: v}
}

func (s *Schedule) remove(v interface{}) {
	it := s.list.Iterator()
	for it.Next() {
		index, value := it.Index(), it.Value()
		if value.(*TimeSpot).Item == v {
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
	freshTicker := time.NewTicker(time.Hour)
	for {
		if nextSpot.IsZero() {
			select {
			case <-s.ctx.Done():
				return
			case spot := <-s.spotChan:
				if spot.T.IsZero() {
					s.remove(spot.Item)
				} else {
					s.add(spot)
				}
			}
		} else {
			ctx, cannel := context.WithDeadline(context.Background(), nextSpot)
			select {
			case <-s.ctx.Done():
				cannel()
				return
			case spot := <-s.spotChan:
				cannel()
				if spot.T.IsZero() {
					s.remove(spot.Item)
				} else {
					s.add(spot)
				}
			case <-ctx.Done():
				cannel()
				for {
					spot, ok := s.FirstSpot()
					if !ok || spot.T.After(nextSpot) {
						break
					} else {
						s.list.Remove(0)
						go func() { s.valueChan <- spot.Item }()
					}
				}
			case <-freshTicker.C:
			}
		}
		nextSpot, _ = s.first()
	}
}
