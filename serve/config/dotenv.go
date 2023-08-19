package config

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/hashicorp/go-envparse"
	"golang.org/x/exp/slog"
)

const globalDotEnvFile = "/config/.env"

type DotEnv struct {
	dir      string
	name     string
	env      map[string]*string
	onChange func(variables map[string]*string)
}

func CreateDotEnv(workingDirectory string, onChange func(variables map[string]*string)) *DotEnv {
	relativeEnv := filepath.Join(workingDirectory, "../config/.env")
	var env map[string]*string
	if _, err := os.Stat(globalDotEnvFile); err == nil {
		env = parseDotEnv(globalDotEnvFile)
	} else if _, err := os.Stat(relativeEnv); err == nil {
		env = parseDotEnv(relativeEnv)
	} else {
		env = parseDotEnv(filepath.Join(workingDirectory, ".env"))
	}

	instance := DotEnv{
		dir:      path.Dir(globalDotEnvFile),
		name:     path.Base(globalDotEnvFile),
		env:      env,
		onChange: onChange,
	}
	onChange(instance.env)
	return &instance
}

func (dotEnv *DotEnv) Dir() string {
	return dotEnv.dir
}

func (dotEnv *DotEnv) Name() string {
	return dotEnv.name
}

func (dotEnv *DotEnv) HandleChange() {
	dotEnv.env = parseDotEnv(filepath.Join(dotEnv.dir, dotEnv.name))
	dotEnv.onChange(dotEnv.env)
}

func parseDotEnv(filePath string) map[string]*string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return make(map[string]*string, 0)
	}

	slog.Info(fmt.Sprintf("Detected .env file at %v. Reading variables and adding watch.", filePath))
	env, err := envparse.Parse(bytes.NewReader(content))
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to parse dot env file at %v. Continuing without dot env file.", filePath))
		return make(map[string]*string, 0)
	} else if len(env) == 0 {
		return make(map[string]*string, 0)
	}

	result := make(map[string]*string, len(env))
	for k, v := range env {
		value := v
		result[k] = &value
	}

	return result
}
