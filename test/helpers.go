package test

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
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

func (context TestDir) WriteFile(fileName string, content string) {
	context.t.Helper()
	filePath := filepath.Join(context.Path, fileName)
	os.WriteFile(filePath, []byte(content), 0644)
}

func (context TestDir) ImportTestApp(app string) {
	context.t.Helper()
	_, b, _, _ := runtime.Caller(0)
	testAppDir := filepath.Join(filepath.Dir(b), "angular/dist", app)
	if _, err := os.Stat(testAppDir); os.IsNotExist(err) {
		// Runs npm run build in the angular directory
		cmd := exec.Command("npm", "run", "build")
		cmd.Dir = filepath.Join(filepath.Dir(b), "angular")
		output, err := cmd.CombinedOutput()
		if err != nil {
			context.t.Fatal(err)
		}
		fmt.Println(output)
	}
	copyDir(testAppDir, context.Path)
}

func (context TestDir) CompressFile(filePath string) {
	context.t.Helper()
	content := context.ReadFile(filePath)
	CompressToFile([]byte(content), filepath.Join(context.Path, filePath))
}

func (context TestDir) FindFile(prefix string) string {
	context.t.Helper()
	files := make([]string, 0)
	err := filepath.Walk(context.Path, func(path string, info os.FileInfo, err error) error {
		if err == nil && strings.Contains(path, "/"+prefix) {
			files = append(files, path)
		}
		return err
	})
	if err != nil {
		panic(err)
	}

	result, err := filepath.Rel(context.Path, files[0])
	if err != nil {
		panic(err)
	}
	return result
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

func AssertEqual(t *testing.T, a interface{}, b interface{}) {
	t.Helper()
	if a != b {
		t.Errorf("%v != %v", a, b)
	}
}

func AssertTrue(t *testing.T, v bool) {
	t.Helper()
	if !v {
		t.Error("Expected value to be true")
	}
}

func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("Expected error to be empty: %v", err)
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
