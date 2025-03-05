package streamgo

func EnableMethods(methods ...string) map[HTTPMethod]bool {
	m := map[HTTPMethod]bool{}
	for _, v := range methods {
		m[HTTPMethod(v)] = true
	}
	return m
}
