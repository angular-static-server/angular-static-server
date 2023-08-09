package response

import (
	"ngstaticserver/constants"
	"ngstaticserver/test"
	"os"
	"path/filepath"
	"testing"
)

const IndexFile = "index.html"

func TestGettingContent(t *testing.T) {
	context := test.NewTestDir(t)
	context.ImportTestApp("minimal")
	content := context.ReadFile(IndexFile)
	resolver := CreateEntityResolver(context.Path, constants.DefaultCacheSize)
	entity := resolver.Resolve(IndexFile)

	test.AssertEqual(t, string(entity.Content), content, "")
}

func TestGettingContentWithNoReadPermission(t *testing.T) {
	context := test.NewTestDir(t)
	context.ImportTestApp("minimal")
	mainFile := context.FindFile("main.")
	err := os.Chmod(filepath.Join(context.Path, mainFile), 0000)
	if err != nil {
		panic(err)
	}
	resolver := CreateEntityResolver(context.Path, constants.DefaultCacheSize)
	entity := resolver.Resolve(mainFile)

	test.AssertTrue(t, entity.Content == nil, "")
	test.AssertTrue(t, entity.ContentBrotli == nil, "")
	test.AssertTrue(t, entity.ContentGzip == nil, "")
}

func TestGettingContentBrotli(t *testing.T) {
	context := test.NewTestDir(t)
	context.ImportTestApp("minimal")
	mainFile := context.FindFile("main.")
	mainContent := context.ReadFile(mainFile)
	context.CompressFile(mainFile)
	resolver := CreateEntityResolver(context.Path, constants.DefaultCacheSize)
	entity := resolver.Resolve(mainFile)
	brotliContent := string(test.DecompressBrotli(entity.ContentBrotli))
	test.AssertEqual(t, brotliContent, mainContent, "")
}

func TestGettingContentGzip(t *testing.T) {
	context := test.NewTestDir(t)
	context.ImportTestApp("minimal")
	mainFile := context.FindFile("main.")
	mainContent := context.ReadFile(mainFile)
	context.CompressFile(mainFile)
	resolver := CreateEntityResolver(context.Path, constants.DefaultCacheSize)
	entity := resolver.Resolve(mainFile)
	brotliContent := string(test.DecompressGzip(entity.ContentGzip))
	test.AssertEqual(t, brotliContent, mainContent, "")
}
