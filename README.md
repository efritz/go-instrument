# go-instrument

[![GoDoc](https://godoc.org/github.com/efritz/go-instrument?status.svg)](https://godoc.org/github.com/efritz/go-instrument)
[![Build Status](https://secure.travis-ci.org/efritz/go-instrument.png)](http://travis-ci.org/efritz/go-instrument)
[![Maintainability](https://api.codeclimate.com/v1/badges/2c875fc6956f08800c99/maintainability)](https://codeclimate.com/github/efritz/go-instrument/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/2c875fc6956f08800c99/test_coverage)](https://codeclimate.com/github/efritz/go-instrument/test_coverage)

An instrumenting code generator.

## Installation

Simply run `go get -u github.com/efritz/go-instrument/...`.

## Binary Usage

Usage coming soon.

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
| list          |            | Dry run - print the interfaces found in the given import paths. |
| metric-prefix |            | A "<regex>:<prefix>" pair. An instrumented method matching the regex emits metrics using the prefix. |

If neither dirname nor filename are supplied, then the generated code is printed to standard out.

## Instrumented Struct Usage

Usage coming soon.

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
