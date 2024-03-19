package gen_handler

import (
	"bytes"
	"fmt"
	"os"

	"github.com/lestrrat-go/codegen"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
)

var typs = []struct {
	Name string
}{
	{Name: "ArrayType"},
	{Name: "AssignStmt"},
	{Name: "BadDecl"},
	{Name: "BadExpr"},
	{Name: "BadStmt"},
	{Name: "BasicLit"},
	{Name: "BinaryExpr"},
	{Name: "BlockStmt"},
	{Name: "BranchStmt"},
	{Name: "CallExpr"},
	{Name: "CaseClause"},
	{Name: "ChanType"},
	{Name: "CommClause"},
	{Name: "Comment"},
	{Name: "CommentGroup"},
	{Name: "CompositeLit"},
	{Name: "DeclStmt"},
	{Name: "DeferStmt"},
	{Name: "Ellipsis"},
	{Name: "EmptyStmt"},
	{Name: "ExprStmt"},
	{Name: "Field"},
	{Name: "FieldList"},
	{Name: "ForStmt"},
	{Name: "FuncDecl"},
	{Name: "FuncLit"},
	{Name: "FuncType"},
	{Name: "GenDecl"},
	{Name: "GoStmt"},
	{Name: "Ident"},
	{Name: "IfStmt"},
	{Name: "ImportSpec"},
	{Name: "IncDecStmt"},
	{Name: "IndexExpr"},
	{Name: "InterfaceType"},
	{Name: "KeyValueExpr"},
	{Name: "LabeledStmt"},
	{Name: "MapType"},
	{Name: "ParenExpr"},
	{Name: "RangeStmt"},
	{Name: "ReturnStmt"},
	{Name: "SelectStmt"},
	{Name: "SelectorExpr"},
	{Name: "SendStmt"},
	{Name: "SliceExpr"},
	{Name: "StarExpr"},
	{Name: "StructType"},
	{Name: "SwitchStmt"},
	{Name: "TypeAssertExpr"},
	{Name: "TypeSpec"},
	{Name: "UnaryExpr"},
	{Name: "ValueSpec"},
}

func GenHandler(dir string) error {
	if err := genHandlers(); err != nil {
		return errors.Wrap(err, `failed to generate handlers`)
	}

	if err := genVisitor(dir); err != nil {
		return errors.Wrap(err, `failed to generate visitor`)
	}
	return nil
}

func getPackageName(dir string) (*packages.Package, error) {
	pkgs, err := packages.Load(&packages.Config{
		Mode:  packages.NeedName | packages.NeedFiles,
		Tests: false,
	}, dir)
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("cannot find any package in %v", dir)
	}
	return pkgs[0], nil
}

func genVisitor(dir string) error {
	var buf bytes.Buffer
	packageName, err := getPackageName(dir)
	if err != nil {
		return err
	}
	fmt.Fprintf(&buf, fmt.Sprintf("package %s", packageName))

	fmt.Fprintf(&buf, "\n\ntype Visitor struct {")
	for _, typ := range typs {
		fmt.Fprintf(&buf, "\nh%[1]s %[1]sHandler", typ.Name)
	}
	fmt.Fprintf(&buf, "\nhDefault DefaultHandler")
	fmt.Fprintf(&buf, "\n}")

	fmt.Fprintf(&buf, "\n\nfunc (v *Visitor) Handler(h interface{}) error {")
	for _, typ := range typs {
		fmt.Fprintf(&buf, "\nif x, ok := h.(%sHandler); ok {", typ.Name)
		fmt.Fprintf(&buf, "\nv.h%s = x", typ.Name)
		fmt.Fprintf(&buf, "\n}")
	}
	fmt.Fprintf(&buf, "\nif x, ok := h.(DefaultHandler); ok {")
	fmt.Fprintf(&buf, "\nv.hDefault = x")
	fmt.Fprintf(&buf, "\n}")
	fmt.Fprintf(&buf, "\nreturn nil")
	fmt.Fprintf(&buf, "\n}")

	fmt.Fprintf(&buf, "\n\nfunc (v *Visitor) Visit(n ast.Node) ast.Visitor {")
	fmt.Fprintf(&buf, "\nswitch n := n.(type) {")
	for _, typ := range typs {
		fmt.Fprintf(&buf, "\ncase *ast.%s:", typ.Name)
		fmt.Fprintf(&buf, "\nif h := v.h%s; h != nil {", typ.Name)
		fmt.Fprintf(&buf, "\nif ! h.%s(n) {", typ.Name)
		fmt.Fprintf(&buf, "\nreturn nil")
		fmt.Fprintf(&buf, "\n}")
		fmt.Fprintf(&buf, "\nreturn v")
		fmt.Fprintf(&buf, "\n}")
	}
	fmt.Fprintf(&buf, "\n}")
	// If it got here, there was no appropriate handler. Invoke the default handler, if available
	fmt.Fprintf(&buf, "\nif h := v.hDefault; h != nil {")
	fmt.Fprintf(&buf, "\nif !h.Handle(n) {")
	fmt.Fprintf(&buf, "\nreturn nil")
	fmt.Fprintf(&buf, "\n}")
	fmt.Fprintf(&buf, "\n}")
	fmt.Fprintf(&buf, "\nreturn v")
	fmt.Fprintf(&buf, "\n}")

	if err := codegen.WriteFile("visitor_gen.go", &buf, codegen.WithFormatCode(true)); err != nil {
		if cfe, ok := err.(codegen.CodeFormatError); ok {
			fmt.Fprint(os.Stderr, cfe.Source())
		}

		return errors.Wrap(err, `failed to write file`)
	}
	return nil
}

func genHandlers() error {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "package astv")

	for _, typ := range typs {
		fmt.Fprintf(&buf, "\n\ntype %sHandler interface {", typ.Name)
		fmt.Fprintf(&buf, "%[1]s(*ast.%[1]s) bool", typ.Name)
		fmt.Fprintf(&buf, "\n}")
	}

	fmt.Fprintf(&buf, "\n\ntype DefaultHandler interface {")
	fmt.Fprintf(&buf, "\nHandle(ast.Node) bool")
	fmt.Fprintf(&buf, "\n}")

	if err := codegen.WriteFile("handlers_gen.go", &buf, codegen.WithFormatCode(true)); err != nil {
		if cfe, ok := err.(codegen.CodeFormatError); ok {
			fmt.Fprint(os.Stderr, cfe.Source())
		}

		return errors.Wrap(err, `failed to write file`)
	}
	return nil
}
