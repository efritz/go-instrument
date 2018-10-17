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

const (
	Name        = "go-instrument"
	PackageName = "github.com/efritz/go-instrument"
	Description = "go-instrument generates instrumented decorators for interfaces."
	Version     = "0.1.0"
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
			PackageName,
			ifaces,
			opts,
			generateFilename,
			generateInterface(*prefixValues),
		)
	}

	if err := command.Run(Name, Description, Version, types.GetInterface, generate, command.WithArgHook(argHook)); err != nil {
		log.Fatalf("error: %s\n", err.Error())
	}
}

func generateFilename(name string) string {
	return fmt.Sprintf("%s_instrumented.go", name)
}

func generateInterface(prefixValues PrefixValues) generation.InterfaceGenerator {
	return func(file *jen.File, iface *types.Interface, prefix string) {
		prefixes := []string{}
		for _, prefixValue := range prefixValues {
			prefixes = append(prefixes, prefixValue.Prefix)
		}

		var (
			titleName              = title(iface.Name)
			instrumentedStructName = fmt.Sprintf("Instrumented%s%s", prefix, titleName)
			constructorName        = fmt.Sprintf("New%s", instrumentedStructName)
		)

		file.Add(generateStruct(iface, instrumentedStructName))

		methods := []jen.Code{
			generateConstructor(iface, instrumentedStructName, constructorName),
		}

		for _, method := range iface.Methods {
			for _, prefixValue := range prefixValues {
				if prefixValue.Pattern.MatchString(method.Name) {
					methods = append(methods, generateInstrumentedMethod(
						iface,
						method,
						instrumentedStructName,
						prefixValue.Prefix,
					))

					break
				}
			}
		}

		for _, method := range methods {
			file.Add(method)
			file.Line()
		}
	}
}

func generateStruct(iface *types.Interface, instrumentedStructName string) jen.Code {
	fields := []jen.Code{
		jen.Qual(iface.ImportPath, iface.Name),
		jen.Id("reporter").Op("*").Qual("github.com/efritz/imperial/red", "Reporter"),
	}

	return jen.
		Type().
		Id(instrumentedStructName).
		Struct(fields...)
}

func generateConstructor(iface *types.Interface, instrumentedStructName, constructorName string) jen.Code {
	params := []jen.Code{
		jen.Id("inner").Qual(iface.ImportPath, iface.Name),
		jen.Id("reporter").Op("*").Qual("github.com/efritz/imperial/red", "Reporter"),
	}

	return generation.GenerateFunction(
		constructorName,
		params,
		[]jen.Code{jen.Op("*").Id(instrumentedStructName)},
		jen.Return().Op("&").Id(instrumentedStructName).Values(
			jen.Id(iface.Name).Op(":").Id("inner"),
			jen.Id("reporter").Op(":").Id("reporter"),
		),
	)
}

func generateInstrumentedMethod(iface *types.Interface, method *types.Method, instrumentedStructName string, prefix string) jen.Code {
	errorInterface := getErrorInterface()

	errArgument := jen.Nil()
	if len(method.Results) > 0 {
		lastResult := method.Results[len(method.Results)-1]

		if gotypes.Implements(lastResult, errorInterface) {
			errArgument = jen.Id(fmt.Sprintf("r%d", len(method.Results)-1))
		}
	}

	emitAttemptMetric := jen.
		Id("i").
		Dot("reporter").
		Dot("ReportAttempt").
		Call(jen.Lit(prefix))

	emitErrorMetric := jen.
		Id("i").
		Dot("reporter").
		Dot("ReportError").
		Call(jen.Lit(prefix), errArgument)

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
		Call(jen.Lit(prefix), jen.Id("duration"))

	return generation.GenerateOverride(
		"i",
		instrumentedStructName,
		iface.ImportPath,
		method,
		jen.Id("start").Op(":=").Qual("time", "Now").Call(),
		emitAttemptMetric,
		generation.GenerateDecoratedCall(method, jen.Id("i").Dot(iface.Name).Dot(method.Name)),
		setDuration,
		emitErrorMetric,
		emitDurationMetric,
		generation.GenerateDecoratedReturn(method),
	)
}

func getErrorInterface() *gotypes.Interface {
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

	pkg, err := conf.Check("error", fset, []*ast.File{file}, nil)
	if err != nil {
		panic(err.Error())
	}

	return pkg.Scope().Lookup("Error").Type().Underlying().(*gotypes.Interface)
}

func title(s string) string {
	if s == "" {
		return s
	}

	return strings.ToUpper(string(s[0])) + s[1:]
}
