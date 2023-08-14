package config

import (
	"ngstaticserver/test"
	"path/filepath"
	"testing"
	"time"
)

func TestShouldUpdateDotEnvOnChange(t *testing.T) {
	context := test.NewTestDir(t)
	context.CreateFile(".env", "ENV =production\nPORT =8080 \nDELAY = 200")

	fileWatcher := CreateFileWatcher()
	test.AssertTrue(t, fileWatcher.watcher != nil, "")
	t.Cleanup(func() {
		fileWatcher.Close()
	})
	var result map[string]*string
	env := CreateDotEnv(filepath.Join(context.Path, ".env"), func(variables map[string]*string) {
		result = variables
	})
	err := fileWatcher.Watch(env)
	if err != nil {
		t.Error(err)
	}

	test.AssertEqual(t, len(result), 3, "")

	context.CreateFile(".env", "TEST = example")

	time.Sleep(time.Millisecond)

	test.AssertEqual(t, len(result), 1, "")
	test.AssertEqual(t, readValue(t, result, "TEST"), "example", "")
}
