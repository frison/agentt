package discovery

import (
	"agent-guidance-service/internal/config"
	"agent-guidance-service/internal/store"
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher manages the discovery and watching of guidance files.
type Watcher struct {
	cfg    *config.ServiceConfig
	store  *store.GuidanceStore
	watcher *fsnotify.Watcher
	mu     sync.Mutex // Protects access to watchedDirs
	watchedDirs map[string]bool
}

// NewWatcher creates a new file watcher and discovery manager.
func NewWatcher(cfg *config.ServiceConfig, store *store.GuidanceStore) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file system watcher: %w", err)
	}
	return &Watcher{
		cfg:         cfg,
		store:       store,
		watcher:     fsWatcher,
		watchedDirs: make(map[string]bool),
	}, nil
}

// InitialScan performs the first scan of all configured paths.
func (w *Watcher) InitialScan() error {
	log.Println("Performing initial scan of guidance files...")
	var initialScanWg sync.WaitGroup

	for _, entityDef := range w.cfg.EntityTypes {
		log.Printf("Scanning for %s entities using glob: %s", entityDef.Name, entityDef.PathGlob)
		matches, err := filepath.Glob(entityDef.PathGlob)
		if err != nil {
			log.Printf("Warning: Error evaluating glob pattern '%s': %v", entityDef.PathGlob, err)
			continue
		}

		for _, match := range matches {
			initialScanWg.Add(1)
			go func(filePath string, def config.EntityTypeDefinition) {
				defer initialScanWg.Done()
				w.processFile(filePath, def)
			}(match, entityDef)
		}
	}
	initialScanWg.Wait()
	log.Println("Initial scan complete.")
	return nil
}

// processFile parses a file and updates the store, logging errors.
func (w *Watcher) processFile(filePath string, entityDef config.EntityTypeDefinition) {
	item, err := parseFile(filePath, entityDef)
	if err != nil {
		log.Printf("Error processing file %s: %v", filePath, err)
		return
	}
	if item == nil {
		// File might have been deleted between glob and parse, or other non-fatal issue
		return
	}

	if !item.IsValid {
		log.Printf("Warning: Invalid content detected in %s: %v", item.SourcePath, item.ValidationErrors)
	}
	w.store.AddOrUpdate(item)
	log.Printf("Processed: %s (Type: %s, Valid: %t)", item.SourcePath, item.EntityType, item.IsValid)

	// Ensure directory is watched
	w.watchDir(filepath.Dir(filePath))
}

// watchDir adds a directory to the fsnotify watcher if not already watched.
func (w *Watcher) watchDir(dirPath string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.watchedDirs[dirPath] {
		return // Already watching
	}

	// Watch recursively? fsnotify doesn't directly support recursive watching.
	// We need to walk the dir and add all subdirs.
	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Skip paths we can't access, log error
			log.Printf("Error accessing path %s during watch setup: %v", path, err)
			return filepath.SkipDir // Skip this whole subtree if we can't access its root
		}
		if d.IsDir() {
			if w.watchedDirs[path] {
				return nil // Already watching this specific subdir
			}
			err = w.watcher.Add(path)
			if err != nil {
				log.Printf("Error adding directory watcher for %s: %v", path, err)
				// Don't necessarily stop walking, maybe other dirs can be watched
			} else {
				log.Printf("Watching directory: %s", path)
				w.watchedDirs[path] = true
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("Error walking directory %s for watching: %v", dirPath, err)
	}
}

// Start runs the file watching loop.
func (w *Watcher) Start(ctx context.Context) {
	log.Println("Starting file watcher...")
	go func() {
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					log.Println("Watcher event channel closed.")
					return
				}
				log.Printf("Watcher event: %s Op: %s", event.Name, event.Op.String())

				// Determine the entity type based on the event path and config
				var matchingDef *config.EntityTypeDefinition
				for _, def := range w.cfg.EntityTypes {
					// Basic matching based on file extension hint or path structure
					if def.FileExtensionHint != "" && strings.HasSuffix(event.Name, def.FileExtensionHint) {
						// Stronger check: Does it match the glob pattern base?
						// This is complex, rely on extension/path structure for now.
						matched, _ := filepath.Match(def.PathGlob, event.Name)
						if matched {
							d := def // Create a local copy for the goroutine
							matchingDef = &d
							break
						}
					}
					// Add more sophisticated path matching if needed
				}

				if matchingDef == nil {
					// It might be a directory creation/deletion, handle watching
					if event.Has(fsnotify.Create) {
						fi, err := os.Stat(event.Name)
						if err == nil && fi.IsDir() {
							w.watchDir(event.Name) // Watch new directories
						}
					}
					// Ignore events for files not matching any entity type pattern
					continue
				}

				// Handle file changes
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					// Debounce? For now, process immediately.
					go w.processFile(event.Name, *matchingDef)
				} else if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
					// If a file is removed or renamed, remove it from the store.
					absPath, _ := filepath.Abs(event.Name)
					if absPath != "" {
						go func(p string) {
							w.store.Remove(p)
							log.Printf("Removed item from store due to file deletion/rename: %s", p)
						}(absPath)
					}
					// If a dir is removed, fsnotify might remove the watch automatically, but cleanup watchedDirs?
					// TODO: Handle directory removal from watcher more robustly if needed.
				}

			case err, ok := <-w.watcher.Errors:
				if !ok {
					log.Println("Watcher error channel closed.")
					return
				}
				log.Printf("Watcher error: %v", err)

			case <-ctx.Done():
				log.Println("Stopping file watcher due to context cancellation...")
				return
			}
		}
	}()
}

// Close stops the file watcher.
func (w *Watcher) Close() error {
	log.Println("Closing file watcher...")
	return w.watcher.Close()
}