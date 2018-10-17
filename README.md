# go-instrument

[![GoDoc](https://godoc.org/github.com/efritz/go-instrument?status.svg)](https://godoc.org/github.com/efritz/go-instrument)
[![Build Status](https://secure.travis-ci.org/efritz/go-instrument.png)](http://travis-ci.org/efritz/go-instrument)
[![Maintainability](https://api.codeclimate.com/v1/badges/2c875fc6956f08800c99/maintainability)](https://codeclimate.com/github/efritz/go-instrument/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/2c875fc6956f08800c99/test_coverage)](https://codeclimate.com/github/efritz/go-instrument/test_coverage)

An instrumenting code generator.

## Installation

Simply run `go get -u github.com/efritz/go-instrument/...`.

## Binary Usage

As an example, we generate an instrumented version of the `Client` interface from the
Redis library [deepjoy](https://github.com/efritz/deepjoy). If the deepjoy library can
be found in the GOPATH, then the following command will generate a file called
`client_instrumented.go` with the following content. This assumes that the current
working directory (Also in the GOPATH) is called *playground*.

```bash
go-instrument github.com/efritz/deepjoy -i Client --metric-prefix Do:redis
```

```go
// Code generated by github.com/efritz/go-instrument; DO NOT EDIT.
// This file was generated by robots at
// 2018-10-16T09:12:34-05:00
// using the command
// $ go-instrument github.com/efritz/deepjoy -i Client --metric-prefix Do:redis

package playground

import (
	deepjoy "github.com/efritz/deepjoy"
	red "github.com/efritz/imperial/red"
	"time"
)

type InstrumentedClient struct {
	deepjoy.Client
	reporter *red.Reporter
}

func NewInstrumentedClient(inner deepjoy.Client, reporter *red.Reporter) *InstrumentedClient {
	return &InstrumentedClient{Client: inner, reporter: reporter}
}

func (i *InstrumentedClient) Do(v0 string, v1 ...interface{}) (interface{}, error) {
	start := time.Now()
	i.reporter.ReportAttempt("redis")
	r0, r1 := i.Client.Do(v0, v1...)
	duration := float64(time.Now().Sub(start)) / float64(time.Second)
	i.reporter.ReportError("redis", r1)
	i.reporter.ReportDuration("redis", duration)
	return r0, r1
}
```

The suggested way to generate instrumented version of an interface for a project is to
use go-generate. Then, when the interface is updated, running `go generate` on the
package will re-generate the instrumented versions.

```go
package foo

//go:generate go-instrument -f github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface --metric-prefix '.*:dynamodb'
```

### Flags

The following flags are defined by the binary.

| Name          | Short Flag | Description  |
| ------------- | ---------- | ------------ |
| package       | p          | The name of the generated package. Is the name of target directory if dirname or filename is supplied by default. |
| prefix        |            | A prefix used in the name of each instrumented struct. Should be TitleCase by convention. |
| interfaces    | i          | A whitelist of interfaces to generate given the import paths. |
| filename      | o          | The target output file. All instrumented structs are writen to this file. |
| dirname       | d          | The target output directory. Each instrumented will be written to a unique file. |
| force         | f          | Do not abort if a write to disk would overwrite an existing file. |
| metric-prefix |            | A "<regex>:<prefix>" pair. An instrumented method matching the regex emits metrics using the prefix. |

If neither dirname nor filename are supplied, then the generated code is printed to standard out.

If no metric prefixes are supplied, then the instrumented version generated is
effectively a no-operation decorator around a concrete implemetnation.

## Instrumented Struct Usage

An instrumented version of an interface is initialized via a constructor that
takes a concrete instance of the underlying interface as well as an instance of
a [RED Metric Reporter](https://github.com/efritz/imperial/tree/master/red).

It is required that this reporter be initialized to configurations for each of
metric prefixes that were supplied at generation time. For example, a valid RED
reporter for the example above is given below.

```go
import "github.com/efritz/imperial/red"

func makeInstrumentedClient(client deepjoy.Client, reporter imperial.Reporter) {
    redisPrefixConfig := red.NewPrefixConfig(
        // Report operation durations in these buckets (in seconds)
        red.WithPrefixBuckets([]float64{0.01, 0.1, 0.5, 1}),
    )

    // Create red reporter with `redis` prefix configuration
    redReporter := red.NewReporter(
        reporter,
        red.WithPrefixConfig("redis", redisPrefixConfig),
    )

    // Wrap the client
    return NewInstrumentedClient(client, redReporter)
}
```

## License

Copyright (c) 2018 Eric Fritz

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
