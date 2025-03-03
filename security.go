package gonanoweb

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
)

type SecurityOptions struct {
	EnableCSRF      bool
	CSRFCookieName  string
	CSRFHeaderName  string
	CSRFTokenLength int
	CSRFCookieOpts  *CookieOptions

	ContentSecurityPolicy string

	EnableHSTS             bool
	HSTSMaxAge             int
	HSTSIncludeSubdomains  bool
	HSTSPreload            bool
	EnableNoSniff          bool
	EnableFrameOptions     bool
	FrameOptionsPolicy     string
	EnableXSSProtection    bool
	ReferrerPolicy         string
	PermissionsPolicy      string
}

type CookieOptions struct {
	Path     string
	Domain   string
	MaxAge   int
	Secure   bool
	HttpOnly bool
	SameSite http.SameSite
}

func DefaultSecurityOptions() SecurityOptions {
	return SecurityOptions{
		EnableCSRF:      true,
		CSRFCookieName:  "_csrf",
		CSRFHeaderName:  "X-CSRF-Token",
		CSRFTokenLength: 32,
		CSRFCookieOpts: &CookieOptions{
			Path:     "/",
			MaxAge:   86400, // 24 hours
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		},
		ContentSecurityPolicy: "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline';",
		EnableHSTS:           true,
		HSTSMaxAge:           31536000, // 1 year
		HSTSIncludeSubdomains: true,
		EnableNoSniff:        true,
		EnableFrameOptions:   true,
		FrameOptionsPolicy:   "SAMEORIGIN",
		EnableXSSProtection:  true,
		ReferrerPolicy:       "strict-origin-when-cross-origin",
	}
}

func SecurityMiddleware(options *SecurityOptions) Middleware {
	if options == nil {
		defaultOptions := DefaultSecurityOptions()
		options = &defaultOptions
	}

	return Middleware{
		Handler: func(res *Response, req *Request) error {
			// Add security headers
			if options.EnableNoSniff {
				res.Headers.Add("X-Content-Type-Options", "nosniff")
			}

			if options.EnableFrameOptions {
				policy := options.FrameOptionsPolicy
				if policy == "" {
					policy = "SAMEORIGIN"
				}
				res.Headers.Add("X-Frame-Options", policy)
			}

			if options.EnableXSSProtection {
				res.Headers.Add("X-XSS-Protection", "1; mode=block")
			}

			if options.ContentSecurityPolicy != "" {
				res.Headers.Add("Content-Security-Policy", options.ContentSecurityPolicy)
			}

			if options.ReferrerPolicy != "" {
				res.Headers.Add("Referrer-Policy", options.ReferrerPolicy)
			}

			if options.PermissionsPolicy != "" {
				res.Headers.Add("Permissions-Policy", options.PermissionsPolicy)
			}

			// HSTS header
			if options.EnableHSTS {
				var hsts strings.Builder
				hsts.WriteString("max-age=")
				hsts.WriteString(string(rune(options.HSTSMaxAge)))
				
				if options.HSTSIncludeSubdomains {
					hsts.WriteString("; includeSubDomains")
				}
				
				if options.HSTSPreload {
					hsts.WriteString("; preload")
				}
				
				res.Headers.Add("Strict-Transport-Security", hsts.String())
			}

			if options.EnableCSRF && (req.Method == "POST" || req.Method == "PUT" || 
				req.Method == "PATCH" || req.Method == "DELETE") {
				
				csrfCookie, hasCookie := req.Headers["cookie"]
				if !hasCookie {
					return errors.New("missing CSRF cookie")
				}
				
				cookieTokenParts := strings.Split(csrfCookie, options.CSRFCookieName+"=")
				if len(cookieTokenParts) < 2 {
					return errors.New("invalid CSRF cookie")
				}
				
				cookieToken := strings.Split(cookieTokenParts[1], ";")[0]
				
				headerToken, hasHeader := req.Headers[strings.ToLower(options.CSRFHeaderName)]
				if !hasHeader || headerToken == "" {
					return errors.New("missing CSRF token in header")
				}
				
				if cookieToken != headerToken {
					return errors.New("CSRF token mismatch")
				}
			}
			
			return nil
		},
	}
}

func GenerateCSRFToken(length int) (string, error) {
	if length <= 0 {
		length = 32
	}
	
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func CSRFTokenMiddleware(options *SecurityOptions) Middleware {
	if options == nil {
		defaultOptions := DefaultSecurityOptions()
		options = &defaultOptions
	}

	return Middleware{
		Handler: func(res *Response, req *Request) error {
			if !options.EnableCSRF {
				return nil
			}
			
			if req.Method != "GET" {
				return nil
			}
			
			token, err := GenerateCSRFToken(options.CSRFTokenLength)
			if err != nil {
				return err
			}
			
			cookieOpts := options.CSRFCookieOpts
			if cookieOpts == nil {
				cookieOpts = &CookieOptions{
					Path:     "/",
					MaxAge:   86400,
					Secure:   true,
					HttpOnly: true,
					SameSite: http.SameSiteStrictMode,
				}
			}
			
			var cookie strings.Builder
			cookie.WriteString(options.CSRFCookieName)
			cookie.WriteString("=")
			cookie.WriteString(token)
			cookie.WriteString("; Path=")
			cookie.WriteString(cookieOpts.Path)
			
			if cookieOpts.Domain != "" {
				cookie.WriteString("; Domain=")
				cookie.WriteString(cookieOpts.Domain)
			}
			
			if cookieOpts.MaxAge > 0 {
				cookie.WriteString("; Max-Age=")
				cookie.WriteString(string(rune(cookieOpts.MaxAge)))
			}
			
			if cookieOpts.Secure {
				cookie.WriteString("; Secure")
			}
			
			if cookieOpts.HttpOnly {
				cookie.WriteString("; HttpOnly")
			}
			
			switch cookieOpts.SameSite {
			case http.SameSiteLaxMode:
				cookie.WriteString("; SameSite=Lax")
			case http.SameSiteStrictMode:
				cookie.WriteString("; SameSite=Strict")
			case http.SameSiteNoneMode:
				cookie.WriteString("; SameSite=None")
			}
			
			res.Headers.Add("Set-Cookie", cookie.String())
			
			req.SetData("csrfToken", token)
			
			return nil
		},
	}
}

func GetCSRFToken(req *Request) (string, error) {
	var tokenInterface interface{}
	err := req.GetData("csrfToken", &tokenInterface)
	if err != nil {
		return "", errors.New("CSRF token not found in request")
	}
	
	token, ok := tokenInterface.(string)
	if !ok {
		return "", errors.New("invalid CSRF token type")
	}
	
	return token, nil
}
