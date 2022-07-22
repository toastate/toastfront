package watcher

import (
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/toastate/toastfront/internal/tlogger"
)

func StartWatcher(folder string) <-chan string {
	wch, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	outCh := make(chan string, 100)

	go func() {
		for {
			select {
			case event, ok := <-wch.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					tlogger.Info("msg", "Detected change", "path", event.Name)
					outCh <- event.Name
				}
			case err, ok := <-wch.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	filepath.Walk(folder, func(path string, fi os.FileInfo, err error) error {
		if fi.IsDir() {
			return wch.Add(path)
		}
		return nil
	})

	return outCh
}
