// Package meta contains intermediate types and utilities used to parse a Go
// AST and generate code to implement a cliche CLI.
package meta

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"io"
	"log/slog"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
)

// CommandInput contains details about how a Command's inputs should be
// mapped to the struct members of the implementing type.
type CommandInput struct {
	FieldName string
	Tag       Tag
	Doc       string
	Type      string
}

// Command compiles details about how a type should be wrapped for cliche from
// the AST describing it. This type is used to execute a Go template, to
// generate the resulting Go source file.
type Command struct {
	// Name of the command to be generated. This defaults to the name of the
	// package, unless overridden.
	Name string

	// Package name from which the  Command is sourced.
	Package string

	// Type name of the  Command implementation.
	Type string

	// Help output for the  Command. This will be displayed along with usage
	// information on the command line. By default, sourced from doc comment for
	// the package in which the wrapped Command will live.
	Help string

	// Description of the command. Should be short and human readable. By
	// default, sourced from the doc comment on the wrapped  Command type.
	Description string

	// Inputs describe the handling of struct fields on the wrapped Command
	// implementation as inputs on the command line. The inputs are derived from
	// struct tags, when set.
	Inputs []CommandInput

	typ string
}

func compileInputs(st *ast.StructType) (inputs []CommandInput) {
	if st == nil || st.Fields == nil {
		return
	}
	var name string
	for _, field := range st.Fields.List {
		// Both nameless fields and fields with multiple names are skipped.
		// Maybe someday it will be worth unwinding the ambiguity of what to do
		// in these cases. That day is not today.
		n := len(field.Names)
		switch {
		case n == 0:
			// A field has no name.
			slog.Info(fmt.Sprintf("Skipping nameless field of type %v", field.Type))
			continue
		case n > 1:
			slog.Warn(fmt.Sprintf("Skipping field with multiple names: %v;  cannot handle this case.", field.Names))
			continue
		default:
			if !field.Names[0].IsExported() {
				// If the only name is unexported, ignore this field as well.
				slog.Info(fmt.Sprintf("Skipping unexported field %v", field.Names[0]))
				continue
			}
			name = field.Names[0].Name
			slog.Info(fmt.Sprintf("Compiling field named %q", name))
		}

		// If the field has a doc comment, capture it for the command usage
		// output.
		var doc string
		if field.Doc != nil {
			doc = field.Doc.Text()
			slog.Info(fmt.Sprintf("Field %v has doc comment: %q", name, doc))
		} else {
			slog.Info(fmt.Sprintf("Field %v has no doc comment", name))
		}

		// If the field has an  struct tag, capture and parse it for setting
		// flags, handling args, and / or setting default values. The reflect
		// package has some built-in struct tag parsing logic. No reason not to
		// use that.
		var stag reflect.StructTag
		if field.Tag != nil {
			tv := field.Tag.Value
			slog.Info(fmt.Sprintf("Field %v has tag: %v", field, tv))
			// The token contained by the AST is still a quoted string.
			utv, err := strconv.Unquote(tv)
			if err == nil {
				stag = reflect.StructTag(utv)
			} else {
				slog.Warn(fmt.Sprintf("Couldn't unquote struct tag %q: %v", tv, err))
			}
		}

		var tag Tag
		if t, ok := stag.Lookup(""); ok {
			tag = Tag(t)
			slog.Info(fmt.Sprintf("Field %v has  tag %q", name, tag))
		} else {
			slog.Info(fmt.Sprintf("Field %v has no  tag", name))
		}

		inputs = append(inputs, CommandInput{
			FieldName: name,
			Tag:       tag,
			Doc:       doc,
			Type:      fmt.Sprintf("%s", field.Type),
		})
	}
	return
}

// Compile the AST of a Go file into command metadata. Designed to be used as
// an argument to ast.Inspect.
func (meta *Command) Compile(n ast.Node) bool {
	if meta == nil || n == nil {
		return false
	}
	switch x := n.(type) {
	case *ast.TypeSpec:
		if x.Name == nil || x.Name.Name != meta.typ {
			// This is not the type we are looking for.
			break
		}
		if st, ok := x.Type.(*ast.StructType); ok {
			meta.Inputs = append(meta.Inputs, compileInputs(st)...)
			// We've got what we came for.
			return false
		}
	}
	return true
}

// commandName makes a decision about what the subcommand will be called on the
// command line. The following procedure is used:
//
// 1) A base command name is selected:
//   - If the --subcommand_name flag is set, its value is used
//   - Otherwise, the name of the package containing the Command implementation is used
//
// 2) The selected base name is converted to kebab-case
func commandName(pkg string) string {
	name := pkg
	// TODO(christian): Overrides?
	return strcase.ToKebab(name)
}

var whitespaceRunsRe = regexp.MustCompile(`\s+`)

func sanitizeHelp(doc, pkg, cmd string) string {
	var ok bool
	if doc, ok = strings.CutPrefix(doc, "Package "); !ok {
		slog.Warn("Package doc comment is malformed; proceeding anyway",
			slog.String("package", pkg))
	}
	if doc == "" {
		slog.Warn("Package has no doc comment", slog.String("package", pkg))
		return ""
	}

	// Replace the package name in the doc comment string, if it exists.
	if strings.HasPrefix(doc, pkg) {
		doc = strings.Replace(doc, pkg, cmd, 1)
	}
	return strings.TrimSpace(whitespaceRunsRe.ReplaceAllLiteralString(doc, " "))
}

type namedReader interface {
	io.Reader
	Name() string
}

const importPath = "idontfixcomputers.com/cliche/fakenotused"

// FromFile parses a Go AST from a file-like object and generates a Command for
// a type matching typeName. If errors are encountered, nil is returned.
func FromFile(from namedReader, typeName string) *Command {
	filename := from.Name()
	src, err := io.ReadAll(from)
	if err != nil {
		slog.Error("Failed reading", slog.Any("error", err))
		return nil
	}

	// First, we must parse the file into an AST. The ParseComments mode is used
	// to include comments during parsing.
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil || f == nil {
		slog.Warn("Failed creating AST from file",
			slog.String("file", filename), slog.Any("error", err))
		return nil
	}

	// Next, do a pass over the AST with interpreter from the go/doc package,
	// which goes to great lengths to compute doc comments. No reason to
	// reimplement that logic. Mode PreserveAST is used so that the AST is not
	// modified during doc generation, so that the same AST can be reused by our
	// own parser, below.
	pkg, err := doc.NewFromFiles(fset, []*ast.File{f}, importPath, doc.PreserveAST)
	if err != nil {
		slog.Warn("Failed to compute documentation from AST from file",
			slog.String("file", filename), slog.Any("error", err))
		return nil
	}

	// After the doc computation is complete, we look for our target type in the
	// results. The return value from NewFromFiles contains AST nodes along with
	// documentation.
	var ourType *doc.Type
	for _, typ := range pkg.Types {
		if typ.Name != typeName {
			continue
		}
		ourType = typ
		break
	}
	if ourType == nil {
		// The type we are looking for does not exist in the AST.
		slog.Warn("Type not found in file",
			slog.String("file", filename), slog.String("type", typeName))
		return nil
	}

	// Finally, create the metadata struct and allow it to parse the AST from
	// the node the doc package found for our type.
	cmdActual := commandName(pkg.Name)
	meta := &Command{
		Name:        cmdActual,
		Package:     pkg.Name,
		Type:        ourType.Name,
		Help:        sanitizeHelp(pkg.Doc, pkg.Name, cmdActual),
		Description: strings.TrimSpace(ourType.Doc),
		// Inputs are generated during Compile().
	}
	ast.Inspect(ourType.Decl, meta.Compile)
	return meta
}
