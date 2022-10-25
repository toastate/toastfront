package server

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/toastate/toastfront/internal/builder"
	"github.com/toastate/toastfront/internal/tlogger"
	"github.com/toastate/toastfront/internal/watcher"

	_ "embed"
)

//go:embed livereload.html
var liveReloadScript []byte

var upgrader = websocket.Upgrader{
	HandshakeTimeout: 10 * time.Second,
	// Subprotocols:     []string{},
	Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		w.WriteHeader(500)
	},
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Server struct {
	sourceDir    string
	buildDir     string
	rootDir      string
	port         string
	override404  string
	reloadBroker *Broker
	buildtool    *builder.Builder
}

func (s *Server) TriggerReload() {
	s.reloadBroker.Publish(struct{}{})
}

func NewServer(sourceDir, buildDir, rootDir string, port string, override404 string) *Server {
	s := &Server{
		sourceDir:    sourceDir,
		rootDir:      rootDir,
		buildDir:     buildDir,
		port:         port,
		override404:  override404,
		reloadBroker: newBroker(),
		buildtool:    builder.NewBuilder(sourceDir, buildDir, rootDir),
	}

	return s
}

func (s *Server) Start(withBuilder bool) error {
	if withBuilder {
		err := s.buildtool.Init()
		if err != nil {
			return err
		}

		buildStart := time.Now()
		err = s.buildtool.Build()
		estBuildTime := time.Since(buildStart)
		estBuildTime *= 2
		if estBuildTime > time.Millisecond*500 {
			estBuildTime = time.Millisecond * 500
		}
		if err != nil {
			os.Exit(1)
		}

		updates := watcher.StartWatcher(s.sourceDir)

		go s.reloadBroker.Start()

		go func() {
			for {
				<-updates
			rootFor:
				for {
					select {
					case <-updates:
						continue
					case <-time.After(time.Millisecond * 500):
						break rootFor
					}
				}
				s.buildtool.Build()
				s.TriggerReload()
			}
		}()
	}

	r := mux.NewRouter()
	r.PathPrefix("/").HandlerFunc(s.fileServer(s.buildDir, s.override404))

	// We use println here so the address can be copied or opened directly from the terminal
	fmt.Println("Listening on http://localhost:" + s.port)

	return http.ListenAndServe(":"+s.port, r)
}

func (s *Server) fileServer(dir string, override404 string) func(http.ResponseWriter, *http.Request) {
	if override404 != "" && !strings.HasPrefix(override404, "/") {
		override404 = "/" + override404
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/__internal/livereload" {
			s.livereloadHandler(w, r)
			return
		}
	begin:
		upath := r.URL.Path
		if !strings.HasPrefix(upath, "/") {
			upath = "/" + upath
			r.URL.Path = upath
		}

		const indexPage = "index.html"

		fullName := filepath.Join(dir, filepath.FromSlash(path.Clean(upath)))

		if fullName[len(fullName)-1] == '/' {
			fullName = filepath.Join(fullName, indexPage)
		}

		info, err := os.Stat(fullName)

		valid := false
		if err != nil || info.IsDir() {
			if err != nil && !os.IsNotExist(err) {
				w.WriteHeader(500)
				w.Write([]byte("Internal error: can't open file: " + err.Error()))
				return
			}

			info, err = os.Stat(fullName + ".html")
			if err != nil || info.IsDir() {
				if err != nil && !os.IsNotExist(err) {
					w.WriteHeader(500)
					w.Write([]byte("Internal error: can't open file: " + err.Error()))
					return
				}

				info, err := os.Stat(filepath.Join(fullName, indexPage))
				if err != nil || info.IsDir() {
					if err != nil && !os.IsNotExist(err) {
						w.WriteHeader(500)
						w.Write([]byte("Internal error: can't open file: " + err.Error()))
					}
				} else {
					fullName = filepath.Join(fullName, indexPage)
					valid = true
				}
			} else {
				fullName = fullName + ".html"
				valid = true
			}
		} else {
			valid = true
		}

		if !valid {
			if override404 != "" && r.URL.Path != override404 {
				r.URL.Path = override404
				goto begin
			}
			w.WriteHeader(404)
			w.Write([]byte("404 page not found"))
			return
		}

		content, err := os.Open(fullName)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Internal error: can't open file"))
		}

		ctype := mime.TypeByExtension(filepath.Ext(fullName))
		if ctype == "" {
			// read a chunk to decide between utf-8 text and binary
			var buf [512]byte
			n, _ := io.ReadFull(content, buf[:])
			ctype = http.DetectContentType(buf[:n])
			_, err := content.Seek(0, io.SeekStart) // rewind to output whole file
			if err != nil {
				w.WriteHeader(500)
				w.Write([]byte("Internal error: can't seek file: " + err.Error()))
			}
		}
		w.Header().Set("Content-Type", ctype)
		io.Copy(w, content)
		if strings.HasPrefix(ctype, "text/html") {
			_, err = w.Write(liveReloadScript)
			if err != nil {
				tlogger.Error("msg", "could not live reload", "error", err)
			}
		}
	}
}

func (s *Server) livereloadHandler(w http.ResponseWriter, r *http.Request) {
	tlogger.Debug("msg", "WS Established")

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	defer c.Close()
	waitCh := s.reloadBroker.Subscribe()
	<-waitCh
	err = c.WriteMessage(websocket.TextMessage, []byte("reload"))
	if err != nil {
		tlogger.Warn("msg", "Reload socket error", "error", err)
	}
	s.reloadBroker.Unsubscribe(waitCh)
}
