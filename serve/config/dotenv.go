package config

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"

	"github.com/hashicorp/go-envparse"
)

type DotEnv struct {
	dir      string
	name     string
	env      map[string]*string
	onChange func(variables map[string]*string)
}

func CreateDotEnv(workingDirectory string, onChange func(variables map[string]*string)) *DotEnv {
	configEnvPath := filepath.Join(workingDirectory, "../config/.env")
	var env map[string]*string
	if _, err := os.Stat(configEnvPath); err == nil {
		slog.Info(fmt.Sprintf("Detected .env file at %v. Reading variables and adding watch.", configEnvPath))
		env = parseDotEnv(configEnvPath)
	} else {
		slog.Info(fmt.Sprintf("Detected .env file at %v. Reading variables.", configEnvPath))
		env = parseDotEnv(filepath.Join(workingDirectory, ".env"))
	}

	instance := DotEnv{
		dir:      path.Dir(configEnvPath),
		name:     path.Base(configEnvPath),
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
	filePath := filepath.Join(dotEnv.dir, dotEnv.name)
	slog.Info(fmt.Sprintf("Detected change in %v. Reading variables.", filePath))
	dotEnv.env = parseDotEnv(filePath)
	dotEnv.onChange(dotEnv.env)
}

func parseDotEnv(filePath string) map[string]*string {
	content, err := os.ReadFile(filePath)
	fmt.Printf("parse: %v", string(content))
	if err != nil {
		return make(map[string]*string, 0)
	}

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
