package gonanoweb

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
)

type ResponseHeaders struct {
	Values []string
}

func (h *ResponseHeaders) Add(key, value string) {
	h.Values = append(h.Values, fmt.Sprintf("%s: %s", key, value))
}

type Response struct {
	Server      *Server
	conn        net.Conn
	Status      int
	Body        []byte
	Headers     ResponseHeaders
	EventStream *EventStream
}

func (r *Response) ApiError(code int, message string) {
	r.Status = code
	r.Headers.Add("content-type", "application/json")
	body := map[string]string{
		"message": message,
	}

	data, _ := json.Marshal(body)
	r.Body = []byte(data)
}

func (r *Response) ApiErrorWithErr(code int, message string, err error) {
	r.Status = code
	r.Headers.Add("content-type", "application/json")
	body := map[string]string{
		"message": message,
	}

	data, _ := json.Marshal(body)
	r.Body = []byte(data)
	
	// Log the full error on the server if needed
	if err != nil {
		// You could add proper logging here
		// log.Printf("Error: %v", err)
	}
}

func (r *Response) Json(status int, body interface{}) {
	r.Status = status
	r.Headers.Add("content-type", "application/json")

	json, err := json.Marshal(body)
	if err != nil {
		r.ApiError(500, "Failed to marshal response body.")
		return
	}
	r.Body = json
}

func (r *Response) TextPlain(status int, body string) {
	r.Status = status
	r.Headers.Add("content-type", "text/plain")
	r.Body = []byte(body)
}

func (r *Response) Raw(status int, body []byte) {
	r.Status = status
	r.Headers.Add("content-type", "application/octet-stream")
	r.Body = body
}

var statusTextMap = map[int]string{
	100: "Continue",
	101: "Switching Protocols",
	102: "Processing",
	103: "Early Hints",
	200: "OK",
	201: "Created",
	202: "Accepted",
	203: "Non Authoritative Information",
	204: "No Content",
	205: "Reset Content",
	206: "Partial Content",
	207: "Multi-Status",
	300: "Multiple Choices",
	301: "Moved Permanently",
	302: "Moved Temporarily",
	303: "See Other",
	304: "Not Modified",
	305: "Use Proxy",
	307: "Temporary Redirect",
	308: "Permanent Redirect",
	400: "Bad Request",
	401: "Unauthorized",
	402: "Payment Required",
	403: "Forbidden",
	404: "Not Found",
	405: "Method Not Allowed",
	406: "Not Acceptable",
	407: "Proxy Authentication Required",
	408: "Request Timeout",
	409: "Conflict",
	410: "Gone",
	411: "Length Required",
	412: "Precondition Failed",
	413: "Request Entity Too Large",
	414: "Request-URI Too Long",
	415: "Unsupported Media Type",
	416: "Requested Range Not Satisfiable",
	417: "Expectation Failed",
	418: "I'm a teapot",
	419: "Insufficient Space on Resource",
	420: "Method Failure",
	421: "Misdirected Request",
	422: "Unprocessable Entity",
	423: "Locked",
	424: "Failed Dependency",
	426: "Upgrade Required",
	428: "Precondition Required",
	429: "Too Many Requests",
	431: "Request Header Fields Too Large",
	451: "Unavailable For Legal Reasons",
	500: "Internal Server Error",
	501: "Not Implemented",
	502: "Bad Gateway",
	503: "Service Unavailable",
	504: "Gateway Timeout",
	505: "HTTP Version Not Supported",
	507: "Insufficient Storage",
	511: "Network Authentication Required",
}

func statusText(code int) string {
	if text, exists := statusTextMap[code]; exists {
		return text
	}
	return "Unknown Status"
}

func (r *Response) Done() {
	r.handleSecurityHeaders()

	response := fmt.Sprintf("HTTP/1.1 %d %s\r\n%s\r\n", r.Status, statusText(r.Status), strings.Join(r.Headers.Values, "\r\n"))
	r.conn.Write([]byte(response))

	if r.Body != nil {
		r.conn.Write([]byte(fmt.Sprintf("Content-Length: %d\r\n\r\n", len(r.Body))))
		r.conn.Write(r.Body)
	} else {
		r.conn.Write([]byte("\r\n\r\n"))
	}

	r.conn.Close()
}

type EventStream struct {
	Identifier string
	Ch         *chan string
}

func (r *Response) StreamEvents() {
	head := fmt.Sprintf(
		"HTTP/1.1 %d %s\r\n%s\r\n",
		r.Status,
		statusText(r.Status),
		strings.Join(r.Headers.Values, "\r\n"),
	)
	r.conn.Write([]byte(head))
	r.conn.Write([]byte("Content-Type: text/event-stream\n\n"))

	defer r.conn.Close()
	for {
		select {
		case msg, ok := <-*r.EventStream.Ch:
			if !ok {
				return
			}

			_, err := r.conn.Write([]byte(fmt.Sprintf("data: %s\n\n", msg)))
			if err != nil {
				return
			}
		}
	}

}
func (r *Response) handleSecurityHeaders() {
	if *&r.Server.SecurityHeaders != nil && *r.Server.SecurityHeaders {
		r.Headers.Add("X-Content-Type-Options", "nosniff")
		r.Headers.Add("X-Frame-Options", "DENY")
		r.Headers.Add("X-XSS-Protection", "1; mode=block")
	}
}
