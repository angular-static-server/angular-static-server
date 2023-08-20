package config

import (
	"log/slog"
	"path"

	"github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
	watcher    *fsnotify.Watcher
	watchables map[string][]WatchableFile
}

type WatchableFile interface {
	HandleChange()
	Dir() string
	Name() string
}

func CreateFileWatcher() *FileWatcher {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Warn("Failed to create file watcher", "error", err)
		return nil
	}

	fileWatcher := &FileWatcher{
		watcher:    watcher,
		watchables: make(map[string][]WatchableFile),
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					dir := path.Dir(event.Name)
					name := path.Base(event.Name)
					watchables, ok := fileWatcher.watchables[dir]
					if ok {
						for _, watchable := range watchables {
							if watchable.Name() == name {
								watchable.HandleChange()
							}
							return
						}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				slog.Error("file watcher encountered an error", "error", err)
			}
		}
	}()

	return fileWatcher
}

func (fileWatcher FileWatcher) Watch(watchable WatchableFile) error {
	dir := watchable.Dir()
	watchables, ok := fileWatcher.watchables[watchable.Dir()]
	if ok {
		fileWatcher.watchables[watchable.Dir()] = append(watchables, watchable)
		return nil
	} else {
		fileWatcher.watchables[watchable.Dir()] = []WatchableFile{watchable}
		return fileWatcher.watcher.Add(dir)
	}
}

func (fileWatcher FileWatcher) Close() error {
	if fileWatcher.watcher == nil {
		return nil
	}
	return fileWatcher.watcher.Close()
}
