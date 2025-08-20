package streamgo

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

type HTTPRequest struct {
	HTTP   *http.Request
	Params map[string]string
}

var (
	BrowserMap map[string]string = map[string]string{
		"Firefox/": "Mozilla Firefox",
		"Chrome/":  "Google Chrome",
		"Safari/":  "Apple Safari",
		"Edge/":    "Microsoft Edge",
		"MSIE":     "Microsoft Internet Explorer",
		"Trident/": "Microsoft Internet Explorer",
	}

	OsMap map[string]string = map[string]string{
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

func (r *HTTPRequest) Headers() http.Header {
	return r.HTTP.Header
}

func (r *HTTPRequest) Header(name string) string {
	return r.HTTP.Header.Get(name)
}

// IP returns the real client IP considering X-Forwarded-For, X-Real-IP, and RemoteAddr.
// trustedProxies is a list of IPs or CIDRs for proxies we trust.
func (r *HTTPRequest) IP(trustedProxies []string) string {
    // Parse trusted proxies into net.IPNet
    var trustedNets []*net.IPNet
    for _, cidr := range trustedProxies {
        if _, network, err := net.ParseCIDR(cidr); err == nil {
            trustedNets = append(trustedNets, network)
        } else if ip := net.ParseIP(cidr); ip != nil {
            trustedNets = append(trustedNets, &net.IPNet{
                IP:   ip,
                Mask: net.CIDRMask(32, 32),
            })
        }
    }

    // Helper: check if IP is in trusted proxies
    isTrusted := func(ip net.IP) bool {
        for _, net := range trustedNets {
            if net.Contains(ip) {
                return true
            }
        }
        return false
    }

    // 1. X-Forwarded-For
    forwarded := r.HTTP.Header.Get("X-Forwarded-For")
    if forwarded != "" {
        ips := strings.Split(forwarded, ",")
        // Traverse from right to left: last IP is closest proxy
        for i := len(ips) - 1; i >= 0; i-- {
            ip := strings.TrimSpace(ips[i])
            parsedIP := net.ParseIP(ip)
            if parsedIP == nil {
                continue
            }
            if !isTrusted(parsedIP) {
                return ip // First non-trusted IP is real client
            }
        }
    }

    // 2. X-Real-IP
    realIP := r.HTTP.Header.Get("X-Real-IP")
    if realIP != "" {
        if parsedIP := net.ParseIP(realIP); parsedIP != nil && !isTrusted(parsedIP) {
            return realIP
        }
    }

    // 3. Fallback: RemoteAddr
    ip, _, err := net.SplitHostPort(r.HTTP.RemoteAddr)
    if err != nil {
        return r.HTTP.RemoteAddr
    }
    return ip
}

func (r *HTTPRequest) Method() string {
	return r.HTTP.Method
}

func (r *HTTPRequest) Query(name string) string {
	return r.HTTP.URL.Query().Get(name)
}

func (r *HTTPRequest) Querys() url.Values {
	return r.HTTP.URL.Query()
}

func (r *HTTPRequest) Device() (string, string) {
	userAgent := r.HTTP.Header.Get("User-Agent")

	browser := "Unknown Browser"
	for key, value := range BrowserMap {
		if strings.Contains(userAgent, key) {
			if key == "Safari/" && strings.Contains(userAgent, "Chrome/") {
				continue
			}
			browser = value
			break
		}
	}

	os := "Unknown OS"
	for key, value := range OsMap {
		if strings.Contains(userAgent, key) {
			os = value
			break
		}
	}

	return browser, os
}

func (r *HTTPRequest) JSON(maxBodySize int64, result any) error {
	defer r.HTTP.Body.Close()
	limitedBody := io.LimitReader(r.HTTP.Body, maxBodySize)

	decoder := json.NewDecoder(limitedBody)
	err := decoder.Decode(result)

	switch err {
	case nil:
		return nil
	case io.EOF:
		return nil
	default:
		return err
	}
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

var (
	ErrUploadExtMismatch = errors.New("invalid file extension")
	ErrUploadSigMismatch = errors.New("invalid signature")
	ErrUploadFileMissing = errors.New("file not found")
)

var bufPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 32<<10)
		return &b
	},
}

func (r *HTTPRequest) UploadIfValid(name, to string, signatures *MimeSignatureList) (string, bool, error) {
	mr, err := r.HTTP.MultipartReader()
	if err != nil {
		return "", false, err
	}

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", false, err
		}

		if part.FormName() != name {
			continue
		}

		return r.UploadIfValidFromPart(part, to, signatures)
	}

	return "", false, ErrUploadFileMissing
}

func (r *HTTPRequest) UploadIfValidFromPart(part *multipart.Part, to string, signatures *MimeSignatureList) (string, bool, error) {
	var (
		headerBuf [512]byte
		filename  string
		fileExt   string
	)
	filename = part.FileName()
	if extPos := strings.LastIndexByte(filename, '.'); extPos > 0 {
		fileExt = strings.ToLower(filename[extPos+1:])
	}

	// Read header with zero-alloc
	n, _ := io.ReadFull(part, headerBuf[:])
	if n == 0 {
		return "", false, io.ErrUnexpectedEOF
	}

	// Main detection logic
	for i := range *signatures {
		sig := &(*signatures)[i]

		// Fast path: length check first
		if len(sig.Signature) > n {
			continue
		}

		// Signature match check
		if !bytes.Equal(headerBuf[:len(sig.Signature)], []byte(sig.Signature)) {
			continue
		}

		// Extension check with pre-split optimization
		ext := fileExt
		if len(ext) == 0 || !strings.Contains(sig.Extensions+",", ext+",") {
			// Fallback to first extension if invalid
			if firstExt, _, _ := strings.Cut(sig.Extensions, ","); firstExt != "" {
				ext = firstExt
			} else {
				return "", false, ErrUploadExtMismatch
			}
		}

		// Write file with buffer pooling
		dst, err := os.Create(to + "." + ext)
		if err != nil {
			return "", false, err
		}
		defer dst.Close()

		if _, err = dst.Write(headerBuf[:n]); err != nil {
			return "", false, err
		}

		bufPtr := bufPool.Get().(*[]byte)
		defer bufPool.Put(bufPtr)
		buf := *bufPtr

		if _, err = io.CopyBuffer(dst, part, buf); err != nil {
			return "", false, err
		}

		return ext, true, nil
	}

	return "", false, ErrUploadSigMismatch
}

func (r *HTTPRequest) IsWebSocketConnection() bool {
	return r.Header("Upgrade") == "websocket"
}
