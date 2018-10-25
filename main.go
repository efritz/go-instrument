package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	gotypes "go/types"
	"log"
	"strings"

	"github.com/alecthomas/kingpin"

	"github.com/dave/jennifer/jen"
	"github.com/efritz/go-genlib/command"
	"github.com/efritz/go-genlib/generation"
	"github.com/efritz/go-genlib/types"
)

type (
	wrappedInterface struct {
		*types.Interface
		prefix                 string
		titleName              string
		instrumentedStructName string
		wrappedMethods         []*wrappedMethod
	}

	wrappedMethod struct {
		*types.Method
		iface  *types.Interface
		prefix string
	}

	topLevelGenerator func(*wrappedInterface) jen.Code
	methodGenerator   func(*wrappedInterface, *wrappedMethod) jen.Code
)

const (
	name        = "go-instrument"
	packageName = "github.com/efritz/go-instrument"
	description = "go-instrument generates instrumented decorators for interfaces."
	version     = "0.1.0"

	instrumentedStructFormat = "Instrumented%s%s"
)

var (
	topLevelGenerators = []topLevelGenerator{
		generateStruct,
		generateConstructor,
	}

	methodGenerators = []methodGenerator{
		generateInstrumentedMethod,
	}
)

func init() {
	log.SetFlags(0)
	log.SetPrefix("go-instrument: ")
}

func main() {
	var prefixValues *PrefixValues

	argHook := func(app *kingpin.Application) {
		prefixValues = PrefixValuesFlag(app.Flag("metric-prefix", ""))
	}

	generate := func(ifaces []*types.Interface, opts *command.Options) error {
		return generation.Generate(
			packageName,
			version,
			ifaces,
			opts,
			generateFilename,
			generateInterface(*prefixValues),
		)
	}

	if err := command.Run(name, description, version, types.GetType, generate, command.WithArgHook(argHook)); err != nil {
		log.Fatalf("error: %s\n", err.Error())
	}
}

func generateFilename(name string) string {
	return fmt.Sprintf("%s_instrumented.go", name)
}

func generateInterface(prefixValues PrefixValues) generation.InterfaceGenerator {
	return func(file *jen.File, iface *types.Interface, prefix string) {
		var (
			titleName              = title(iface.Name)
			instrumentedStructName = fmt.Sprintf(instrumentedStructFormat, prefix, titleName)
			wrappedInterface       = wrapInterface(iface, prefix, titleName, instrumentedStructName, prefixValues)
		)

		for _, generator := range topLevelGenerators {
			file.Add(generator(wrappedInterface))
			file.Line()
		}

		for _, method := range wrappedInterface.wrappedMethods {
			for _, generator := range methodGenerators {
				file.Add(generator(wrappedInterface, method))
				file.Line()
			}
		}
	}
}

func wrapInterface(iface *types.Interface, prefix, titleName, instrumentedStructName string, prefixValues PrefixValues) *wrappedInterface {
	wrapped := &wrappedInterface{
		Interface:              iface,
		prefix:                 prefix,
		titleName:              titleName,
		instrumentedStructName: instrumentedStructName,
	}

	for _, method := range iface.Methods {
		for _, prefixValue := range prefixValues {
			if !prefixValue.Pattern.MatchString(method.Name) {
				continue
			}

			method := wrapMethod(iface, method, prefixValue.Prefix)
			wrapped.wrappedMethods = append(wrapped.wrappedMethods, method)
			break
		}
	}

	return wrapped
}

func wrapMethod(iface *types.Interface, method *types.Method, prefix string) *wrappedMethod {
	m := &wrappedMethod{
		Method: method,
		iface:  iface,
		prefix: prefix,
	}

	return m
}

//
// Instrumented Struct Generation

func generateStruct(iface *wrappedInterface) jen.Code {
	comment := generation.GenerateComment(
		1,
		"%s is an wrapper around the %s interface (from the package %s) that emits request, duration, and error metrics.",
		iface.instrumentedStructName,
		iface.Name,
		iface.ImportPath,
	)

	fields := []jen.Code{
		jen.Qual(iface.ImportPath, iface.Name),
		jen.Id("reporter").Op("*").Qual("github.com/efritz/imperial/red", "Reporter"),
	}

	return comment.
		Type().
		Id(iface.instrumentedStructName).
		Struct(fields...)
}

//
// Constructor Generation

func generateConstructor(iface *wrappedInterface) jen.Code {
	name := fmt.Sprintf("New%s", iface.instrumentedStructName)

	comment := generation.GenerateComment(
		1,
		"%s creates a new instrumented version of the %s interface.",
		name,
		iface.Name,
	)

	params := []jen.Code{
		jen.Id("inner").Qual(iface.ImportPath, iface.Name),
		jen.Id("reporter").Op("*").Qual("github.com/efritz/imperial/red", "Reporter"),
	}

	decl := generation.GenerateFunction(
		name,
		params,
		[]jen.Code{jen.Op("*").Id(iface.instrumentedStructName)},
		jen.Return().Op("&").Id(iface.instrumentedStructName).Values(
			jen.Id(iface.Name).Op(":").Id("inner"),
			jen.Id("reporter").Op(":").Id("reporter"),
		),
	)

	return generation.Compose(comment, decl)
}

//
// Instrumented Method Generation

func generateInstrumentedMethod(iface *wrappedInterface, method *wrappedMethod) jen.Code {
	comment := generation.GenerateComment(
		1,
		"%s delegates to the wrapped implementation and emits metrics with the prefix '%s'.",
		method.Name,
		method.prefix,
	)

	errorInterface := getErrorInterface()

	errArgument := jen.Nil()
	if len(method.Results) > 0 {
		if gotypes.Implements(method.Results[len(method.Results)-1], errorInterface) {
			errArgument = jen.Id(fmt.Sprintf("r%d", len(method.Results)-1))
		}
	}

	emitRequestMetric := jen.
		Id("i").
		Dot("reporter").
		Dot("ReportRequest").
		Call(jen.Lit(method.prefix))

	emitErrorMetric := jen.
		Id("i").
		Dot("reporter").
		Dot("ReportError").
		Call(jen.Lit(method.prefix), errArgument)

	setDuration := jen.
		Id("duration").
		Op(":=").
		Id("float64").
		Call(jen.Qual("time", "Now").Call().Dot("Sub").Call(jen.Id("start"))).
		Op("/").
		Id("float64").
		Call(jen.Qual("time", "Second"))

	emitDurationMetric := jen.
		Id("i").
		Dot("reporter").
		Dot("ReportDuration").
		Call(jen.Lit(method.prefix), jen.Id("duration"))

	override := generation.GenerateOverride(
		jen.Id("i").Op("*").Id(iface.instrumentedStructName),
		iface.ImportPath,
		method.Method,
		jen.Id("start").Op(":=").Qual("time", "Now").Call(),
		emitRequestMetric,
		generation.GenerateDecoratedCall(method.Method, jen.Id("i").Dot(iface.Name).Dot(method.Name)),
		setDuration,
		emitErrorMetric,
		emitDurationMetric,
		generation.GenerateDecoratedReturn(method.Method),
	)

	return generation.Compose(comment, override)
}

//
// Helpers

func title(s string) string {
	if s == "" {
		return s
	}

	return strings.ToUpper(string(s[0])) + s[1:]
}

func getErrorInterface() *gotypes.Interface {
	return getErrorType().Underlying().(*gotypes.Interface)
}

func getErrorType() gotypes.Type {
	errorSource := `
	package error

	type Error interface {
		Error() string
	}
	`

	fset := token.NewFileSet()
	conf := gotypes.Config{Importer: importer.Default()}

	file, err := parser.ParseFile(fset, "error.go", errorSource, 0)
	if err != nil {
		panic(err.Error())
	}

	pkg, err := conf.Check("mock-error", fset, []*ast.File{file}, nil)
	if err != nil {
		panic(err.Error())
	}

	return pkg.Scope().Lookup("Error").Type()
}
