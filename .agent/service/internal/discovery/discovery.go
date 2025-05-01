package discovery

import (
	"agentt/internal/config"
	// "agentt/internal/content" // Unused
	"agentt/internal/store"
	// "bytes" // Unused
	"context"
	"errors" // Re-add import
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// Watcher manages the discovery and watching of guidance files.
type Watcher struct {
	cfg         *config.ServiceConfig
	store       *store.GuidanceStore
	watcher     *fsnotify.Watcher
	mu          sync.Mutex // Protects access to watchedDirs
	watchedDirs map[string]bool
	configDir   string // Store the directory containing the config file
}

// NewWatcher creates a new file watcher and discovery manager.
func NewWatcher(cfg *config.ServiceConfig, store *store.GuidanceStore, cfgPath string) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file system watcher: %w", err)
	}
	// Ensure the config path is absolute before getting its directory
	absCfgPath, err := filepath.Abs(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for config file %s: %w", cfgPath, err)
	}
	cfgDir := filepath.Dir(absCfgPath)

	return &Watcher{
		cfg:         cfg,
		store:       store,
		watcher:     fsWatcher,
		watchedDirs: make(map[string]bool),
		configDir:   cfgDir, // Store the directory
	}, nil
}

// InitialScan performs the first scan of all configured paths.
func (w *Watcher) InitialScan() error {
	log.Println("Performing initial scan of guidance files...")
	var initialScanWg sync.WaitGroup

	for _, entityDef := range w.cfg.EntityTypes {
		// Construct glob relative to the config file's directory
		// PathGlob is assumed to be relative to the config file's location
		globPattern := filepath.Join(w.configDir, entityDef.PathGlob)
		log.Printf("Scanning for %s entities using glob relative to config dir: %s", entityDef.Name, globPattern)

		matches, err := filepath.Glob(globPattern)
		if err != nil {
			log.Printf("Warning: Error evaluating glob pattern '%s': %v", globPattern, err)
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
	item, err := ParseFile(filePath, entityDef)
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
	if err := w.store.AddOrUpdate(item); err != nil {
		// Handle potential duplicate ID error - Treat as FATAL during discovery
		if errors.Is(err, store.ErrDuplicateID) {
			log.Fatalf("FATAL: Configuration error: %v", err) // Make it fatal here
		} else {
			// Log other unexpected errors from AddOrUpdate
			log.Printf("ERROR: Unexpected error adding/updating item %s: %v", item.SourcePath, err)
		}
		// Don't log success message or watch dir if there was an error
		return
	}

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
					// Directory might have been removed while processing, ignore error
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
