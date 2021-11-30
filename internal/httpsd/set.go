package httpsd

var exists = struct{}{}

type set struct {
	m map[string]struct{}
}

func NewSet(values []string) *set {
	s := &set{}
	s.m = make(map[string]struct{})
	if len(values) != 0 {
		for _, v := range values {
			s.m[v] = exists
		}
	}
	return s
}

func (s *set) Add(values []string) {
	for _, value := range values {
		s.m[value] = exists
	}
}

func (s *set) Remove(values []string) {
	for _, value := range values {
		delete(s.m, value)
	}
}

func (s *set) Contains(value string) bool {
	_, c := s.m[value]
	return c
}

func (s *set) Array() []string {
	keys := []string{}
	for k := range s.m {
		keys = append(keys, k)
	}
	return keys
}
