package server

import "github.com/toastate/toastfront/internal/server"

type Server interface {
	Start(withBuilder bool) error
}

func NewServer(sourceDir, buildDir, rootDir string, port string, override404 string) Server {
	return server.NewServer(sourceDir, buildDir, rootDir, port, override404)
}
