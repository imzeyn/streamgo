package streamgo

func NewServer(addr string, on404, on403 HTTPHandler, ws *WSServer) HTTPServer {
	server := HTTPServer{Addr: addr}

	server.paths.temps  = make(map[string]tempPaths)
	server.paths.static = make(map[string]Path)
	server.paths.regex  = make(map[string][]rePath)

	server.On403 = on403
	server.On404 = on404

	if ws != nil {
		server.WebSocket = ws
	}

	return server
}