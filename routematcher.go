package streamgo

type RouteMatcher[Payload any] struct {
	Static map[string]*Path[Payload]
	Regex  map[string]*RouterRegex[Payload]
}

type RouterRegex[Payload any] struct {
	ListLen        int
	List           []*Path[Payload]
	SplitedList    [][]*Path[Payload]
	SplitedListLen int
}
