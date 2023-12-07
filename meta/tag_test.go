package meta

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

var benchmarkTag Tag = `default:BAR;arg:FOO;nonsense:CANTFINDTHIS!;flag:BAZ;`

func TestTagArg(t *testing.T) {
	type test struct {
		tag    Tag
		want   *ArgSpec
		wantOK bool
	}

	for tn, tc := range map[string]test{
		"empty":                            {},
		"plain index":                      {"arg:42", &ArgSpec{42, 0}, true},
		"slice index":                      {"arg:[42]", &ArgSpec{42, 0}, true},
		"range between":                    {"arg:[2:4]", &ArgSpec{2, 4}, true},
		"range between from explicit zero": {"arg:[0:4]", &ArgSpec{0, 4}, true},
		"range consume all to":             {"arg:[:4]", &ArgSpec{0, 4}, true},
		"range consume all from":           {"arg:[2:]", &ArgSpec{2, -1}, true},
		"range consume all":                {"arg:[:]", &ArgSpec{0, -1}, true},
		"explicitly unset not ok":          {"arg:", nil, false},
		"range same not ok":                {"arg:[2:2]", nil, false},
		"range end before start not ok":    {"arg:[4:2]", nil, false},
		"range to zero not ok":             {"arg:[2:0]", nil, false},
		"range malformed not ok":           {"arg:[2:a]", nil, false},
		"slice index empty not ok":         {"arg:[]", nil, false},
		"malformed not ok":                 {"arg:I thrive in chaos.", nil, false},
	} {
		t.Run(tn, func(t *testing.T) {
			got, ok := tc.tag.Arg()
			if ok != tc.wantOK {
				t.Errorf("Arg(): ok mismatch: got: %v want: %v", got, tc.wantOK)
			}
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("Arg(): mismatch (-got,+want):\n%v", diff)
			}
		})
	}
}

func BenchmarkTagArg(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = benchmarkTag.Arg()
	}
}

func TestTagDecompose(t *testing.T) {
	type values [3]string
	type test struct {
		tag  Tag
		want values
	}

	for tn, tc := range map[string]test{
		"implicitly empty":   {},
		"explicitly empty":   {"arg:;default:;flag:", values{}},
		"all":                {"arg:FOO;default:BAR;flag:BAZ", values{"FOO", "BAR", "BAZ"}},
		"all shifted order":  {"default:BAR;flag:BAZ;arg:FOO", values{"FOO", "BAR", "BAZ"}},
		"whitespace trimmed": {"arg: FOO ; default: BAR ; flag: BAZ ;", values{"FOO", "BAR", "BAZ"}},
		"extra ignored":      {"arg:FOO;default:BAR;nonsense:CANTFINDTHIS!;flag:BAZ", values{"FOO", "BAR", "BAZ"}},
	} {
		t.Run(tn, func(t *testing.T) {
			arg, def, flag := tc.tag.decompose()
			got := values{arg, def, flag}
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("decompose(): mismatch (-got,+want):\n%v", diff)
			}
		})
	}
}

func BenchmarkTagDecompose(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _ = benchmarkTag.decompose()
	}
}

func TestTagDefault(t *testing.T) {
	type test struct {
		tag    Tag
		want   string
		wantOK bool
	}

	for tn, tc := range map[string]test{
		"empty":                   {},
		"value":                   {"default:42", "42", true},
		"value quotes preserved":  {`default:"foo"`, `"foo"`, true},
		"explicitly unset not ok": {`default:`, "", false},
	} {
		t.Run(tn, func(t *testing.T) {
			got, ok := tc.tag.Default()
			if ok != tc.wantOK {
				t.Errorf("Default(): ok mismatch: got: %v want: %v", got, tc.wantOK)
			}
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("Default(): mismatch (-got,+want):\n%v", diff)
			}
		})
	}
}

func BenchmarkTagDefault(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = benchmarkTag.Default()
	}
}

func TestTagFlag(t *testing.T) {
	type test struct {
		tag    Tag
		want   *FlagSpec
		wantOK bool
	}

	for tn, tc := range map[string]test{
		"empty":                   {},
		"go style":                {"flag:foo", &FlagSpec{"foo", ""}, true},
		"posix style":             {"flag:foo,F", &FlagSpec{"foo", "F"}, true},
		"two short flags not ok":  {"flag:f,b", nil, false},
		"two long flags not ok":   {"flag:foo,bar", nil, false},
		"explicitly unset not ok": {`default:`, nil, false},
	} {
		t.Run(tn, func(t *testing.T) {
			got, ok := tc.tag.Flag()
			if ok != tc.wantOK {
				t.Errorf("Flag(): ok mismatch: got: %v want: %v", got, tc.wantOK)
			}
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("Flag(): mismatch (-got,+want):\n%v", diff)
			}
		})
	}
}

func BenchmarkTagFlag(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = benchmarkTag.Flag()
	}
}
