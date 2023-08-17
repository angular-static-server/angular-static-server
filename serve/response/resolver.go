package response

import (
	"fmt"
	"mime"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"golang.org/x/exp/slog"
)

type IndexResolver func(filePath string) (indexPath *string, isInRequestedDir bool)

type EntityResolver struct {
	root          string
	indexPaths    []string
	indexResolver IndexResolver
	cache         *lru.TwoQueueCache[string, ResponseEntity]
}

var fingerprintRegex, _ = regexp.Compile("\\.[a-zA-Z0-9]{16,}\\.(js|mjs|css)$")

func CreateEntityResolver(root string, cacheBuffer int) EntityResolver {
	cache, _ := lru.New2Q[string, ResponseEntity](cacheBuffer)
	indexPaths := findIndexPaths(root)
	return EntityResolver{
		root:          root,
		indexPaths:    indexPaths,
		indexResolver: createIndexResolver(root, indexPaths),
		cache:         cache,
	}
}

func findIndexPaths(root string) []string {
	indexPaths := make([]string, 0)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err == nil && strings.HasSuffix(path, "/index.html") {
			indexPath, _ := filepath.Rel(root, filepath.Dir(path))
			indexPaths = append(indexPaths, indexPath)
		}
		return err
	})
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to look up index.html files in %v: %v", root, err))
	}

	return indexPaths
}

func createIndexResolver(root string, indexPaths []string) IndexResolver {
	if len(indexPaths) == 1 && indexPaths[0] == "." {
		indexHtmlPath := filepath.Join(root, "index.html")
		return func(filePath string) (*string, bool) {
			return &indexHtmlPath, isSameIndexPath(indexHtmlPath, filePath)
		}
	} else if len(indexPaths) >= 1 {
		return func(filePath string) (*string, bool) {
			resolvedPath := filePath
			for {
				indexPath := path.Join(resolvedPath, "index.html")
				if fileExists(indexPath) {
					return &indexPath, isSameIndexPath(indexPath, filePath)
				} else if rel, _ := filepath.Rel(root, resolvedPath); rel == "." {
					break
				}

				resolvedPath = path.Dir(resolvedPath)
			}

			return nil, false
		}
	} else {
		return func(filePath string) (*string, bool) {
			return nil, false
		}
	}
}

func isSameIndexPath(indexPath, requestPath string) bool {
	rel, _ := filepath.Rel(filepath.Dir(indexPath), requestPath)
	return rel == "."
}

func (resolver EntityResolver) Resolve(filePath string) ResponseEntity {
	resolvedPath := path.Join(resolver.root, filePath)
	entity, ok := resolver.cache.Get(resolvedPath)
	if ok && entity.IsIndexProxy() {
		entity, ok = resolver.cache.Get(filepath.Dir(entity.Path))
	}

	if ok {
		return entity
	} else if filePath == "/__version__" {
		versionFilePath := filepath.Join(resolver.root, "version.json")
		_, modTime, contentType := fileMeta(versionFilePath)
		content := readFile(resolvedPath)
		if content == nil {
			content = []byte("{\n  \"undefined\": \"app does not have a version.json file\"\n}")
		}
		entity = ResponseEntity{
			Path:        versionFilePath,
			fileType:    VERSION,
			Size:        int64(len(content)),
			ModTime:     modTime,
			ContentType: contentType,
			Compressed:  false,
			Content:     content,
		}
		resolver.cache.Add(resolvedPath, entity)
	} else if fileExists(resolvedPath) {
		var category FileType
		category = FILE
		if fingerprintRegex.MatchString(filePath) {
			category = FINGERPRINTED_FILE
		}

		fileSize, modTime, contentType := fileMeta(resolvedPath)
		contentBrotli := readFileDebugLogOnError(resolvedPath + ".br")
		contentGzip := readFileDebugLogOnError(resolvedPath + ".gz")
		entity = ResponseEntity{
			Path:          resolvedPath,
			fileType:      category,
			Size:          fileSize,
			ModTime:       modTime,
			ContentType:   contentType,
			Compressed:    contentBrotli != nil && contentGzip != nil,
			Content:       readFile(resolvedPath),
			ContentBrotli: contentBrotli,
			ContentGzip:   contentGzip,
		}
		resolver.cache.Add(resolvedPath, entity)
	} else if indexPath, isInRequestedDir := resolver.indexResolver(resolvedPath); indexPath != nil {
		if !isInRequestedDir {
			entity, ok = resolver.cache.Get(filepath.Dir(*indexPath))
			if ok {
				resolver.cache.Add(resolvedPath, ResponseEntity{
					Path:     *indexPath,
					fileType: INDEX_PROXY,
				})
				return entity
			}
		}

		fileSize, modTime, contentType := fileMeta(*indexPath)
		entity = ResponseEntity{
			Path:          *indexPath,
			fileType:      INDEX,
			Size:          fileSize,
			ModTime:       modTime,
			ContentType:   contentType,
			Compressed:    true,
			Content:       readFile(*indexPath),
			ContentBrotli: readFileDebugLogOnError(*indexPath + ".br"),
			ContentGzip:   readFileDebugLogOnError(*indexPath + ".gz"),
		}
		if isInRequestedDir {
			resolver.cache.Add(resolvedPath, entity)
		} else {
			resolver.cache.Add(resolvedPath, ResponseEntity{
				Path:     *indexPath,
				fileType: INDEX_PROXY,
			})
			resolver.cache.Add(filepath.Dir(*indexPath), entity)
		}
	} else {
		entity = ResponseEntity{fileType: NOT_FOUND}
		resolver.cache.Add(resolvedPath, entity)
	}

	return entity
}

func (resolver EntityResolver) MatchLanguage(languages []string) string {
	if len(resolver.indexPaths) == 0 {
		return ""
	}

	for _, l := range languages {
		for _, a := range resolver.indexPaths {
			if l == a {
				return l
			}
		}
	}

	for _, l := range languages {
		for _, a := range resolver.indexPaths {
			if strings.HasPrefix(a, l) || strings.HasPrefix(l, a) {
				return a
			}
		}
	}

	return resolver.indexPaths[0]
}

func fileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	return err == nil && !info.IsDir()
}

func fileMeta(filePath string) (size int64, modTime time.Time, contentType string) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, time.Time{}, mime.TypeByExtension(filepath.Ext(filePath))
	}
	return info.Size(), info.ModTime(), mime.TypeByExtension(filepath.Ext(filePath))
}

func readFile(filePath string) []byte {
	content, err := os.ReadFile(filePath)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to read %v", filePath))
	}
	return content
}

func readFileDebugLogOnError(filePath string) []byte {
	content, err := os.ReadFile(filePath)
	if err != nil {
		slog.Debug(fmt.Sprintf("Failed to read %v", filePath))
	}
	return content
}
