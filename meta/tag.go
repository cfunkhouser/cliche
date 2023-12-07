package meta

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ArgSpec describes parsed positional arguments as defined in a facile struct
// tag.
type ArgSpec struct {
	// Start index, inclusive.
	Start int
	// End index, exclusive. A negative value indicates that all remaining
	// arguments beginning with Start should be consumed.
	End int
}

// String representation of the arg spec. Will be equivalent to the parsed tag,
// but may not be identical.
func (spec *ArgSpec) String() string {
	if spec == nil {
		return ""
	}
	if spec.Start == 0 && spec.End == -1 {
		return "arg:[:]"
	}
	if spec.Start == spec.End {
		return fmt.Sprint(spec.Start)
	}
	return fmt.Sprintf("[%d:%d]", spec.Start, spec.End)
}

// FlagSpec describes parsed flags as defined in a facile struct tag.
type FlagSpec struct {
	Long  string
	Short string
}

// String representation of the flag spec. Will be equivalent to the parsed tag,
// but may not be identical.
func (spec *FlagSpec) String() string {
	if spec == nil {
		return ""
	}
	if spec.Posixy() {
		return fmt.Sprintf("flag:%v,%v", spec.Long, spec.Short)
	}
	return fmt.Sprintf("flag:%v", spec.Long)
}

// Posixy is true when the flag spec represents a flag which has both long and
// short flag forms.
func (spec *FlagSpec) Posixy() bool {
	if spec == nil {
		return false
	}
	return spec.Long != "" && len(spec.Short) == 1
}

// Tag as on members of a struct which will be used to define cliche command
// inputs.
type Tag string

// decompose a struct tag into distinct declarative components.
func (tag Tag) decompose() (arg, def, flag string) {
	if tag == "" {
		return
	}

	components := strings.Split(string(tag), ";")
	for _, c := range components {
		c = strings.TrimSpace(c)
		if a, ok := strings.CutPrefix(c, "arg:"); ok {
			arg = strings.TrimSpace(a)
			continue
		}
		if f, ok := strings.CutPrefix(c, "flag:"); ok {
			flag = strings.TrimSpace(f)
			continue
		}
		if d, ok := strings.CutPrefix(c, "default:"); ok {
			def = strings.TrimSpace(d)
			continue
		}
	}
	return
}

var argRe = regexp.MustCompile(`(\d+)|\[([^:]+)?(\:)?([^\]]+)?\]`)

// parseArg parses the arg value from a cliche struct tag.
func parseArg(tval string, spec *ArgSpec) bool {
	tval = strings.TrimSpace(tval)
	if tval == "" {
		return false
	}

	m := argRe.FindStringSubmatch(tval)
	if len(m) != 5 {
		// Better to return early than panic. This should only happen if argRe
		// is modified.
		return false
	}

	// Handle plain number (not slice index) notation.
	if iconv, err := strconv.Atoi(m[1]); m[1] != "" && err == nil {
		spec.Start = iconv
		return true
	}

	// Handle the slice index notation.
	s, r, e := m[2], m[3], m[4]
	if s == "" && r == "" && e == "" {
		// Empty brackets are verboten.
		return false
	}

	var err error
	var iconv int
	if s == "" {
		spec.Start = 0
	} else if iconv, err = strconv.Atoi(s); err == nil {
		spec.Start = iconv
	}

	if r == "" {
		// This arg spec is not a range. Nothing else needs doing.
		return (err == nil)
	}

	if e == "" {
		// No end of the range, so consume all remaining.
		spec.End = -1
	} else if iconv, err = strconv.Atoi(e); err == nil {
		spec.End = iconv
	}

	return (err == nil && // range end is a number
		spec.Start >= 0 && // range start is not negative
		((spec.End == -1) || (spec.End > spec.Start))) // end is larger than start, unless unbounded
}

// Arg returns the argument specification from a Tag, if any.
func (tag Tag) Arg() (*ArgSpec, bool) {
	arg, _, _ := tag.decompose()
	if arg == "" {
		return nil, false
	}

	var ret ArgSpec
	if parseArg(arg, &ret) {
		return &ret, true
	}
	return nil, false
}

// Default returns the string representation of the default value as specified
// in the struct tag.
func (tag Tag) Default() (string, bool) {
	_, def, _ := tag.decompose()
	if def != "" {
		return def, true
	}
	return "", false
}

var flagRe = regexp.MustCompile(`^([a-zA-Z][a-zA-Z0-9_-]+)(?:,\s*([a-zA-z]))?$`)

// parseFlag parses the flag value from a cliche struct tag.
func parseFlag(tval string, spec *FlagSpec) bool {
	tval = strings.TrimSpace(tval)
	if tval == "" {
		return false
	}
	m := flagRe.FindStringSubmatch(tval)
	if m == nil {
		return false
	}
	if len(m) != 3 {
		// This should only happen when flagRe is modified.
		return false
	}
	spec.Long = m[1]
	spec.Short = m[2]
	return true
}

// Flag returns the flag specifications from a Tag, if any.
func (tag Tag) Flag() (*FlagSpec, bool) {
	_, _, flag := tag.decompose()
	if flag == "" {
		return nil, false
	}

	var ret FlagSpec
	if parseFlag(flag, &ret) {
		return &ret, true
	}
	return nil, false
}
