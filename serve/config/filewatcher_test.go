package config

import (
	"ngstaticserver/test"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestShouldUpdateDotEnvOnChange(t *testing.T) {
	context := test.NewTestDir(t)
	envFilePath := filepath.Join(context.Path, "../config/.env")
	os.WriteFile(envFilePath, []byte("ENV =production\nPORT =8080 \nDELAY = 200"), 0644)

	fileWatcher := CreateFileWatcher()
	test.AssertTrue(t, fileWatcher.watcher != nil)
	t.Cleanup(func() {
		fileWatcher.Close()
	})
	var result map[string]*string
	env := CreateDotEnv(context.Path, func(variables map[string]*string) {
		result = variables
	})
	err := fileWatcher.Watch(env)
	test.AssertNoError(t, err)

	test.AssertEqual(t, len(result), 3)

	os.WriteFile(envFilePath, []byte("TEST = example"), 0644)

	time.Sleep(time.Millisecond)

	test.AssertEqual(t, len(result), 1)
	test.AssertEqual(t, readValue(t, result, "TEST"), "example")
}
