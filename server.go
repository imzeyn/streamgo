package streamgo

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"sync"

	"strings"

	"github.com/gorilla/websocket"
)

type Server[Payload any] struct {
	Paths            RouteMatcher[Payload]
	RegexOptions     *RegexOptions
	HTTPHandle404    func(request *HTTPRequest, response *HTTPResponse, payload Payload)
	HTTPHandle405    func(request *HTTPRequest, response *HTTPResponse, payload Payload)
	HTTPHandler      func(request *HTTPRequest, response *HTTPResponse, payload Payload)
	WebSocketHandler func(request *HTTPRequest, response *HTTPResponse, payload Payload, upgrader *websocket.Upgrader)
}

func NewServer[PayloadType any](regexOpts RegexOptions) Server[PayloadType] {
	if regexOpts.paramPatternOptional == nil || regexOpts.paramPatternRequired == nil {
		regexOpts = NewRegexOptions(regexOpts.ParallelSearchCount)
	}
	return Server[PayloadType]{
		RegexOptions: &regexOpts,
	}
}

func (s *Server[PayloadType]) BuildPaths(paths []Path[PayloadType], perfix string) {
	var fullname strings.Builder
	var usednames map[string]bool = map[string]bool{}

	for i := 0; i < len(paths); i++ {
		fullname.WriteString(perfix)
		fullname.WriteString(paths[i].Name)

		name := ClearURL(fullname.String())
		if paths[i].Include != nil {
			s.BuildPaths(paths[i].Include, name)
		}

		fullname.Reset()
		paths[i].NormalizeMethods()
		if s.RegexOptions.IsParamURL(name) {
			perfix := s.RegexOptions.GetPerfix(name)
			unPerfixed := name[len(perfix):]

			paths[i].mappedParams = s.RegexOptions.ParseParamNames(unPerfixed)

			var regexName strings.Builder

			for _, v := range strings.Split(name, "/") {
				regexName.WriteString(s.RegexOptions.ReplaceForFind(v))
				regexName.WriteString("/")
			}

			fullName := "^" + ClearURL(regexName.String()) + "$"
			paths[i].regexName = regexp.MustCompile(fullName)

			if s.Paths.Regex == nil {
				s.Paths.Regex = map[string]*RouterRegex[PayloadType]{}
			}

			if _, ok := s.Paths.Regex[perfix]; !ok {
				s.Paths.Regex[perfix] = &RouterRegex[PayloadType]{
					List:        []*Path[PayloadType]{},
					SplitedList: [][]*Path[PayloadType]{},
				}
			}

			s.Paths.Regex[perfix].List = append(s.Paths.Regex[perfix].List, &paths[i])

			if _, ok := usednames[fullName]; ok {
				log.Fatalf("This URL path already exists: %v", name)
			}
			usednames[fullName] = true

		} else {
			if s.Paths.Static == nil {
				s.Paths.Static = map[string]*Path[PayloadType]{}
			}

			if _, ok := s.Paths.Static[name]; ok {
				log.Fatalf("This URL path already exists: %v", name)
			}

			s.Paths.Static[name] = &paths[i]

			//Trim last '/'
			s.Paths.Static[name[:len(name)-1]] = &paths[i]

		}

	}

}
func (s *Server[PayloadType]) Listen(ctx context.Context, addr, unixSocketPath string) {
	// Regex yollarını işle
	for i := range s.Paths.Regex {
		s.Paths.Regex[i].SplitedList = SplitArray(s.Paths.Regex[i].List, s.RegexOptions.ParallelSearchCount)
		s.Paths.Regex[i].SplitedListLen = len(s.Paths.Regex[i].SplitedList)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.mainHandler)

	errCh := make(chan error, 2)

	if addr != "" {
		s.TCPServer = &http.Server{Addr: addr, Handler: mux}
		go func() {
			go func() {
				<-ctx.Done()
				if err := s.TCPServer.Close(); err != nil {
					log.Printf("Error closing TCP server: %v", err)
				}
			}()
			if err := s.TCPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				errCh <- err
			}
		}()
	}

	if unixSocketPath != "" {
		if _, err := os.Stat(unixSocketPath); err == nil {
			if err := os.Remove(unixSocketPath); err != nil {
				log.Fatalf("Failed to remove existing socket file: %v", err)
			}
		}

		listener, err := net.Listen("unix", unixSocketPath)
		if err != nil {
			log.Fatalf("Error creating Unix socket: %v", err)
		}
		s.unixListener = listener

		if err := os.Chmod(unixSocketPath, 0666); err != nil {
			log.Fatalf("Error setting socket permissions: %v", err)
		}

		s.UnixSocketServer = &http.Server{Handler: mux}
		go func() {
			go func() {
				<-ctx.Done()
				if err := s.UnixSocketServer.Close(); err != nil {
					log.Printf("Error closing Unix socket server: %v", err)
				}
			}()
			if err := s.UnixSocketServer.Serve(listener); err != nil && err != http.ErrServerClosed {
				errCh <- err
			}
		}()
	}

	if addr == "" && unixSocketPath == "" {
		log.Println("No address or Unix socket path provided; exiting")
		return
	}

	select {
	case <-ctx.Done():
		log.Println("Server stopped due to context cancellation")
	case err := <-errCh:
		log.Fatalf("Server error: %v", err)
	}
}

func (s *Server[PayloadType]) mainHandler(w http.ResponseWriter, r *http.Request) {
	var path *Path[PayloadType]
	var params map[string]string = map[string]string{}

	if p, ok := s.Paths.Static[r.URL.Path]; ok {
		path = p
	} else if path == nil {

		for perfix, v := range s.Paths.Regex {

			if p := r.URL.Path; len(p) == 0 || p[len(p)-1] != '/' {
				r.URL.Path += "/"
			}

			if !strings.HasPrefix(r.URL.Path, perfix) {
				continue
			}

			var wg sync.WaitGroup
			ctx, cancel := context.WithCancel(r.Context())
			defer cancel()
			for i := 0; i < v.SplitedListLen; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					for i2 := 0; i2 < len(v.SplitedList[i]); i2++ {
						// Eğer birisi sonucu bulmuşsa, çık
						select {
						case <-ctx.Done():
							return
						default:
							if !v.SplitedList[i][i2].regexName.MatchString(r.URL.Path) {
								continue
							}
							cancel()
							perfixedPath := r.URL.Path[len(perfix):]
							if p := perfixedPath; len(p) == 0 || p[len(p)-1] != '/' {
								perfixedPath += "/"
							}

							url := strings.Split(perfixedPath, "/")
							for i, v := range v.SplitedList[i][i2].mappedParams {
								if url[i] == "" {
									continue
								}
								params[v] = url[i]
							}

							path = v.SplitedList[i][i2]
							return
						}
					}
				}(i)
			}

			wg.Wait()
			break
		}
	}

	request := HTTPRequest{HTTP: r, Params: params}
	response := HTTPResponse{Writer: w}

	if path != nil {
		switch request.IsWebSocketConnection() {
		case true:
			s.WebSocketHandler(&request, &response, path.Payload, path.WebSocket.Upgrader)
			return
		default:
			switch path.IsMethodAllowed(request.Method()) {
			case true:
				s.HTTPHandler(&request, &response, path.Payload)
			default:
				s.HTTPHandle405(&request, &response, path.Payload)
			}
			return
		}
	}
	var zeroValue PayloadType
	s.HTTPHandle404(&request, &response, zeroValue)
}
