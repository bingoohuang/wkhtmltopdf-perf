package wkp

import (
	"github.com/fsnotify/fsnotify"
	"log"
)

func GoFsNotify(path string) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func() {
		defer w.Close()

		for {
			select {
			case e, ok := <-w.Events:
				if !ok {
					return
				}
				log.Println("event:", e)
				if e.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", e.Name)
				}
			case err, ok := <-w.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	if err := w.Add(path); err != nil {
		return err
	}
	return nil
}
