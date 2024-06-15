package common

type Set map[string]struct{}

func NewSet() Set {
	return make(Set)
}
func (s Set) Add(value string) {
	s[value] = struct{}{}
}

func (s Set) Values() []string {
	values := make([]string, 0, len(s))
	for key := range s {
		values = append(values, key)
	}
	return values
}
