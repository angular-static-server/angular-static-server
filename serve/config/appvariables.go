package config

import (
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type AppVariables struct {
	Variant                       string
	EnvironmentVariables          []string
	LastChangedAt                 time.Time
	populatedEnvironmentVariables map[string]*string
}

// ngsscJSON corresponds to the relevant JSON structure of ngssc.json
type ngsscJSON struct {
	Variant              string
	EnvironmentVariables []string
}

func DefaultAppVariables() *AppVariables {
	return &AppVariables{
		Variant:                       "global",
		EnvironmentVariables:          make([]string, 0),
		LastChangedAt:                 time.Now(),
		populatedEnvironmentVariables: make(map[string]*string),
	}
}

func InitializeAppVariables(root string) *AppVariables {
	path := filepath.Join(root, "ngssc.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultAppVariables()
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
		return DefaultAppVariables()
	}

	return &AppVariables{
		Variant:                       ngssc.Variant,
		EnvironmentVariables:          ngssc.EnvironmentVariables,
		LastChangedAt:                 time.Now(),
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

func (ngsscConfig AppVariables) Insert(htmlBytes []byte, calculateCspHash bool) ([]byte, string) {
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
	var cspHash string
	iifeContent := fmt.Sprintf("(function(self){%v;})(window)", iife)
	if calculateCspHash {
		cspHash = fmt.Sprintf("'sha512-%x'", sha512.Sum512([]byte(iifeContent)))
	}

	iifeScript := fmt.Sprintf("<script>%v</script>", iifeContent)
	html := string(htmlBytes)
	configRegex := regexp.MustCompile(`<!--\s*CONFIG\s*-->`)
	if configRegex.Match(htmlBytes) {
		html = configRegex.ReplaceAllString(html, iifeScript)
	} else if strings.Contains(html, "</title>") {
		html = strings.Replace(html, "</title>", "</title>"+iifeScript, 1)
	} else {
		html = strings.Replace(html, "</head>", iifeScript+"</head>", 1)
	}

	return []byte(html), cspHash
}

func (appVariables *AppVariables) MergeVariables(variables map[string]*string) {
	appVariables.LastChangedAt = time.Now()
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
