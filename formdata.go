package gonanoweb

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/textproto"
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
