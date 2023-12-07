package meta

import (
	"embed"
	"io/fs"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

//go:embed testdata
var testdataFS embed.FS

type testFile struct {
	Path string
	File fs.File

	tb testing.TB
}

func (f *testFile) Name() string {
	return f.Path
}

func (f *testFile) Read(b []byte) (int, error) {
	f.tb.Helper()
	return f.File.Read(b)
}

func file(tb testing.TB, path string) *testFile {
	tb.Helper()
	f, err := testdataFS.Open(path)
	if err != nil {
		tb.Fatal(err)
	}
	return &testFile{
		Path: path,
		File: f,
		tb:   tb,
	}
}

func TestFromFile(t *testing.T) {
	type test struct {
		path string
		typ  string
		want *Command
	}

	for _, tc := range []test{
		{
			"testdata/simple/simple.go", "Tester", &Command{
				Name:        "simple",
				Package:     "simple",
				Type:        "Tester",
				Help:        "simple is a simple test for cliche. It contains a single Command with no tags.",
				Description: "Tester is a cliche command which exercises default inputs.",
			},
		},
	} {
		t.Run(tc.path, func(t *testing.T) {
			got := FromFile(file(t, tc.path), tc.typ)
			if diff := cmp.Diff(got, tc.want, cmpopts.IgnoreUnexported(Command{})); diff != "" {
				t.Errorf("FromFile(): mismatch(-got,+want):\n%v", diff)
			}
		})
	}
}
