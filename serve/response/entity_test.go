package response

import (
	"ngstaticserver/constants"
	"ngstaticserver/test"
	"os"
	"path/filepath"
	"testing"
)

const MainFile = "main.676ae13716545088.js"
const IndexFile = "index.html"

func TestGettingContent(t *testing.T) {
	context := test.NewTestDir(t)
	context.ImportTestNgsscApp()
	content := context.ReadFile(IndexFile)
	resolver := CreateEntityResolver(context.Path, constants.DefaultCacheSize)
	entity := resolver.Resolve(IndexFile)

	test.AssertEqual(t, string(entity.Content), content, "")
}

func TestGettingContentWithNoReadPermission(t *testing.T) {
	context := test.NewTestDir(t)
	context.ImportTestNgsscApp()
	err := os.Chmod(filepath.Join(context.Path, MainFile), 0000)
	if err != nil {
		panic(err)
	}
	resolver := CreateEntityResolver(context.Path, constants.DefaultCacheSize)
	entity := resolver.Resolve(MainFile)

	test.AssertTrue(t, entity.Content == nil, "")
	test.AssertTrue(t, entity.ContentBrotli == nil, "")
	test.AssertTrue(t, entity.ContentGzip == nil, "")
}

func TestGettingContentBrotli(t *testing.T) {
	context := test.NewTestDir(t)
	context.ImportTestNgsscApp()
	mainContent := context.ReadFile(MainFile)
	context.CompressFile(MainFile)
	resolver := CreateEntityResolver(context.Path, constants.DefaultCacheSize)
	entity := resolver.Resolve(MainFile)
	brotliContent := string(test.DecompressBrotli(entity.ContentBrotli))
	test.AssertEqual(t, brotliContent, mainContent, "")
}

func TestGettingContentGzip(t *testing.T) {
	context := test.NewTestDir(t)
	context.ImportTestNgsscApp()
	mainContent := context.ReadFile(MainFile)
	context.CompressFile(MainFile)
	resolver := CreateEntityResolver(context.Path, constants.DefaultCacheSize)
	entity := resolver.Resolve(MainFile)
	brotliContent := string(test.DecompressGzip(entity.ContentGzip))
	test.AssertEqual(t, brotliContent, mainContent, "")
}
