package streamgo

import (
	"encoding/json"
	"fmt"
)

func (resp *HTTPResponse) String(s string) (int, error){
	wr := *resp.writer
	wr.Header().Add("Content-Type", "text/html; charset=utf-8;")
	return wr.Write([]byte(s))
}

func (resp *HTTPResponse) JSON(w interface{}) error {
	data, e := json.Marshal(w)
	if e != nil {
		return e
	}

	wr := *resp.writer
	
	if wr.Header().Get("Content-Type") == "" {
		wr.Header().Add("Content-Type", "application/json; charset=utf-8")
	}
	
	wr.Write(data)

	return nil
}

func (resp *HTTPResponse) Headers(vals map[string]string) *HTTPResponse{
	wr := *resp.writer
	
	for k, v := range vals {
		wr.Header().Add(k, v)		
	}

	return resp
}

func (resp *HTTPResponse) Cookie(name, value, maxage, samesite, path, domain string, http, secure bool) *HTTPResponse{
    cookieStr := fmt.Sprintf("%s=%s; Max-Age=%s; Path=%s", name, value, maxage, path) // value burada eklendi

    if domain != "" {
        cookieStr += fmt.Sprintf("; Domain=%s", domain)
    }

    if http {
        cookieStr += "; HttpOnly"
    }

    if secure {
        cookieStr += "; Secure"
    }

	
    if samesite != "" {
        cookieStr += fmt.Sprintf("; SameSite=%s", samesite)
    }

	resp.Headers(map[string]string{"Set-Cookie": cookieStr})
	
	return resp
}

func (resp *HTTPResponse) CookieWithDefaults(name, value, maxage string, secure bool) *HTTPResponse { 
	return resp.Cookie(name, value, maxage, "Lax", "/", "", true, secure)
}

func (resp *HTTPResponse) Status(code int) *HTTPResponse{
	wr := *resp.writer
	wr.WriteHeader(code)
	return resp	
}