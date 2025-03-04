package gonanoweb

import (
	"bytes"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/textproto"
	"strings"
)

type FormFile struct {
	Filename string
	Header   textproto.MIMEHeader
	Size     int64
	Reader   io.Reader
}

type FormData struct {
	Fields map[string][]string
	Files  map[string][]*FormFile
}

type FormDataOptions struct {
	MaxMemory       int64  // Maximum size in memory before using temp files
	MaxFileSize     int64  // Maximum size of individual files
	TempDir         string // Temporary files Directory
	StreamingParser bool   // Whether to enable streaming parser
}

func DefaultFormDataOptions() FormDataOptions {
	return FormDataOptions{
		MaxMemory:       10 << 20, // 10 MB
		MaxFileSize:     32 << 20, // 32 MB
		TempDir:         "",
		StreamingParser: false,
	}
}

func FormDataMiddleware(options *FormDataOptions) Middleware {
	if options == nil {
		defaultOptions := DefaultFormDataOptions()
		options = &defaultOptions
	}

	return Middleware{
		Handler: func(res *Response, req *Request) error {
			contentType := req.Headers["content-type"]
			if contentType == "" {
				return nil
			}

			mediaType, params, err := mime.ParseMediaType(contentType)
			if err != nil || !strings.HasPrefix(mediaType, "multipart/form-data") {
				return nil
			}

			boundary, ok := params["boundary"]
			if !ok {
				return errors.New("invalid multipart boundary")
			}

			if req.Body == nil {
				req.Body = &[]byte{}
				return errors.New("request body is nil")
			}

			bodyReader := bytes.NewReader(*req.Body)
			reader := multipart.NewReader(bodyReader, boundary)

			if options.StreamingParser {
				formData := &FormData{
					Fields: make(map[string][]string),
					Files:  make(map[string][]*FormFile),
				}

				// Store the multipart reader directly in the Request struct
				req.MultipartReader = reader

				err = req.SetData("formData", formData)
				if err != nil {
					return err
				}

				return nil
			} else {
				formData, err := parseEntireForm(reader, options)
				if err != nil {
					return err
				}

				err = req.SetData("formData", formData)
				if err != nil {
					return err
				}

				return nil
			}
		},
	}
}

func parseEntireForm(reader *multipart.Reader, options *FormDataOptions) (*FormData, error) {
	formData := &FormData{
		Fields: make(map[string][]string),
		Files:  make(map[string][]*FormFile),
	}

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		name := part.FormName()
		if name == "" {
			continue
		}

		filename := part.FileName()
		if filename != "" {
			// This is a file upload
			var buffer bytes.Buffer
			size, err := io.CopyN(&buffer, part, options.MaxFileSize+1)
			if err != nil && err != io.EOF {
				return nil, err
			}
			if size > options.MaxFileSize {
				return nil, errors.New("file exceeds maximum allowed size")
			}

			formFile := &FormFile{
				Filename: filename,
				Header:   part.Header,
				Size:     size,
				Reader:   bytes.NewReader(buffer.Bytes()),
			}
			formData.Files[name] = append(formData.Files[name], formFile)
		} else {
			value, err := io.ReadAll(part)
			if err != nil {
				return nil, err
			}
			formData.Fields[name] = append(formData.Fields[name], string(value))
		}
	}

	return formData, nil
}

func GetFormData(req *Request) (*FormData, error) {
	var formDataInterface interface{}
	err := req.GetData("formData", &formDataInterface)
	if err != nil {
		return nil, errors.New("form data not found in request")
	}

	formData, ok := formDataInterface.(*FormData)
	if !ok {
		return nil, errors.New("invalid form data type")
	}

	return formData, nil
}

func GetMultipartReader(req *Request) (*multipart.Reader, error) {
	// Now directly return the MultipartReader from the Request struct
	if req.MultipartReader == nil {
		return nil, errors.New("multipart reader not found in request")
	}

	return req.MultipartReader, nil
}
