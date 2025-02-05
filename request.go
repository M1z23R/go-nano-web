package gonanoweb

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
)

type Request struct {
	Method         string
	Path           string
	Headers        map[string]string
	QueryParams    map[string][]string
	Params         map[string]string
	Body           *[]byte
	MaxRequestSize *int64
	data           map[string]interface{}
	reader         *bufio.Reader
	conn           *net.Conn
}

func NewRequest() *Request {
	return &Request{
		data: make(map[string]interface{}),
	}
}

func (r *Request) parseRequest(conn net.Conn) error {
	if r.MaxRequestSize != nil && *r.MaxRequestSize > 0 {
		r.reader = bufio.NewReaderSize(io.LimitReader(conn, *r.MaxRequestSize), 4096)
	} else {
		r.reader = bufio.NewReader(conn)
	}

	line, err := r.reader.ReadString('\n')
	if err != nil {
		return nil
	}
	line = strings.TrimSpace(line)

	parts := strings.Split(line, " ")
	if len(parts) < 3 {
		return fmt.Errorf("invalid request line")
	}
	method := parts[0]
	fullPath := parts[1]
	path, queryString := splitPathAndQuery(fullPath)
	queryParams := parseQueryParams(queryString)
	headers := make(map[string]string)
	for {
		line, err := r.reader.ReadString('\n')
		if err != nil {
			return err
		}
		line = strings.TrimSpace(line)

		if line == "" {
			break
		}

		headerParts := strings.SplitN(line, ": ", 2)
		if len(headerParts) == 2 {
			headers[strings.ToLower(headerParts[0])] = headerParts[1]
		}
	}

	r.Method = method
	r.Path = path
	r.QueryParams = queryParams
	r.Headers = headers
	return nil
}

func splitPathAndQuery(fullPath string) (string, string) {
	if i := strings.Index(fullPath, "?"); i != -1 {
		return fullPath[:i], fullPath[i+1:]
	}
	return fullPath, ""
}

func parseQueryParams(queryString string) map[string][]string {
	qp := make(map[string][]string)

	if queryString == "" {
		return qp
	}

	pairs := strings.Split(queryString, "&")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		key := kv[0]
		var value string
		if len(kv) > 1 {
			value = kv[1]
		}

		qp[key] = append(qp[key], value)
	}

	return qp
}

func (r *Request) parseBody() error {
	var body []byte
	if length, ok := r.Headers["content-length"]; ok {
		var contentLength int
		fmt.Sscanf(length, "%d", &contentLength)

		body = make([]byte, contentLength)
		_, err := r.reader.Read(body)
		if err != nil {
			return err
		}
	}
	r.Body = &body

	return nil
}

func (r *Request) SetData(key string, data interface{}) error {
	_, ok := r.data[key]
	if ok {
		return errors.New("Key already in use.")
	}

	r.data[key] = data
	return nil
}

func (r *Request) GetData(key string, data *interface{}) error {
	v, ok := r.data[key]
	if !ok {
		return errors.New("No data found for given key.")
	}
	*data = v
	return nil
}
