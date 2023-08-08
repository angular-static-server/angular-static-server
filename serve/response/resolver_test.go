package response

import (
	"ngstaticserver/constants"
	"ngstaticserver/test"
	"path/filepath"
	"testing"
)

func TestResolvingNotFound(t *testing.T) {
	context := test.NewTestDir(t)
	resolver := CreateEntityResolver(context.Path, constants.DefaultCacheSize)
	entity := resolver.Resolve("does-not-exist.txt")
	test.AssertTrue(t, entity.IsNotFound(), "")
}

func TestResolvingNotFoundIndex(t *testing.T) {
	context := test.NewTestDir(t)
	context.CreateDirectory("de").CreateFile("index.html", "content")
	context.CreateDirectory("en").CreateFile("index.html", "content")
	resolver := CreateEntityResolver(context.Path, constants.DefaultCacheSize)
	entity := resolver.Resolve("does-not-exist.txt")
	test.AssertTrue(t, entity.IsNotFound(), "")
}

func TestMatchingLanguageEmpty(t *testing.T) {
	context := test.NewTestDir(t)
	resolver := CreateEntityResolver(context.Path, constants.DefaultCacheSize)
	test.AssertEqual(t, resolver.MatchLanguage([]string{}), "", "")
	test.AssertEqual(t, resolver.MatchLanguage([]string{"en"}), "", "")
	test.AssertEqual(t, resolver.MatchLanguage([]string{"fr", "en"}), "", "")
}

func TestMatchingLanguage(t *testing.T) {
	context := test.NewTestDir(t)
	context.CreateDirectory("de-CH").CreateFile("index.html", "content")
	context.CreateDirectory("en").CreateFile("index.html", "content")
	resolver := CreateEntityResolver(context.Path, constants.DefaultCacheSize)
	test.AssertEqual(t, resolver.MatchLanguage([]string{}), "de-CH", "")
	test.AssertEqual(t, resolver.MatchLanguage([]string{"en"}), "en", "")
	test.AssertEqual(t, resolver.MatchLanguage([]string{"fr", "en"}), "en", "")
	test.AssertEqual(t, resolver.MatchLanguage([]string{"de"}), "de-CH", "")
	test.AssertEqual(t, resolver.MatchLanguage([]string{"en-US"}), "en", "")
}

func TestResolvingExistingFile(t *testing.T) {
	context := test.NewTestDir(t)
	context.CreateFile("exist.txt", "example")
	resolver := CreateEntityResolver(context.Path, constants.DefaultCacheSize)
	entity := resolver.Resolve("exist.txt")
	test.AssertEqual(t, entity.ContentType, "text/plain; charset=utf-8", "")
	test.AssertEqual(t, entity.Path, filepath.Join(context.Path, "exist.txt"), "")
	test.AssertTrue(t, !entity.IsIndex(), "")
	test.AssertTrue(t, !entity.IsFingerprinted(), "")
	test.AssertTrue(t, !entity.IsNotFound(), "")
	test.AssertTrue(t, !entity.Compressed, "")
}

func TestResolvingExistingFileMetaInfo(t *testing.T) {
	context := test.NewTestDir(t)
	context.CreateFile("exist.txt", "example")
	resolver := CreateEntityResolver(context.Path, constants.DefaultCacheSize)
	entity := resolver.Resolve("exist.txt")
	size, modTime, contentType := fileMeta(filepath.Join(context.Path, "exist.txt"))
	test.AssertEqual(t, entity.Size, size, "")
	test.AssertEqual(t, entity.ModTime, modTime, "")
	test.AssertEqual(t, entity.ContentType, contentType, "")
}

func TestResolvingExistingFileWithBrotliAndGzip(t *testing.T) {
	context := test.NewTestDir(t)
	context.CreateFile("exist.txt", "example")
	context.CreateFile("exist.txt.br", "noop")
	context.CreateFile("exist.txt.gz", "noop")
	resolver := CreateEntityResolver(context.Path, constants.DefaultCacheSize)
	entity := resolver.Resolve("exist.txt")
	test.AssertTrue(t, !entity.IsIndex(), "")
	test.AssertTrue(t, !entity.IsFingerprinted(), "")
	test.AssertTrue(t, !entity.IsNotFound(), "")
	test.AssertTrue(t, entity.Compressed, "")
}

func TestResolvingExistingFingerprintedFile(t *testing.T) {
	context := test.NewTestDir(t)
	context.CreateFile("main.676ae13716545088.js", "example")
	resolver := CreateEntityResolver(context.Path, constants.DefaultCacheSize)
	entity := resolver.Resolve("main.676ae13716545088.js")
	test.AssertEqual(t, entity.ContentType, "text/javascript; charset=utf-8", "")
	test.AssertTrue(t, entity.IsFingerprinted(), "")
	test.AssertTrue(t, !entity.Compressed, "")
}

func TestResolvingExistingFingerprintedFileWithBrotliAndGzip(t *testing.T) {
	context := test.NewTestDir(t)
	context.CreateFile("main.676ae13716545088.js", "example")
	context.CreateFile("main.676ae13716545088.js.br", "noop")
	context.CreateFile("main.676ae13716545088.js.gz", "noop")
	resolver := CreateEntityResolver(context.Path, constants.DefaultCacheSize)
	entity := resolver.Resolve("main.676ae13716545088.js")
	test.AssertTrue(t, entity.IsFingerprinted(), "")
	test.AssertTrue(t, entity.Compressed, "")
}

func TestResolvingIndex(t *testing.T) {
	context := test.NewTestDir(t)
	context.CreateFile("index.html", "example")
	resolver := CreateEntityResolver(context.Path, constants.DefaultCacheSize)
	entity := resolver.Resolve("")
	test.AssertEqual(t, entity.Path, filepath.Join(context.Path, "index.html"), "")
	test.AssertTrue(t, entity.IsIndex(), "")
	test.AssertTrue(t, entity.Compressed, "")
}

func TestResolvingIndexWithSubpath(t *testing.T) {
	context := test.NewTestDir(t)
	context.CreateFile("index.html", "example")
	resolver := CreateEntityResolver(context.Path, constants.DefaultCacheSize)
	entity := resolver.Resolve("some/path")
	test.AssertEqual(t, entity.Path, filepath.Join(context.Path, "index.html"), "")
	test.AssertTrue(t, entity.IsIndex(), "")
	test.AssertTrue(t, entity.Compressed, "")

	entity = resolver.Resolve("some/path")
	test.AssertEqual(t, entity.Path, filepath.Join(context.Path, "index.html"), "")
}

func TestResolvingMultipleIndex(t *testing.T) {
	context := test.NewTestDir(t)
	context.CreateFile("index.html", "example")
	context.CreateDirectory("nested").CreateFile("index.html", "example2")
	resolver := CreateEntityResolver(context.Path, constants.DefaultCacheSize)
	entity := resolver.Resolve("nested/some/path")
	test.AssertEqual(t, entity.Path, filepath.Join(context.Path, "nested/index.html"), "")
	test.AssertTrue(t, entity.IsIndex(), "")
	test.AssertTrue(t, entity.Compressed, "")
}

func TestResolvingExistingBinaryFile(t *testing.T) {
	context := test.NewTestDir(t)
	context.CreateFile("exist.exe", "noop")
	resolver := CreateEntityResolver(context.Path, constants.DefaultCacheSize)
	entity := resolver.Resolve("exist.exe")
	test.AssertTrue(t, !entity.IsIndex(), "")
	test.AssertTrue(t, !entity.IsFingerprinted(), "")
	test.AssertTrue(t, !entity.IsNotFound(), "")
	test.AssertTrue(t, !entity.Compressed, "")
}

func TestResolvingExistingFileSize(t *testing.T) {
	context := test.NewTestDir(t)
	content := "noop"
	context.CreateFile("exist.txt", content)
	resolver := CreateEntityResolver(context.Path, constants.DefaultCacheSize)
	entity := resolver.Resolve("exist.txt")
	test.AssertTrue(t, entity.Size > int64(len(content)-1), "")
	test.AssertTrue(t, !(entity.Size > int64(len(content))), "")
}
