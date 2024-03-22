package gen_handler

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/HobbyOSs/astv/internal/option"
	"github.com/lestrrat-go/codegen"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
)

func GenHandler(opts *option.Options) error {
	pkg, err := getPackageInfo(opts.Dir)
	if err != nil {
		return err
	}
	packageName := filepath.Base(pkg.PkgPath)

	if err := genHandlers(opts, packageName); err != nil {
		return errors.Wrap(err, `failed to generate handlers`)
	}
	if err := genVisitor(opts, packageName); err != nil {
		return errors.Wrap(err, `failed to generate visitor`)
	}
	return nil
}

func getPackageInfo(dir string) (*packages.Package, error) {
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

func genVisitor(opts *option.Options, packageName string) error {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, fmt.Sprintf("package %s", packageName))

	fmt.Fprintf(&buf, "\n\ntype Visitor struct {")
	for _, typ := range opts.AstTypes {
		fmt.Fprintf(&buf, "\nh%[1]s %[1]sHandler", typ)
	}
	fmt.Fprintf(&buf, "\nhDefault DefaultHandler")
	fmt.Fprintf(&buf, "\n}")

	fmt.Fprintf(&buf, "\n\nfunc (v *Visitor) Handler(h interface{}) error {")
	for _, typ := range opts.AstTypes {
		fmt.Fprintf(&buf, "\nif x, ok := h.(%sHandler); ok {", typ)
		fmt.Fprintf(&buf, "\nv.h%s = x", typ)
		fmt.Fprintf(&buf, "\n}")
	}
	fmt.Fprintf(&buf, "\nif x, ok := h.(DefaultHandler); ok {")
	fmt.Fprintf(&buf, "\nv.hDefault = x")
	fmt.Fprintf(&buf, "\n}")
	fmt.Fprintf(&buf, "\nreturn nil")
	fmt.Fprintf(&buf, "\n}")

	fmt.Fprintf(&buf, "\n\nfunc (v *Visitor) Visit(n ast.Node) ast.Visitor {")
	fmt.Fprintf(&buf, "\nswitch n := n.(type) {")
	for _, typ := range opts.AstTypes {
		fmt.Fprintf(&buf, "\ncase *ast.%s:", typ)
		fmt.Fprintf(&buf, "\nif h := v.h%s; h != nil {", typ)
		fmt.Fprintf(&buf, "\nif ! h.%s(n) {", typ)
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

	visitors_gen_path := filepath.Join(opts.Dir, "visitor_gen.go")
	if err := codegen.WriteFile(visitors_gen_path, &buf, codegen.WithFormatCode(opts.FormatCode)); err != nil {
		if cfe, ok := err.(codegen.CodeFormatError); ok {
			fmt.Fprint(os.Stderr, cfe.Source())
		}

		return errors.Wrap(err, `failed to write file`)
	}
	return nil
}

func genHandlers(opts *option.Options, packageName string) error {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, fmt.Sprintf("package %s", packageName))

	for _, typ := range opts.AstTypes {
		fmt.Fprintf(&buf, "\n\ntype %sHandler interface {", typ)
		fmt.Fprintf(&buf, "%[1]s(*ast.%[1]s) bool", typ)
		fmt.Fprintf(&buf, "\n}")
	}

	fmt.Fprintf(&buf, "\n\ntype DefaultHandler interface {")
	fmt.Fprintf(&buf, "\nHandle(ast.Node) bool")
	fmt.Fprintf(&buf, "\n}")

	handlers_gen_path := filepath.Join(opts.Dir, "handlers_gen.go")
	if err := codegen.WriteFile(handlers_gen_path, &buf, codegen.WithFormatCode(opts.FormatCode)); err != nil {
		fmt.Printf("written: %s \n", handlers_gen_path)
		if cfe, ok := err.(codegen.CodeFormatError); ok {
			fmt.Fprint(os.Stderr, cfe.Source())
		}

		return errors.Wrap(err, `failed to write file`)
	}
	return nil
}
