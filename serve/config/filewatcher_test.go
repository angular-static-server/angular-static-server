package config

import (
	"ngstaticserver/test"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type testEnvState struct {
	env map[string]*string
}

func (env *testEnvState) handleChange(variables map[string]*string) {
	env.env = variables
}

func TestShouldUpdateDotEnvOnChange(t *testing.T) {
	context := test.NewTestDir(t)
	envFilePath := filepath.Join(context.Path, "../config/.env")
	os.WriteFile(envFilePath, []byte("ENV =production\nPORT =8080 \nDELAY = 200"), 0644)

	fileWatcher := CreateFileWatcher()
	test.AssertTrue(t, fileWatcher.watcher != nil)
	t.Cleanup(func() {
		fileWatcher.Close()
	})
	testEnv := &testEnvState{make(map[string]*string)}
	env := CreateDotEnv(context.Path, testEnv.handleChange)
	err := fileWatcher.Watch(env)
	test.AssertNoError(t, err)

	test.AssertEqual(t, len(testEnv.env), 3)

	os.WriteFile(envFilePath, []byte("TEST = example"), 0644)

	// This test is flaky on GitHub Actions, so we do this workaround
	counter := 0
	for counter < 20 && len(testEnv.env) == 3 {
		time.Sleep(time.Millisecond * 50)
		counter++
	}

	test.AssertEqual(t, len(testEnv.env), 1)
	test.AssertEqual(t, readValue(t, testEnv.env, "TEST"), "example")
}
