package api

import (
	"fmt"
	"net/http"

	"github.com/spf13/viper"
)

type Server struct {
	httpServer *http.Server
}

func NewServer() *Server {
	viper.SetEnvPrefix("nux")
	viper.AutomaticEnv()

	mux := http.NewServeMux()

	port := viper.GetString("api_port")
	if port == "" {
		port = "8080"
	}

	return &Server{
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%s", port),
			Handler: mux,
		},
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}
