package server

import (
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
	"github.com/toastate/toastfront/internal/tlogger"

	_ "embed"
)

//go:embed livereload.html
var liveReloadScript []byte

var upgrader = websocket.Upgrader{
	HandshakeTimeout: 10 * time.Second,
	// Subprotocols:     []string{},
	Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		w.WriteHeader(500)
		return
	},
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var reloadBroker = NewBroker()

func init() {
	go reloadBroker.Start()
}
func TriggerReload() {
	reloadBroker.Publish(struct{}{})
}

func Start(buildDir string, port string, override404 string) {
	r := mux.NewRouter()
	r.PathPrefix("/").HandlerFunc(fileServer(buildDir, override404))
	http.Handle("/", r)
	tlogger.Warn("msg", "Listening", "port", port)
	http.ListenAndServe(":"+port, nil)
}

func fileServer(dir string, override404 string) func(http.ResponseWriter, *http.Request) {
	if override404 != "" && !strings.HasPrefix(override404, "/") {
		override404 = "/" + override404
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/__internal/livereload" {
			livereloadHandler(w, r)
			return
		}
	begin:
		upath := r.URL.Path
		if !strings.HasPrefix(upath, "/") {
			upath = "/" + upath
			r.URL.Path = upath
		}

		const indexPage = "index.html"

		fullName := filepath.Join(dir, filepath.FromSlash(path.Clean("/"+upath)))

		if fullName[len(fullName)-1] == '/' {
			fullName = filepath.Join(fullName, indexPage)
		}

		info, err := os.Stat(fullName)

		valid := false
		if err != nil || info.IsDir() {
			if err != nil && !os.IsNotExist(err) {
				w.WriteHeader(500)
				w.Write([]byte("Internal error: can't open file: " + err.Error()))
			}
			fullName = filepath.Join(fullName, indexPage)
			info, err := os.Stat(fullName)
			if err != nil || info.IsDir() {
				if err != nil && !os.IsNotExist(err) {
					w.WriteHeader(500)
					w.Write([]byte("Internal error: can't open file: " + err.Error()))
				}
			} else {
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
		}
	}
}

func livereloadHandler(w http.ResponseWriter, r *http.Request) {
	tlogger.Debug("msg", "WS Established")

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	defer c.Close()
	waitCh := reloadBroker.Subscribe()
	<-waitCh
	err = c.WriteMessage(websocket.TextMessage, []byte("reload"))
	if err != nil {
		tlogger.Warn("msg", "Reload socket error", "error", err)
	}
	reloadBroker.Unsubscribe(waitCh)
}
