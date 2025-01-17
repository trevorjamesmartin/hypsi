package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type Publisher interface {
	register(subscriber *Subscriber)
	unregister(subscriber *Subscriber)
	notify(path, event string)
	observe()
}

type Subscriber interface {
	receive(path, event string)
}

type PathWatcher struct {
	subscribers []*Subscriber
	rootPath    string
}

func (pw *PathWatcher) register(subscriber *Subscriber) {
	pw.subscribers = append(pw.subscribers, subscriber)
}

func (pw *PathWatcher) unregister(subscriber *Subscriber) {
	length := len(pw.subscribers)
	for i, sb := range pw.subscribers {
		if sb == subscriber {
			// swap
			pw.subscribers[i] = pw.subscribers[length-1]
			// pop()
			pw.subscribers = pw.subscribers[:length-1]
			break
		}
	}
}
func (pw *PathWatcher) notify(path, event string) {
	for _, sb := range pw.subscribers {
		(*sb).receive(path, event)
	}

}
func (pw *PathWatcher) observe() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("Error", err)
	}
	defer watcher.Close()

	if err := filepath.Walk(pw.rootPath,
		func(path string, info os.FileInfo, err error) error {
			if info.Mode().IsDir() {
				fmt.Printf("\nwatching path: [%s]", path)
				return watcher.Add(path)
			}

			return nil
		}); err != nil {
		fmt.Println("ERROR", err)
	}

	done := make(chan bool)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				pw.notify(event.Name, event.Op.String())
			case err := <-watcher.Errors:
				fmt.Println("Error", err)
			}
		}
	}()

	<-done
}

func NewPathWatcher(path string) Publisher {
	var pathWatcher Publisher = &PathWatcher{
		rootPath: path,
	}
	return pathWatcher
}
