package streamgo

import (
	"regexp"

	"github.com/gorilla/websocket"
)

// Path represents an endpoint and its associated configurations.
// PayloadType is a generic parameter defining the type of payload associated with the endpoint.
type Path[Payload any] struct {
	// Name specifies the name of the endpoint.
	Name string

	// regexName is a preprocessed regular expression representation of the endpoint name.
	regexName *regexp.Regexp

	// Include contains sub-paths that belong to this endpoint.
	// The URL structure follows the pattern "/endpoint/...".
	Include []Path[Payload]

	// Payload holds additional data associated with the endpoint.
	// This field is optional and can be used for middleware or custom payloads.
	Payload Payload

	// HTTP contains details related to the HTTP connection for this endpoint.
	// It must be configured if an HTTP connection is required.
	HTTP HTTP

	// WebSocket holds the configuration details for a WebSocket connection.
	// This must be set if a WebSocket connection is required.
	WebSocket WS

	// mappedParams is a map of parameter indices to their corresponding names.
	mappedParams map[int]string
}

// HTTP represents the HTTP configuration for an endpoint.
// PayloadType is a generic parameter defining the type of payload associated with the endpoint.
type HTTP struct {
	// Methods is a map that defines the allowed HTTP methods for the endpoint.
	// The key represents the HTTP method, and the value indicates whether it is allowed.
	Methods map[HTTPMethod]bool
}

// WS represents the WebSocket configuration for an endpoint.
// PayloadType is a generic parameter defining the type of payload associated with the endpoint.
type WS struct {
	// Upgrader is responsible for upgrading an HTTP connection to a WebSocket connection.
	Upgrader *websocket.Upgrader
}

// NormalizeMethods ensures that the HTTP.Methods map is initialized.
// If no methods are defined, it defaults to allowing only the GET method.
func (p *Path[Payload]) NormalizeMethods() {
	if p.HTTP.Methods != nil {
		return
	}
	p.HTTP.Methods = map[HTTPMethod]bool{GET: true}
}

// IsMethodAllowed checks whether a given HTTP method is allowed for the endpoint.
// It returns true if the method is allowed, otherwise false.
func (p *Path[Payload]) IsMethodAllowed(method string) bool {
	return p.HTTP.Methods[HTTPMethod(method)]
}
