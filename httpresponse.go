package streamgo

import (
	"io"
	"net/http"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

var (
	json            = jsoniter.ConfigCompatibleWithStandardLibrary
	contentType     = "Content-Type"
	contentTypeHTML = []string{"text/html; charset=utf-8"}
	contentTypeJSON = []string{"application/json; charset=utf-8"}
)

type HTTPResponse struct {
	Writer http.ResponseWriter // Pointer yerine direkt interface'i kullan!
}

func (resp *HTTPResponse) Status(i int) {

	resp.Writer.WriteHeader(i)
}

func (resp *HTTPResponse) HTML(s string) (int, error) {
	h := resp.Writer.Header()
	h[contentType] = contentTypeHTML
	if sw, ok := resp.Writer.(io.StringWriter); ok {
		return sw.WriteString(s)
	}
	return resp.Writer.Write([]byte(s))
}

func (resp *HTTPResponse) Write(v []byte) (int, error) {
	return resp.Writer.Write(v)
}

func (resp *HTTPResponse) JSON(v any) (int, error) {
	h := resp.Writer.Header()
	h[contentType] = contentTypeJSON
	b, err := json.Marshal(v)
	if err != nil {
		return 0, err
	}
	return resp.Write(b)
}

func (resp *HTTPResponse) Headers(vals map[string]string) {
	headers := resp.Writer.Header() // get headers map once
	for k, v := range vals {
		headers.Add(k, v)
	}
}

func (resp *HTTPResponse) Cookie(name, value, maxage, samesite, path, domain string, http, secure bool) {
	var builder strings.Builder

	builder.WriteString(name)
	builder.WriteString("=")
	builder.WriteString(value)
	builder.WriteString("; Max-Age=")
	builder.WriteString(maxage)
	builder.WriteString("; Path=")
	builder.WriteString(path)

	if domain != "" {
		builder.WriteString("; Domain=")
		builder.WriteString(domain)
	}

	if http {
		builder.WriteString("; HttpOnly")
	}

	if secure {
		builder.WriteString("; Secure")
	}

	if samesite != "" {
		builder.WriteString("; SameSite=")
		builder.WriteString(samesite)
	}

	// Set the header with the built cookie string
	resp.Headers(map[string]string{"Set-Cookie": builder.String()})
}

func (resp *HTTPResponse) CookieWithDefaults(name, value, maxage string, secure bool) {
	resp.Cookie(name, value, maxage, "Lax", "/", "", true, secure)
}
