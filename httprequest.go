package streamgo

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var (
    browserMap map[string]string = map[string]string{
        "Firefox/":  "Mozilla Firefox",
        "Chrome/":   "Google Chrome",
        "Safari/":   "Apple Safari",
        "Edge/":     "Microsoft Edge",
        "MSIE":      "Microsoft Internet Explorer",
        "Trident/":  "Microsoft Internet Explorer",
    }

    osMap map[string]string = map[string]string{
        "Windows NT": "Windows",
        "Mac OS X":   "Mac OS X",
        "Linux":      "Linux",
        "Android":    "Android",
        "iPhone":     "iOS",
        "iPad":       "iOS",
    }
)


func (r *HTTPRequest) Cookies() map[string]*http.Cookie {
	cookies := make(map[string]*http.Cookie)
	for _, cookie := range r.HTTP.Cookies() {
		cookies[cookie.Name] = cookie
	}
	return cookies
}

func (r *HTTPRequest) Cookie(name string) (*http.Cookie, bool) {
    cookies := r.Cookies()
    cookie, exists := cookies[name]
    return cookie, exists
}

func (r *HTTPRequest) Headers() *http.Header{
	return &r.HTTP.Header
}

func (r *HTTPRequest) Header(name string) string{
	return r.HTTP.Header.Get(name)
}

func (r *HTTPRequest) IP() string {
    forwarded := r.HTTP.Header.Get("X-Forwarded-For")
    if forwarded != "" {
        ips := strings.Split(forwarded, ",")
        ip := strings.TrimSpace(ips[0])
        if net.ParseIP(ip) != nil {
            return ip 
        }
    }
   
    realIP := r.HTTP.Header.Get("X-Real-IP")
    if realIP != "" {
        if net.ParseIP(realIP) != nil {
            return realIP 
        }
    }
   
    ip, _, err := net.SplitHostPort(r.HTTP.RemoteAddr)
    if err != nil {
        return ""
    }
    
    return ip
}

func (r *HTTPRequest) Method() string{
    return r.HTTP.Method
}

func (r *HTTPRequest) Query(name string) string{
    return r.HTTP.URL.Query().Get(name)
}

func (r *HTTPRequest) Querys() url.Values{
    return r.HTTP.URL.Query()
}

func (r *HTTPRequest) Device() (string, string) {
	userAgent := r.HTTP.Header.Get("User-Agent")
	
    browser   := "Unknown Browser"
	for key, value := range browserMap {
		if strings.Contains(userAgent, key) {
			if key == "Safari/" && strings.Contains(userAgent, "Chrome/") {
				continue
			}
			browser = value
			break
		}
	}

	os := "Unknown OS"
	for key, value := range osMap {
		if strings.Contains(userAgent, key) {
			os = value
			break
		}
	}

	return browser, os
}

 

func (r *HTTPRequest) JSON(maxBodySize int64) (*interface{}, error) {
	defer r.HTTP.Body.Close()
	limitedBody := io.LimitReader(r.HTTP.Body, maxBodySize)
	var data interface{}
	
    decoder := json.NewDecoder(limitedBody)
	err := decoder.Decode(&data)
	
    if err != nil {
		if err == io.EOF {
			return &data, nil
		}
		return nil, err
	}
	return &data, nil
}

func (r *HTTPRequest) Upload(to, name string) (bool, error) {
	mr, err := r.HTTP.MultipartReader()
	if err != nil {
		return false, err
	}

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break 
		}
		if err != nil {
			return false, err
		}

		if part.FormName() == name {
			dst, err := os.Create(to)
			if err != nil {
				return false, err
			}
			defer dst.Close()

			_, err = io.Copy(dst, part)
			if err != nil {
				return false, err
			}
		}
	}

	return true, nil
}

func (r *HTTPRequest) UploadIfValid(name, to string, signature *MimeSignatureList) (bool, string, error) {
    mr, err := r.HTTP.MultipartReader()
    if err != nil {
        return false, "", err
    }

    for {
        part, err := mr.NextPart()
        if err == io.EOF {
            break
        }
        if err != nil {
            return false, "", err
        }

        if part.FormName() == name {
            file_name := strings.Split(part.FileName(), ".")
            file_ext := strings.ToLower(file_name[len(file_name)-1])
            file_name = nil
            header := make([]byte, 512)
            n, err := part.Read(header)
            if err != nil && err != io.EOF {
                return false, "", err
            }

            for _, v := range *signature {
                if bytes.HasPrefix(header[:n], v.Signature) {
                    for _, ext := range v.Extensions {
                        if file_ext == ext {
                            dst, err := os.Create(to + "." + ext)
                            if err != nil {
                                return false, "", err
                            }
                            defer dst.Close()

                            if _, err := dst.Write(header[:n]); err != nil {
                                return false, "", err
                            }

                            if _, err := io.Copy(dst, part); err != nil {
                                return false, "", err
                            }

                            return true, ext, nil
                        }
                    }
                    return false, "", ErrUploadExtMismatch
                }
            }

            return false, "", ErrUploadSigMismatch
        }
    }

    return false, "", ErrUploadFileMissing
}