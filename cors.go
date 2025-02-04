package gonanoweb

type CorsOptions struct {
	Origins []string
	Headers map[string]string
}

func (s *Server) Cors(cors CorsOptions) {
	s.CorsOptions = &cors
}
