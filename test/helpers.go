package test

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type TestDir struct {
	Path string
	t    *testing.T
}

func NewTestDir(t *testing.T) TestDir {
	t.Helper()
	dir := t.TempDir()
	return TestDir{Path: dir, t: t}
}

func (context TestDir) CreateFile(fileName string, content string) {
	context.t.Helper()
	filePath := filepath.Join(context.Path, fileName)
	os.WriteFile(filePath, []byte(content), 0644)
}

func (context TestDir) ImportTestNgsscApp() {
	context.t.Helper()
	_, b, _, _ := runtime.Caller(0)
	testAppDir := filepath.Join(filepath.Dir(b), "./ngssc-app")
	copyDir(testAppDir, context.Path)
}

func (context TestDir) ReadFile(fileName string) string {
	context.t.Helper()
	filePath := filepath.Join(context.Path, fileName)
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	return string(fileContent)
}

func (context TestDir) RemoveFile(fileName string) {
	context.t.Helper()
	filePath := filepath.Join(context.Path, fileName)
	err := os.Remove(filePath)
	if err != nil {
		panic(err)
	}
}

func (context TestDir) CreateDirectory(language string) TestDir {
	context.t.Helper()
	path := filepath.Join(context.Path, language)
	err := os.Mkdir(path, 0755)
	if err != nil {
		panic(err)
	}

	return TestDir{Path: path, t: context.t}
}

func Chdir(t *testing.T, dir string) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("restoring working directory: %v", err)
		}
	})
}

func AssertContains(t *testing.T, s string, substring string, message string) {
	t.Helper()
	debugMessage := fmt.Sprintf("Expected %v to contain %v", s, substring)
	AssertTrue(t, strings.Contains(s, substring), appendDebugMessage(message, debugMessage))
}

func AssertEqual(t *testing.T, a interface{}, b interface{}, message string) {
	t.Helper()
	AssertTrue(t, a == b, appendDebugMessage(message, fmt.Sprintf("%v != %v", a, b)))
}

func AssertTrue(t *testing.T, v bool, message string) {
	t.Helper()
	if v {
		return
	}
	if len(message) == 0 {
		message = "Expected value to be true"
	}
	t.Fatal(message)
}

func AssertNoError(t *testing.T, err error, message string) {
	t.Helper()
	if err == nil {
		return
	}
	if len(message) == 0 {
		message = "Expected error to be empty"
	}
	t.Fatal(message)
}

func appendDebugMessage(message string, debugMessage string) string {
	if len(message) == 0 {
		return debugMessage
	} else {
		return message + "\n" + debugMessage
	}
}

func copyDir(source, target string) error {
	return filepath.Walk(source, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		targetPath := filepath.Join(target, strings.TrimPrefix(path, source))
		if info.IsDir() {
			os.MkdirAll(targetPath, info.Mode())
			return nil
		} else if !info.Mode().IsRegular() {
			switch info.Mode().Type() & os.ModeType {
			case os.ModeSymlink:
				link, err := os.Readlink(path)
				if err != nil {
					return err
				}
				return os.Symlink(link, targetPath)
			}
			return nil
		} else {
			sourceFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer sourceFile.Close()

			targetFile, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			defer targetFile.Close()

			targetFile.Chmod(info.Mode())
			_, err = io.Copy(targetFile, sourceFile)
			return err
		}
	})
}
