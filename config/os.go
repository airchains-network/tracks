package config

import (
	"errors"
	"fmt"
	"github.com/cometbft/cometbft/libs/log"
	"io"
	"os"
	"os/signal"
	"syscall"
)

// Logger defines the logging interface that this package will use.
type Logger interface {
	Info(msg string, keyvals ...interface{})
}

// TrapSignal listens for SIGTERM/SIGINT signals and executes a callback function before exiting.
func TrapSignal(logger Logger, cb func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		for sig := range c {
			logger.Info("signal trapped", "msg", log.NewLazySprintf("captured %v, exiting...", sig))
			if cb != nil {
				cb()
			}
			os.Exit(0)
		}
	}()
}

// Kill sends a SIGTERM signal to the current process.
func Kill() error {
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		return err
	}
	return p.Signal(syscall.SIGTERM)
}

// Exit prints a message to standard output and exits with status code 1.
func Exit(message string) {
	fmt.Println(message)
	os.Exit(1)
}

// EnsureDir checks if the specified directory exists and creates it with the specified
// FileMode if it does not.
func EnsureDir(dir string, mode os.FileMode) error {
	if err := os.MkdirAll(dir, mode); err != nil {
		return fmt.Errorf("could not create directory %q: %w", dir, err)
	}
	return nil
}

// FileExists checks if the specified file path exists.
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// ReadFile reads and returns the content of the specified file.
func ReadFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

// MustReadFile reads the specified file and panics if an error occurs.
func MustReadFile(filePath string) []byte {
	data, err := ReadFile(filePath)
	if err != nil {
		Exit(fmt.Sprintf("MustReadFile failed: %v", err))
	}
	return data
}

// WriteFile writes the specified contents to a file at the given path.
func WriteFile(filePath string, contents []byte, mode os.FileMode) error {
	return os.WriteFile(filePath, contents, mode)
}

// MustWriteFile writes the specified contents to a file and panics if an error occurs.
func MustWriteFile(filePath string, contents []byte, mode os.FileMode) {
	if err := WriteFile(filePath, contents, mode); err != nil {
		Exit(fmt.Sprintf("MustWriteFile failed: %v", err))
	}
}

// CopyFile copies the contents of the source file to the destination file.
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	info, err := sourceFile.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		return errors.New("source cannot be a directory")
	}

	// Create or truncate the destination file with the same permissions as the source
	destFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
