package config

import (
	"ngstaticserver/test"
	"reflect"
	"testing"
)

func TestIsEmpty(t *testing.T) {
	appVariables := DefaultAppVariables()
	test.AssertTrue(t, appVariables.IsEmpty())
	test.AssertTrue(t, !appVariables.Has("LABEL"))
}

func TestReadingNgsscJson(t *testing.T) {
	context := test.NewTestDir(t)
	context.ImportTestApp("ngssc")
	appVariables := InitializeAppVariables(context.Path)
	test.AssertTrue(
		t,
		reflect.DeepEqual(appVariables.EnvironmentVariables, []string{"LABEL", "NGSS_CSP_NONCE"}))
	test.AssertTrue(t, appVariables.Has("LABEL"))
	test.AssertTrue(t, !appVariables.Has("LABEL2"))
}

func TestInsert(t *testing.T) {
	context := test.NewTestDir(t)
	context.ImportTestApp("ngssc")
	appVariables := InitializeAppVariables(context.Path)
	content, _ := appVariables.Insert([]byte("<!--CONFIG-->"), false)
	test.AssertEqual(t, string(content), "<script>(function(self){self.process={\"env\":{\"LABEL\":null,\"NGSS_CSP_NONCE\":null}};})(window)</script>")
}

func TestInsertProcess(t *testing.T) {
	context := test.NewTestDir(t)
	context.ImportTestApp("ngssc")
	appVariables := InitializeAppVariables(context.Path)
	appVariables.Variant = "global"
	content, _ := appVariables.Insert([]byte("</title>"), false)
	test.AssertEqual(t, string(content), "</title><script>(function(self){Object.assign(self,{\"LABEL\":null,\"NGSS_CSP_NONCE\":null});})(window)</script>")
}

func TestInsertNgEnv(t *testing.T) {
	context := test.NewTestDir(t)
	context.ImportTestApp("ngssc")
	appVariables := InitializeAppVariables(context.Path)
	appVariables.Variant = "NG_ENV"
	content, _ := appVariables.Insert([]byte("</head>"), false)
	test.AssertEqual(t, string(content), "<script>(function(self){self.NG_ENV={\"LABEL\":null,\"NGSS_CSP_NONCE\":null};})(window)</script></head>")
}

func TestUpdate(t *testing.T) {
	context := test.NewTestDir(t)
	context.ImportTestApp("ngssc")
	appVariables := InitializeAppVariables(context.Path)
	content, _ := appVariables.Insert([]byte("<!--CONFIG-->"), false)
	test.AssertEqual(t, string(content), "<script>(function(self){self.process={\"env\":{\"LABEL\":null,\"NGSS_CSP_NONCE\":null}};})(window)</script>")
	appVariables.Update("LABEL", "label")
	content, _ = appVariables.Insert([]byte("<!--CONFIG-->"), false)
	test.AssertEqual(t, string(content), "<script>(function(self){self.process={\"env\":{\"LABEL\":\"label\",\"NGSS_CSP_NONCE\":null}};})(window)</script>")
}
