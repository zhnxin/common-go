package common

import "strings"

type StringSet struct {
	set map[string]struct{}
}

func (s *StringSet) Size() int   { return len(s.set) }
func (s *StringSet) Length() int { return s.Size() }
func (s *StringSet) Add(str string) {
	if str == "" {
		return
	}
	s.set[str] = struct{}{}
}
func (s *StringSet) ToSlice() []string {
	sslice := []string{}
	for k := range s.set {
		sslice = append(sslice, k)
	}
	return sslice
}

func (s *StringSet) JoinToString(seg string) string {
	return strings.Join(s.ToSlice(), seg)
}

func (s *StringSet) Contains(str string) bool {
	_, ok := s.set[str]
	return ok
}

func NewStringSet(init string) *StringSet {
	ss := &StringSet{
		set: make(map[string]struct{}),
	}
	if init != "" {
		ss.Add(init)
	}
	return ss
}

func (s *StringSet) Extend(sSlice []string) {
	for _, v := range sSlice {
		s.Add(v)
	}
}
