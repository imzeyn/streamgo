package streamgo

import (
	"math/rand"
	"net/http"
	"regexp"
	"strings"
)

func (s *HTTPServer) AddPaths(p []*Path) *HTTPServer{
	for _, v := range p {
		s.AddOnePath(v)
	}

	return s
}

func (s *HTTPServer) AddOnePath(p *Path) *HTTPServer{
	if len(p.AllowedMethods) == 0 {
		p.AllowedMethods = map[HTTPMethod]bool{
			GET: true,
		}
	}
	s.makeTemp(p, "")
	return s
}

func (s *HTTPServer) BuildPaths() *HTTPServer{
	s.buildTemps()
	return s
}

func (s *HTTPServer) makeTemp(p *Path, parent string) {
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
			s.makeTemp(v, key)			
		}
	}
}

func (s *HTTPServer) buildTemps(){
	for key, v := range s.paths.temps {
		fullName := s.paths.getTempPathFullName(key)

		if !singleBracePattern.MatchString(fullName){
			s.paths.static[fullName] = v.Path
			continue
		}

		reURL := ""
		paramList := make(map[string]int)

		for i, v := range strings.Split(fullName, "/") {
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
		
		s.paths.regex = append(s.paths.regex, rePath{
			RegexName: *regexp.MustCompile( "^" + ClearURL(reURL) + "$" ),
			FullName: fullName,
			Path: v.Path,
			paramList: paramList,
		})
		
	}

	s.paths.temps = make(map[string]tempPaths)
}

func (s *HTTPServer) Listen(){
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		s.handler(&w, r, false)
	})

	if s.WebSocket != nil {
		s.WebSocket.Perfix = ClearURL(s.WebSocket.Perfix)
		
		mux.HandleFunc(s.WebSocket.Perfix, func(w http.ResponseWriter, r *http.Request) {
			s.handler(&w, r, true)
		})
	}

	http.ListenAndServe(s.Addr, mux)
}

func (s *HTTPServer) handler(w *http.ResponseWriter, r *http.Request, webSocket bool){

	url := ClearURL(r.URL.Path)

	var path *Path
	var params map[string] string
	if webSocket{
		unPerfix := strings.TrimPrefix(url, s.WebSocket.Perfix)
		if !strings.HasPrefix("/", unPerfix){
			unPerfix = "/" + unPerfix
		}
		path, params = s.getPathAndParams(unPerfix)
	}else{
		path, params = s.getPathAndParams(url)
	}

		
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
		Params: &params,
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

	if webSocket{
		conn, err := s.WebSocket.Upgrader.Upgrade(*w, r, nil)
		
		
		ws := WSConnection{
			Conn: conn,
			Err: err,
		}

		path.WSHandler(&request, &ws)
		return
	}
	path.Handler(&request, &response)
}

func (s *HTTPServer) getPathAndParams(u string) (*Path, map[string] string){
	params := make(map[string]string)
	if v, ok := s.paths.static[u]; ok{
		return v, params
	}else{
		
		for _, v := range s.paths.regex {
			if v.RegexName.MatchString(u) {  
				paramURL := strings.Split(u, "/")  
				
				nonEmptyParams := make([]string, 0, len(paramURL))
				for _, part := range paramURL {
					trimmed := strings.TrimSpace(part)
					if trimmed != "" {
						nonEmptyParams = append(nonEmptyParams, trimmed)
					}
				}
		
				paramLen := len(nonEmptyParams)
				
				for k, index := range v.paramList {
					if paramLen > index-1 {
						params[k] = nonEmptyParams[index-1] 
					}
				}
		
				return v.Path, params
			}
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