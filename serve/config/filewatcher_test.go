package config

import (
	"fmt"
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
	os.WriteFile(envFilePath, []byte("ENV =production\nPORT =8080 \nDELAY = 200"), 0666)

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

	printState(env, envFilePath)
	err = os.WriteFile(envFilePath, []byte("TEST = example"), 0666)
	if err != nil {
		t.Fatalf("failed to write to file: %s", err)
	}

	printState(env, envFilePath)
	// This test is flaky on GitHub Actions, so we do this workaround
	counter := 0
	for counter < 200 && len(testEnv.env) != 1 {
		printState(env, envFilePath)
		time.Sleep(time.Millisecond * 50)
		counter++
	}

	test.AssertEqual(t, len(testEnv.env), 1)
	test.AssertEqual(t, readValue(t, testEnv.env, "TEST"), "example")
}

func printState(env *DotEnv, file string) {
	fmt.Println(env.env)
	content, _ := os.ReadFile(file)
	fmt.Println(string(content))
}
