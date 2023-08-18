package endpoints

import (
	"ngstaticserver/constants"
	"ngstaticserver/serve/config"
	"ngstaticserver/serve/headers"
	"ngstaticserver/test"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

var CspTemplate = regexp.MustCompile("\\$\\{_CSP_[^_]+_SRC\\}").ReplaceAllString(constants.CspTemplate, "")

func TestVersionEndpoint_noVersionFile(t *testing.T) {
	dir := t.TempDir()
	endpoint := VersionEndpoint(filepath.Join(dir, "version.json"))
	_, isInlineString := endpoint.(InlineStringEndpoint)
	test.AssertTrue(t, isInlineString)
}

func TestVersionEndpoint_withVersionFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "version.json"), []byte("{}"), 0644)
	endpoint := VersionEndpoint(filepath.Join(dir, "version.json"))
	_, isUncompressedFileEndpoint := endpoint.(UncompressedFileEndpoint)
	test.AssertTrue(t, isUncompressedFileEndpoint)
}

func TestHeartbeatEndpoint(t *testing.T) {
	endpoint := HeartbeatEndpoint()
	_, isInlineString := endpoint.(InlineStringEndpoint)
	test.AssertTrue(t, isInlineString)
}

func TestFileEndpoint_uncompressed(t *testing.T) {
	context := test.NewTestDir(t)
	context.WriteFile(File, strings.Repeat("example", 10))
	endpoint, err := ResolveFileEndpoint(filepath.Join(context.Path, File), 0)
	test.AssertNoError(t, err)
	_, isType := endpoint.(UncompressedFileEndpoint)
	test.AssertTrue(t, isType)
}

func TestFileEndpoint_uncompressed_fingerprinted(t *testing.T) {
	context := test.NewTestDir(t)
	context.WriteFile("main.458f86595498b767.js", strings.Repeat("example", 10))
	endpoint, err := ResolveFileEndpoint(filepath.Join(context.Path, "main.458f86595498b767.js"), 0)
	test.AssertNoError(t, err)
	_, isType := endpoint.(UncompressedFileEndpoint)
	test.AssertTrue(t, isType)
}

func TestFileEndpoint_brotli_gzip(t *testing.T) {
	context := test.NewTestDir(t)
	context.WriteFile(File, strings.Repeat("example", 10))
	context.CompressFile(File)
	endpoint, err := ResolveFileEndpoint(filepath.Join(context.Path, File), 0)
	test.AssertNoError(t, err)
	_, isType := endpoint.(BrotliGzipFileEndpoint)
	test.AssertTrue(t, isType)
}

func TestFileEndpoint_brotli(t *testing.T) {
	context := test.NewTestDir(t)
	context.WriteFile(File, strings.Repeat("example", 10))
	context.CompressFile(File)
	context.RemoveFile(File + ".gz")
	endpoint, err := ResolveFileEndpoint(filepath.Join(context.Path, File), 0)
	test.AssertNoError(t, err)
	_, isType := endpoint.(BrotliFileEndpoint)
	test.AssertTrue(t, isType)
}

func TestFileEndpoint_gzip(t *testing.T) {
	context := test.NewTestDir(t)
	context.WriteFile(File, strings.Repeat("example", 10))
	context.CompressFile(File)
	context.RemoveFile(File + ".br")
	endpoint, err := ResolveFileEndpoint(filepath.Join(context.Path, File), 0)
	test.AssertNoError(t, err)
	_, isType := endpoint.(GzipFileEndpoint)
	test.AssertTrue(t, isType)
}

func TestIndexEndpoint_noCsp(t *testing.T) {
	context := test.NewTestDir(t)
	context.WriteFile("index.html", strings.Repeat("example", 10))
	endpoint := ResolveIndexEndpoint(
		filepath.Join(context.Path, "index.html"),
		0,
		"",
		config.DefaultAppVariables())
	indexEndpoint, isType := endpoint.(IndexEndpoint)
	test.AssertTrue(t, isType)
	test.AssertEqual(t, indexEndpoint.PreCompression, headers.Encoding(headers.NO_COMPRESSION))
}

func TestIndexEndpoint_emptyCsp(t *testing.T) {
	context := test.NewTestDir(t)
	context.WriteFile("index.html", strings.Repeat("example", 10)+"${NGSS_CSP_NONCE}")
	context.CompressFile("index.html")
	endpoint := ResolveIndexEndpoint(
		filepath.Join(context.Path, "index.html"),
		0,
		"",
		config.DefaultAppVariables())
	indexEndpoint, isType := endpoint.(IndexEndpoint)
	test.AssertTrue(t, isType)
	test.AssertEqual(t, indexEndpoint.PreCompression, headers.Encoding(headers.BROTLI|headers.GZIP))
}

func TestIndexEndpoint_withCsp(t *testing.T) {
	context := test.NewTestDir(t)
	context.ImportTestApp("minimal")
	endpoint := ResolveIndexEndpoint(
		filepath.Join(context.Path, "index.html"),
		0,
		CspTemplate,
		config.DefaultAppVariables())
	indexEndpoint, isType := endpoint.(CspIndexEndpoint)
	test.AssertTrue(t, isType)
	test.AssertEqual(t, indexEndpoint.Csp, "default-src 'self' ; connect-src 'self' ; font-src 'self' ; img-src 'self' ; script-src 'self' ${NGSS_CSP_NONCE} ${NGSS_CSP_SCRIPT_HASH} ; style-src 'self' ${NGSS_CSP_NONCE}  ;")
}

func TestIndexEndpoint_withCsp_customHtml(t *testing.T) {
	context := test.NewTestDir(t)
	context.ImportTestApp("minimal")
	indexHtml := context.ReadFile("index.html")
	indexHtml = strings.Replace(indexHtml, " nonce=\"${NGSS_CSP_NONCE}\"", "", -1)
	context.WriteFile("index.html", indexHtml)
	endpoint := ResolveIndexEndpoint(
		filepath.Join(context.Path, "index.html"),
		0,
		CspTemplate,
		config.DefaultAppVariables())
	indexEndpoint, isType := endpoint.(CspIndexEndpoint)
	test.AssertTrue(t, isType)
	test.AssertEqual(
		t,
		indexEndpoint.Csp,
		"default-src 'self' ; connect-src 'self' ; font-src 'self' ; img-src 'self' ; script-src 'self' ${NGSS_CSP_NONCE} ${NGSS_CSP_SCRIPT_HASH} 'sha512-21c5f0b047d843c7b26d6e1a508fef78d2056ae72e7bc392601abb7e87f0a89609846fa3945ed1d424e252f77af4647942ae1351fbe72e791a0b41bdde45df69' ; style-src 'self' ${NGSS_CSP_NONCE} 'sha512-c98ccde2b855cec4911a237a882eec81ac48dd8f911c94f85179eb1340e356b9870b276051336963daf7f8cf9c208abc5de2b2a2cc0afb43da4a7ae1f0b31c1a' 'sha512-f49c1282a60bb2a14bdda5e4eae292fa2bd04c171c4394eaeacfec5190f7bae7b36ad21f2e8ee86c87af9d504e3ca997eb80d7444294704ed55229641ae3c6bb' 'sha512-65317c8bd2ccab71d02144c1da0bd489a58826b27fc7e35911b6d28fff4e088904e4ec805cab1e105fdca2e16988df79f167be92bc585903efddcb3ac43b4de1' ;",
	)
}

func TestRootEndpoint(t *testing.T) {
	context := test.NewTestDir(t)
	context.ImportTestApp("i18n")
	endpoint := ResolveRootEndpoint(context.Path, "")
	rootEndpoint, isType := endpoint.(RootEndpoint)
	test.AssertTrue(t, isType)
	expectedLanguages := []string{"de-CH", "en-US", "fr"}
	test.AssertTrue(t, reflect.DeepEqual(rootEndpoint.AvailablePaths, expectedLanguages))
}

func TestRootEndpoint_withEmptyDefault(t *testing.T) {
	context := test.NewTestDir(t)
	context.ImportTestApp("i18n")
	endpoint := ResolveRootEndpoint(context.Path, "")
	rootEndpoint, isType := endpoint.(RootEndpoint)
	test.AssertTrue(t, isType)
	test.AssertEqual(t, rootEndpoint.DefaultPath, "de-CH")
}

func TestRootEndpoint_withMissingDefault(t *testing.T) {
	context := test.NewTestDir(t)
	context.ImportTestApp("i18n")
	endpoint := ResolveRootEndpoint(context.Path, "lol")
	rootEndpoint, isType := endpoint.(RootEndpoint)
	test.AssertTrue(t, isType)
	test.AssertEqual(t, rootEndpoint.DefaultPath, "de-CH")
}

func TestRootEndpoint_withDefault(t *testing.T) {
	context := test.NewTestDir(t)
	context.ImportTestApp("i18n")
	endpoint := ResolveRootEndpoint(context.Path, "en-US")
	rootEndpoint, isType := endpoint.(RootEndpoint)
	test.AssertTrue(t, isType)
	test.AssertEqual(t, rootEndpoint.DefaultPath, "en-US")
}
