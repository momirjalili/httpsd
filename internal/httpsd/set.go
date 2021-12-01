package httpsd

var exists = struct{}{}

type set struct {
	m map[string]struct{}
}

func NewSet(values []Target) *set {
	s := &set{}
	s.m = make(map[string]struct{})
	if len(values) != 0 {
		for _, v := range values {
			s.m[v.Addr] = exists
		}
	}
	return s
}

func (s *set) Add(values []Target) {
	for _, value := range values {
		s.m[value.Addr] = exists
	}
}

func (s *set) Remove(values []Target) {
	for _, value := range values {
		delete(s.m, value.Addr)
	}
}

func (s *set) Contains(value Target) bool {
	_, c := s.m[value.Addr]
	return c
}

func (s *set) Array() []Target {
	keys := []Target{}
	for k := range s.m {
		keys = append(keys, Target{Addr: k})
	}
	return keys
}
