package streamgo

import (
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func (s *HTTPServer) AddPaths(p []Path) *HTTPServer{
	for _, v := range p {
		s.AddOnePath(v)
	}
	return s
}

func (s *HTTPServer) AddOnePath(p Path) *HTTPServer{
	if len(p.AllowedMethods) == 0 {
		p.AllowedMethods = map[HTTPMethod]bool{
			GET: true,
		}
	}

	if p.WebSocket.Handler != nil{
		if p.WebSocket.Upgrader.Error == nil{
			p.WebSocket.Upgrader.Error = onWSUpgradeErr
		}
	}
	
	s.makeTemp(p, "")
	return s
}

func (s *HTTPServer) BuildPaths() *HTTPServer{
	s.buildTemps()
	return s
}

func (s *HTTPServer) makeTemp(p Path, parent string) {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	
	id := make([]byte, 16)
	for i := range id {
		id[i] = charset[rand.Intn(len(charset))]
	}

	key := string(id)

	s.paths.temps[key] = tempPaths{
		Parent: parent,
		Path: p,
	}

	if p.Include != nil{
		for _, v := range p.Include {
			
			if len(v.AllowedMethods) == 0 {
				v.AllowedMethods = map[HTTPMethod]bool{
					GET: true,
				}
			}
			
			s.makeTemp(v, key)
		}
	}
}

func (s *HTTPServer) buildTemps(){
	re := regexp.MustCompile(`^(.*?){[^/]+}/`)

	for key, v := range s.paths.temps {
		fullName := s.paths.getTempPathFullName(key)

		if !singleBracePattern.MatchString(fullName){
			s.paths.static[fullName] = v.Path
			continue
		}

		reURL := ""
		paramList := map[string]int{}

		matches := re.FindStringSubmatch(fullName)
		matchParam := strings.TrimPrefix(fullName, matches[1])
		
		for i, v := range strings.Split(matchParam, "/") {
			paramArr := singleBracePattern.FindStringSubmatch(v)
			if len(paramArr) > 0{
				paramName := paramArr[0]
				paramName = strings.ReplaceAll(paramName, "{", "")
				paramName = strings.ReplaceAll(paramName, "}", "")
				if doubleBracePattern.MatchString(v){
					reURL += doubleBracePattern.ReplaceAllString(v, `?([\p{L}\p{N}\p{M}.@_-]*)?/`)
					paramList[paramName] = i			
				}else if singleBracePattern.MatchString(v){
					reURL += singleBracePattern.ReplaceAllString(v, `[\p{L}\p{N}\p{M}.@_-]+/`)
					paramList[paramName] = i
				}else{
					reURL += v + "/"
				}
			}else{
				reURL += v + "/"
			}
		}

		rePathData := rePath{
			RegexName: *regexp.MustCompile( "^" + ClearURL(reURL) + "$" ),
			FullName: fullName,
			Path: v.Path,
			paramList: paramList,
		}

		s.paths.regex[matches[1]] = append(s.paths.regex[matches[1]], rePathData)
	}

	s.paths.temps = make(map[string]tempPaths)
}

func (s *HTTPServer) Listen(proxyAddress, unixSocketPath string){
	if unixSocketPath != "" {
		if _, err := os.Stat(unixSocketPath); err == nil {
			err := os.Remove(unixSocketPath)
			if err != nil {
				log.Fatalf("Failed to remove existing socket file: %v", err)
			}
		}
	
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			s.handler(&w, r)
		})
		
		listener, err := net.Listen("unix", unixSocketPath)
		if err != nil {
			log.Fatalf("Error creating Unix socket: %v", err)
		}

		defer listener.Close()
		err = os.Chmod(unixSocketPath, 0666)
		if err != nil {
			log.Fatal("Error setting socket permissions:", err)
		}
		
		go func ()  {
			if err := http.Serve(listener, mux); err != nil {
				log.Fatalf("Failed to start HTTP server on Unix socket: %v", err)
			}
		}()
	}
	
	if proxyAddress != ""{
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			s.handler(&w, r)
		})
		
		go http.ListenAndServe(proxyAddress, mux)	
	}

	select{}
}

func (s *HTTPServer) handler(w *http.ResponseWriter, r *http.Request){
	url := ClearURL(r.URL.Path)

	var path *Path
	
	isws := r.Header.Get("Upgrade") == "websocket"

	path, params := s.getPathAndParams(url, isws)

	pathDetailsData := &pathDetails{}
	if path != nil{
		pathDetailsData.ID = path.ID
		pathDetailsData.Name = path.Name
		pathDetailsData.Additional = path.Additional
	}else{
		pathDetailsData = nil
	}
	
	request := HTTPRequest{
		HTTP: r,
		Params: params,
		FullName: url,
		PathDetails: pathDetailsData,
	}

	response := HTTPResponse{
		writer: w,
	}

	if s.HandleMiddleware(&request, &response){
		return
	}
	
	if path == nil{

		s.on404(&request, &response)
		return
	}else if !path.MethodAllowed(HTTPMethod(r.Method)){
		s.on403(&request, &response)
		return
	}

	if isws{
		conn, err := path.WebSocket.Upgrader.Upgrade(*w, r, nil)
		ws := WSConnection{
			Conn: conn,
			Err: err,
		}
		path.WebSocket.Handler(&request, &ws)
		return
	}
	path.Handler(&request, &response)
}

func (s *HTTPServer) getPathAndParams(u string, isws bool) (*Path, map[string] string){
	params := map[string]string{}

	if v, ok := s.paths.static[u]; ok{
		 
		if v.Handler == nil && !isws{
			return nil, params
		}

		if v.WebSocket.Handler == nil && isws{
			return nil, params
		}

		return &v, params
	}


	for k, vP := range s.paths.regex {	
		if !strings.HasPrefix(u, k){
			continue
		}
		
		paramStr :=  "/" + strings.TrimPrefix(u, k)

		for _, v := range vP {
			
			if !v.RegexName.MatchString(paramStr) {  
				continue
			} 	
			
			parts 		  := strings.Split(paramStr, "/")
			nonEmptyParts := parts[1:len(parts)-1]
			lenParts 	  := len(nonEmptyParts)

			if lenParts != 0{
				for id, index := range v.paramList {
					if index <= (lenParts - 1) {
						params[id] = nonEmptyParts[index]
					}
				}	
			}
			
			if v.Path.Handler == nil && !isws{
				return nil, params
			}
	
			if v.Path.WebSocket.Handler == nil && isws{
				return nil, params
			}

			return &v.Path, params
		}
	}

	return nil, params
}

func (s *HTTPServer) HandleMiddleware(request *HTTPRequest, response *HTTPResponse) bool{
	for _, v := range s.Middleware {
		if !v(request, response){
			return true
		}
	}
	return false
}

func (s *HTTPServer) on404(request *HTTPRequest, response *HTTPResponse){
	response.Status(404)
	if s.On404 != nil{
		s.On404(request, response)
		return
	}
	response.String("")
}

func (s *HTTPServer) on403(request *HTTPRequest, response *HTTPResponse){
	response.Status(403)
	if s.On403 != nil{
		s.On403(request, response)
		return
	}
	response.String("")
}

func onWSUpgradeErr(w http.ResponseWriter, r *http.Request, status int, reason error) {
	w.WriteHeader(400)
}