// Package config provides file watching capabilities for configuration files.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileWatcher watches configuration files for changes
type FileWatcher struct {
	watcher   *fsnotify.Watcher
	callbacks map[string]func()
	mu        sync.RWMutex
	done      chan bool
}

// NewFileWatcher creates a new file watcher
func NewFileWatcher() *FileWatcher {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		// Fallback to polling-based watcher if fsnotify is not available
		return NewPollingFileWatcher()
	}

	fw := &FileWatcher{
		watcher:   watcher,
		callbacks: make(map[string]func()),
		done:      make(chan bool),
	}

	go fw.watchLoop()
	return fw
}

// Watch starts watching a file for changes
func (fw *FileWatcher) Watch(filePath string, callback func()) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	// Convert to absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Store callback
	fw.callbacks[absPath] = callback

	// Add file to watcher
	return fw.watcher.Add(absPath)
}

// Unwatch stops watching a file
func (fw *FileWatcher) Unwatch(filePath string) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	delete(fw.callbacks, absPath)
	return fw.watcher.Remove(absPath)
}

// Close closes the file watcher
func (fw *FileWatcher) Close() error {
	close(fw.done)
	return fw.watcher.Close()
}

// watchLoop handles file system events
func (fw *FileWatcher) watchLoop() {
	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}

			// Handle write events (file modifications)
			if event.Op&fsnotify.Write == fsnotify.Write {
				fw.mu.RLock()
				if callback, exists := fw.callbacks[event.Name]; exists {
					// Debounce rapid fire events
					go func() {
						time.Sleep(100 * time.Millisecond)
						callback()
					}()
				}
				fw.mu.RUnlock()
			}

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("File watcher error: %v\n", err)

		case <-fw.done:
			return
		}
	}
}

// PollingFileWatcher is a fallback file watcher that uses polling
type PollingFileWatcher struct {
	files     map[string]*fileInfo
	callbacks map[string]func()
	mu        sync.RWMutex
	ticker    *time.Ticker
	done      chan bool
}

type fileInfo struct {
	modTime time.Time
	size    int64
}

// NewPollingFileWatcher creates a new polling-based file watcher
func NewPollingFileWatcher() *FileWatcher {
	pfw := &PollingFileWatcher{
		files:     make(map[string]*fileInfo),
		callbacks: make(map[string]func()),
		ticker:    time.NewTicker(1 * time.Second),
		done:      make(chan bool),
	}

	go pfw.pollLoop()

	// Create a wrapper to match FileWatcher interface
	wrapper := &FileWatcher{
		callbacks: make(map[string]func()),
		done:      make(chan bool),
	}

	// Store reference for delegation
	wrapper.watcher = nil // Will use polling watcher methods instead

	return wrapper
}

// Watch starts watching a file using polling
func (pfw *PollingFileWatcher) Watch(filePath string, callback func()) error {
	pfw.mu.Lock()
	defer pfw.mu.Unlock()

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Get initial file info
	stat, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	pfw.files[absPath] = &fileInfo{
		modTime: stat.ModTime(),
		size:    stat.Size(),
	}
	pfw.callbacks[absPath] = callback

	return nil
}

// Unwatch stops watching a file
func (pfw *PollingFileWatcher) Unwatch(filePath string) error {
	pfw.mu.Lock()
	defer pfw.mu.Unlock()

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	delete(pfw.files, absPath)
	delete(pfw.callbacks, absPath)
	return nil
}

// Close closes the polling file watcher
func (pfw *PollingFileWatcher) Close() error {
	pfw.ticker.Stop()
	close(pfw.done)
	return nil
}

// pollLoop checks files for changes at regular intervals
func (pfw *PollingFileWatcher) pollLoop() {
	for {
		select {
		case <-pfw.ticker.C:
			pfw.checkFiles()
		case <-pfw.done:
			return
		}
	}
}

// checkFiles checks all watched files for modifications
func (pfw *PollingFileWatcher) checkFiles() {
	pfw.mu.RLock()
	defer pfw.mu.RUnlock()

	for filePath, info := range pfw.files {
		stat, err := os.Stat(filePath)
		if err != nil {
			continue // File may have been deleted
		}

		// Check if file was modified
		if stat.ModTime().After(info.modTime) || stat.Size() != info.size {
			// Update stored info
			info.modTime = stat.ModTime()
			info.size = stat.Size()

			// Call callback
			if callback, exists := pfw.callbacks[filePath]; exists {
				go callback()
			}
		}
	}
}

// FileWatcherOptions provides options for file watching
type FileWatcherOptions struct {
	PollInterval time.Duration
	UsePolling   bool
	DebounceTime time.Duration
}

// DefaultFileWatcherOptions returns default file watcher options
func DefaultFileWatcherOptions() FileWatcherOptions {
	return FileWatcherOptions{
		PollInterval: 1 * time.Second,
		UsePolling:   false,
		DebounceTime: 100 * time.Millisecond,
	}
}

// NewFileWatcherWithOptions creates a new file watcher with options
func NewFileWatcherWithOptions(opts FileWatcherOptions) *FileWatcher {
	if opts.UsePolling {
		return NewPollingFileWatcherWithInterval(opts.PollInterval)
	}
	return NewFileWatcher()
}

// NewPollingFileWatcherWithInterval creates a polling file watcher with custom interval
func NewPollingFileWatcherWithInterval(interval time.Duration) *FileWatcher {
	pfw := &PollingFileWatcher{
		files:     make(map[string]*fileInfo),
		callbacks: make(map[string]func()),
		ticker:    time.NewTicker(interval),
		done:      make(chan bool),
	}

	go pfw.pollLoop()

	wrapper := &FileWatcher{
		callbacks: make(map[string]func()),
		done:      make(chan bool),
	}

	return wrapper
}

// BatchFileWatcher watches multiple files efficiently
type BatchFileWatcher struct {
	watcher   *FileWatcher
	callbacks []func()
	mu        sync.RWMutex
}

// NewBatchFileWatcher creates a new batch file watcher
func NewBatchFileWatcher() *BatchFileWatcher {
	return &BatchFileWatcher{
		watcher: NewFileWatcher(),
	}
}

// AddFile adds a file to the batch watcher
func (bfw *BatchFileWatcher) AddFile(filePath string, callback func()) error {
	bfw.mu.Lock()
	defer bfw.mu.Unlock()

	bfw.callbacks = append(bfw.callbacks, callback)

	return bfw.watcher.Watch(filePath, func() {
		bfw.mu.RLock()
		for _, cb := range bfw.callbacks {
			go cb()
		}
		bfw.mu.RUnlock()
	})
}

// Close closes the batch file watcher
func (bfw *BatchFileWatcher) Close() error {
	return bfw.watcher.Close()
}
