package utils

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"go.uber.org/zap"
)

// CleanupOldFiles removes files older than maxAge from the specified directory
func CleanupOldFiles(dir string, maxAge time.Duration) error {
	now := time.Now()

	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		// Check if file is older than maxAge
		if now.Sub(info.ModTime()) > maxAge {
			Logger.Info("Removing old file",
				zap.String("path", path),
				zap.Duration("age", now.Sub(info.ModTime())),
			)
			return os.Remove(path)
		}

		return nil
	})
}

// CheckFolderSize checks if the folder size exceeds maxSize and removes oldest files if needed
func CheckFolderSize(dir string, maxSize int64) error {
	var totalSize int64
	var files []fileInfo

	// Walk directory and collect file info
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		totalSize += info.Size()
		files = append(files, fileInfo{
			path:    path,
			size:    info.Size(),
			modTime: info.ModTime(),
		})

		return nil
	})

	if err != nil {
		return err
	}

	// If total size is within limit, we're done
	if totalSize <= maxSize {
		return nil
	}

	// Sort files by modification time (oldest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.Before(files[j].modTime)
	})

	// Remove oldest files until size is within limit
	for _, file := range files {
		if totalSize <= maxSize {
			break
		}

		Logger.Info("Removing file to reduce folder size",
			zap.String("path", file.path),
			zap.Int64("size", file.size),
		)
		if err := os.Remove(file.path); err != nil {
			Logger.Error("Failed to remove file",
				zap.String("path", file.path),
				zap.Error(err),
			)
			continue
		}

		totalSize -= file.size
	}

	return nil
}

type fileInfo struct {
	path    string
	size    int64
	modTime time.Time
}


// StartCleanupService starts a background service that periodically cleans up files
func StartCleanupService(dir string, maxAge time.Duration, maxSize int64, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			if err := CleanupOldFiles(dir, maxAge); err != nil {
				Logger.Error("Cleanup error", zap.Error(err))
			}
			if err := CheckFolderSize(dir, maxSize); err != nil {
				Logger.Error("Folder size check error", zap.Error(err))
			}
		}
	}()
}