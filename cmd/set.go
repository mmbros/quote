package cmd

// set of string type
type set map[string]struct{}

func newSet(keys []string) set {
	s := set{}
	for _, k := range keys {
		s[k] = struct{}{}
	}
	return s
}

func (s set) has(key string) bool {
	_, ok := s[key]
	return ok
}

func (s set) add(key string) {
	s[key] = struct{}{}
}
