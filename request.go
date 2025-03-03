package gonanoweb

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"strings"
)

type Request struct {
	Method          string
	Path            string
	Headers         map[string]string
	QueryParams     map[string][]string
	Params          map[string]string
	Body            *[]byte
	MaxRequestSize  *int64
	data            map[string]interface{}
	reader          *bufio.Reader
	conn            *net.Conn
	MultipartReader *multipart.Reader
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
		var contentLength int64
		_, err := fmt.Sscanf(length, "%d", &contentLength)
		if err != nil {
			return fmt.Errorf("invalid content-length: %w", err)
		}

		// prevents unreasonable body sizes
		if r.MaxRequestSize != nil && contentLength > *r.MaxRequestSize {
			return fmt.Errorf("content length %d exceeds maximum allowed size %d", contentLength, *r.MaxRequestSize)
		}

		// default size limit if MaxRequestSize not set
		if r.MaxRequestSize == nil && contentLength > 10*1024*1024 {
			return fmt.Errorf("content length %d exceeds default maximum size", contentLength)
		}

		// safe size for int conversion for allocation
		if contentLength > 2147483647 {
			return fmt.Errorf("content length too large")
		}

		// prevent empty allocation
		if contentLength <= 0 {
			body = []byte{}
			r.Body = &body
			return nil
		}

		body = make([]byte, contentLength)
		n, err := io.ReadFull(r.reader, body)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return err
		}

		// trim to actual size if we read less than expected
		if int64(n) < contentLength {
			body = body[:n]
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
