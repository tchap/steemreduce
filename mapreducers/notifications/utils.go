package notifications

type StringSet map[string]struct{}

func (set StringSet) Contains(v string) bool {
	_, ok := set[v]
	return ok
}

func MakeStringSet(values []string) StringSet {
	set := make(map[string]struct{}, len(values))
	for _, v := range values {
		set[v] = struct{}{}
	}
	return StringSet(set)
}
