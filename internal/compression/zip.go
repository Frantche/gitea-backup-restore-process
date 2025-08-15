package compression

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Frantche/gitea-backup-restore-process/internal/config"
	"github.com/Frantche/gitea-backup-restore-process/pkg/logger"
)

// CreateZip creates a zip archive from the backup tmp folder
func CreateZip(settings *config.Settings) error {
	logger.Info("Creating zip archive")
	
	zipFile, err := os.Create(settings.BackupTmpFilename)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()
	
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()
	
	err = filepath.Walk(settings.BackupTmpFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip the root folder itself
		if path == settings.BackupTmpFolder {
			return nil
		}
		
		// Get relative path
		relPath, err := filepath.Rel(settings.BackupTmpFolder, path)
		if err != nil {
			return err
		}
		
		// Normalize path separators for zip
		relPath = strings.ReplaceAll(relPath, "\\", "/")
		
		if info.IsDir() {
			// Add directory to zip (with trailing slash)
			_, err := zipWriter.Create(relPath + "/")
			return err
		}
		
		// Add file to zip
		fileWriter, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}
		
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		
		_, err = io.Copy(fileWriter, file)
		return err
	})
	
	if err != nil {
		return fmt.Errorf("failed to create zip archive: %w", err)
	}
	
	logger.Info("Zip archive created successfully")
	return nil
}

const (
	dirPerm  = 0o700 // rwx owner
	filePerm = 0o700 // rwx owner (change to 0o600 if you want rw only for files)
)

var copyBufPool = sync.Pool{
	New: func() any {
		// 256 KiB buffer tends to perform well for zip streams
		buf := make([]byte, 256<<10)
		return buf
	},
}

// ExtractZip extracts a zip archive to the restore tmp folder with rwx perms.
func ExtractZip(settings *config.Settings) error {
	logger.Info("Extracting zip archive")

	// Ensure destination root exists with rwx
	if err := os.MkdirAll(settings.RestoreTmpFolder, dirPerm); err != nil {
		return fmt.Errorf("failed to create restore tmp folder: %w", err)
	}
	_ = os.Chmod(settings.RestoreTmpFolder, dirPerm)

	zr, err := zip.OpenReader(settings.RestoreTmpFilename)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer zr.Close()

	base := filepath.Clean(settings.RestoreTmpFolder)
	for _, f := range zr.File {
		if err := extractFileRWX(f, base); err != nil {
			return fmt.Errorf("failed to extract %s: %w", f.Name, err)
		}
	}

	logger.Info("Zip archive extracted successfully")
	return nil
}

func extractFileRWX(f *zip.File, destDir string) error {
	// Zip Slip prevention
	dest := filepath.Join(destDir, f.Name)
	if !strings.HasPrefix(dest, destDir+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", f.Name)
	}

	mode := f.Mode()

	// Directories
	if f.FileInfo().IsDir() {
		if err := os.MkdirAll(dest, dirPerm); err != nil {
			return err
		}
		return os.Chmod(dest, dirPerm)
	}

	// Skip symlinks (safer default). If you need them, validate target then os.Symlink.
	if mode&os.ModeSymlink != 0 {
		logger.Debugf("Skipping symlink from zip: %s", f.Name)
		return nil
	}

	// Ensure parent dir exists (with rwx)
	if err := os.MkdirAll(filepath.Dir(dest), dirPerm); err != nil {
		return err
	}

	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	// Create file; set perms after write to override umask
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, filePerm)
	if err != nil {
		return err
	}
	buf := copyBufPool.Get().([]byte)
	_, cpErr := io.CopyBuffer(out, rc, buf)
	putErr := out.Close()
	copyBufPool.Put(buf)
	if cpErr != nil {
		return cpErr
	}
	if putErr != nil {
		return putErr
	}

	// Force desired perms (ensures rwx even if umask interfered)
	return os.Chmod(dest, filePerm)
}