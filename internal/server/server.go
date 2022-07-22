package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/toastate/toastfront/internal/tlogger"
)

func Start(buildDir string, port string) {
	r := mux.NewRouter()
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(buildDir)))
	http.Handle("/", r)
	tlogger.Warn("msg", "Listening", "port", port)
	http.ListenAndServe(":"+port, nil)
}
