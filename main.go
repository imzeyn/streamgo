package streamgo

func NewServer(on404, on403 HTTPHandler) HTTPServer{
	return HTTPServer{
		paths: serverPaths{
			temps: map[string]tempPaths{},
			static: map[string]Path{},
			regex: map[string][]rePath{},
		},
		On403: on403,
		On404: on404,
	}
}