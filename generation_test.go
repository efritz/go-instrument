package main

import (
	"fmt"
	gotypes "go/types"
	"regexp"
	"strings"

	"github.com/aphistic/sweet"
	"github.com/dave/jennifer/jen"
	"github.com/efritz/go-genlib/types"
	. "github.com/onsi/gomega"
)

type GenerationSuite struct{}

const (
	TestPrefix                 = "Test"
	TestTitleName              = "Client"
	TestInstrumentedStructName = "InstrumentedTestClient"
	TestImportPath             = "github.com/efritz/go-instrument/test"
)

var (
	TestPrefixValues = PrefixValues{
		PrefixValue{regexp.MustCompile(".+"), "test"},
	}

	stringType      = getType(gotypes.String)
	stringSliceType = gotypes.NewSlice(getType(gotypes.String))
	errorType       = getErrorType()

	TestMethodDo = &types.Method{
		Name:     "Do",
		Params:   []gotypes.Type{stringType, stringSliceType},
		Results:  []gotypes.Type{stringType, errorType},
		Variadic: true,
	}

	TestMethodTry = &types.Method{
		Name:     "Try",
		Params:   []gotypes.Type{stringType, stringSliceType},
		Results:  []gotypes.Type{stringType},
		Variadic: true,
	}
)

func (s *GenerationSuite) TestGenerateInterface(t sweet.T) {
	expectedDecls := []string{
		"type InstrumentedTestClient struct",
		"func NewInstrumentedTestClient(inner test.Client, reporter *red.Reporter) *InstrumentedTestClient",
		"func (i *InstrumentedTestClient) Do(v0 string, v1 ...string) (string, mockerror.Error)",
		"func (i *InstrumentedTestClient) Try(v0 string, v1 ...string) string",
	}

	file := jen.NewFile("test")
	g := &generator{""}
	g.generateInterface(TestPrefixValues)(file, makeBareInterface(TestMethodDo, TestMethodTry), TestPrefix)
	rendered := fmt.Sprintf("%#v\n", file)

	for _, decl := range expectedDecls {
		Expect(rendered).To(ContainSubstring(decl))
	}
}

func (s *GenerationSuite) TestGenerateStruct(t sweet.T) {
	g := &generator{""}
	code := g.generateStruct(makeInterface())

	Expect(fmt.Sprintf("%#v", code)).To(Equal(strip(`
	// InstrumentedTestClient is an wrapper around the Client interface (from
	// the package github.com/efritz/go-instrument/test) that emits request,
	// duration, and error metrics.
	type InstrumentedTestClient struct {
		test.Client
		reporter *red.Reporter
	}
	`)))
}

func (s *GenerationSuite) TestGenerateConstructor(t sweet.T) {
	g := &generator{""}
	code := g.generateConstructor(makeInterface())

	Expect(fmt.Sprintf("%#v", code)).To(Equal(strip(`
	// NewInstrumentedTestClient creates a new instrumented version of the
	// Client interface.
	func NewInstrumentedTestClient(inner test.Client, reporter *red.Reporter) *InstrumentedTestClient {
		return &InstrumentedTestClient{Client: inner, reporter: reporter}
	}
	`)))
}

func (s *GenerationSuite) TestGenerateInstrumentedMethod(t sweet.T) {
	g := &generator{""}
	code := g.generateInstrumentedMethod(makeMethod(TestMethodDo))

	Expect(fmt.Sprintf("%#v", code)).To(Equal(strip(`
	// Do delegates to the wrapped implementation and emits metrics with the
	// prefix 'test'.
	func (i *InstrumentedTestClient) Do(v0 string, v1 ...string) (string, mockerror.Error) {
		start := time.Now()
		i.reporter.ReportRequest("test")
		r0, r1 := i.Client.Do(v0, v1...)
		duration := float64(time.Now().Sub(start)) / float64(time.Second)
		i.reporter.ReportError("test", r1)
		i.reporter.ReportDuration("test", duration)
		return r0, r1
	}
	`)))
}

func (s *GenerationSuite) TestGenerateInstrumentedMethodNoError(t sweet.T) {
	g := &generator{""}
	code := g.generateInstrumentedMethod(makeMethod(TestMethodTry))

	Expect(fmt.Sprintf("%#v", code)).To(Equal(strip(`
	// Try delegates to the wrapped implementation and emits metrics with the
	// prefix 'test'.
	func (i *InstrumentedTestClient) Try(v0 string, v1 ...string) string {
		start := time.Now()
		i.reporter.ReportRequest("test")
		r0 := i.Client.Try(v0, v1...)
		duration := float64(time.Now().Sub(start)) / float64(time.Second)
		i.reporter.ReportError("test", nil)
		i.reporter.ReportDuration("test", duration)
		return r0
	}
	`)))
}

//
// Helpers

func getType(kind gotypes.BasicKind) gotypes.Type {
	return gotypes.Typ[kind].Underlying()
}

func makeBareInterface(methods ...*types.Method) *types.Interface {
	return &types.Interface{
		Name:       TestTitleName,
		ImportPath: TestImportPath,
		Type:       types.InterfaceTypeInterface,
		Methods:    methods,
	}
}

func makeInterface(methods ...*types.Method) *wrappedInterface {
	return wrapInterface(makeBareInterface(methods...), TestPrefix, TestTitleName, TestInstrumentedStructName, TestPrefixValues)
}

func makeMethod(methods ...*types.Method) (*wrappedInterface, *wrappedMethod) {
	wrapped := makeInterface(methods...)
	return wrapped, wrapped.wrappedMethods[0]
}

func strip(block string) string {
	lines := strings.Split(block, "\n")
	for i, line := range lines {
		if len(line) > 0 && line[0] == '\t' {
			lines[i] = line[1:]
		}
	}

	return strings.TrimSpace(strings.Join(lines, "\n"))
}
