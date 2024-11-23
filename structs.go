package streamgo

import (
	"net/http"
	"regexp"

	"github.com/gorilla/websocket"
)

type HTTPServer struct {
	Addr        string
    WebSocket   *WSServer
    Middleware  []HTTPMiddlewareHandler
	paths       serverPaths
    On403       HTTPHandler
    On404       HTTPHandler
}

type WSServer struct{
    Perfix      string
    Upgrader    websocket.Upgrader
}

type serverPaths struct {
	temps  map[string]tempPaths
	static map[string]Path
	regex  map[string][]rePath
}

type Path struct {
    ID             string
	Name           string
	Handler        HTTPHandler
	WSHandler      WSHandler
    AllowedMethods map[HTTPMethod]bool
    Include        []Path
    Additional     interface{}
}

type pathDetails struct{
    ID              string
    Name            string
    Additional      interface{}
}

type rePath struct{
    RegexName   regexp.Regexp
    FullName    string
    paramList   map[string]int
    Path        Path
}

type tempPaths struct{
    Parent     string
    Path       Path
}

type HTTPRequest struct{
    HTTP         *http.Request
    PathDetails  *pathDetails
    FullName     string
    Params       *map[string]string
}

type HTTPResponse struct{
    writer  *http.ResponseWriter
}

type WSConnection struct{
    Conn    *websocket.Conn
    Err     error
}

type HTTPHandler func(request *HTTPRequest, response *HTTPResponse)
type HTTPMiddlewareHandler func(request *HTTPRequest, response *HTTPResponse) bool
type WSHandler   func(request *HTTPRequest, ws *WSConnection)
