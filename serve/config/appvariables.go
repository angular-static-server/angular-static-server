package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/exp/slog"
)

type AppVariables struct {
	Variant                       string
	EnvironmentVariables          []string
	populatedEnvironmentVariables map[string]*string
}

// ngsscJSON corresponds to the relevant JSON structure of ngssc.json
type ngsscJSON struct {
	Variant              string
	EnvironmentVariables []string
}

var defaultAppVariables = &AppVariables{
	Variant:                       "global",
	EnvironmentVariables:          make([]string, 0),
	populatedEnvironmentVariables: make(map[string]*string),
}

func InitializeAppVariables(root string) *AppVariables {
	path := filepath.Join(root, "ngssc.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return defaultAppVariables
	}

	slog.Info(fmt.Sprintf("Detected ngssc.json file at %v. Reading configuration.", path))
	var ngssc *ngsscJSON
	err = json.Unmarshal(data, &ngssc)
	if err != nil {
		err = fmt.Errorf("failed to parse %v\n%v", path, err)
	} else if ngssc == nil {
		err = fmt.Errorf("invalid ngssc.json at %v (Must not be empty)", path)
	} else if ngssc.EnvironmentVariables == nil {
		err = fmt.Errorf("invalid ngssc.json at %v (environmentVariables must be defined)", path)
	} else if ngssc.Variant != "process" && ngssc.Variant != "global" && ngssc.Variant != "NG_ENV" {
		err = fmt.Errorf("invalid ngssc.json at %v (variant must either be process, NG_ENV or global)", path)
	}

	if err != nil {
		slog.Warn(fmt.Sprintf("%v, creating default configuration", err))
		return defaultAppVariables
	}

	return &AppVariables{
		Variant:                       ngssc.Variant,
		EnvironmentVariables:          ngssc.EnvironmentVariables,
		populatedEnvironmentVariables: populateEnvironmentVariables(ngssc.EnvironmentVariables),
	}
}

func populateEnvironmentVariables(environmentVariables []string) map[string]*string {
	envMap := make(map[string]*string)
	for _, env := range environmentVariables {
		value, exists := os.LookupEnv(env)
		if exists {
			envMap[env] = &value
		} else {
			envMap[env] = nil
		}
	}

	return envMap
}

// Insert the environment variables into the given content
func (ngsscConfig AppVariables) Insert(htmlBytes []byte, nonce string) string {
	if len(nonce) > 0 {
		nonce = fmt.Sprintf(" nonce=\"%v\"", nonce)
	}

	jsonBytes, _ := json.Marshal(ngsscConfig.populatedEnvironmentVariables)
	envMapJSON := string(jsonBytes)
	var iife string
	if ngsscConfig.Variant == "NG_ENV" {
		iife = fmt.Sprintf("self.NG_ENV=%v", envMapJSON)
	} else if ngsscConfig.Variant == "global" {
		iife = fmt.Sprintf("Object.assign(self,%v)", envMapJSON)
	} else {
		iife = fmt.Sprintf(`self.process={"env":%v}`, envMapJSON)
	}

	iifeScript := fmt.Sprintf("<script%v>(function(self){%v;})(window)</script>", nonce, iife)
	html := string(htmlBytes)
	configRegex := regexp.MustCompile(`<!--\s*CONFIG\s*-->`)
	if configRegex.Match(htmlBytes) {
		return configRegex.ReplaceAllString(html, iifeScript)
	} else if strings.Contains(html, "</title>") {
		return strings.Replace(html, "</title>", "</title>"+iifeScript, 1)
	} else {
		return strings.Replace(html, "</head>", iifeScript+"</head>", 1)
	}
}

func (appVariables *AppVariables) MergeVariables(variables map[string]*string) {
	if len(appVariables.EnvironmentVariables) > 0 {
		for k := range appVariables.populatedEnvironmentVariables {
			value, ok := variables[k]
			if ok {
				appVariables.populatedEnvironmentVariables[k] = value
			} else {
				value, ok := os.LookupEnv(k)
				if ok {
					appVariables.populatedEnvironmentVariables[k] = &value
				} else {
					appVariables.populatedEnvironmentVariables[k] = nil
				}
			}
		}
	} else {
		appVariables.populatedEnvironmentVariables = variables
	}
}

func (appVariables *AppVariables) IsEmpty() bool {
	return len(appVariables.populatedEnvironmentVariables) == 0
}

func (appVariables *AppVariables) Has(key string) bool {
	_, ok := appVariables.populatedEnvironmentVariables[key]
	return ok
}

func (appVariables *AppVariables) Update(key string, value string) {
	if appVariables.Has(key) {
		appVariables.populatedEnvironmentVariables[key] = &value
	}
}
