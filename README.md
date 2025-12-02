# xbrl-go

[![License: MIT](https://img.shields.io/badge/License-MIT-brightgreen?style=flat-square)](/LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/aethiopicuschan/xbrl-go.svg)](https://pkg.go.dev/github.com/aethiopicuschan/xbrl-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/aethiopicuschan/xbrl-go)](https://goreportcard.com/report/github.com/aethiopicuschan/xbrl-go)
[![CI](https://github.com/aethiopicuschan/xbrl-go/actions/workflows/ci.yaml/badge.svg)](https://github.com/aethiopicuschan/xbrl-go/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/aethiopicuschan/xbrl-go/graph/badge.svg?token=6A4Y75PXH5)](https://codecov.io/gh/aethiopicuschan/xbrl-go)

# xbrl-go

`xbrl-go` is a Go library and CLI tool for parsing and working with **XBRL (eXtensible Business Reporting Language)** documents.

Its goals are:

- Reliable and specification-compliant XBRL instance and taxonomy parser
- Extensible architecture to support common XBRL modules (Dimensions, iXBRL, Formula, ...)
- Developer-friendly API for extracting and validating business facts

---

## 📦 Installation

```sh
go get github.com/aethiopicuschan/xbrl-go/pkg/xbrl
```

`xbrl-go` is a Go library for parsing and handling XBRL (eXtensible Business Reporting Language) documents.

CLI (optional):

```sh
go install github.com/aethiopicuschan/xbrl-go/cmd/xbrl@latest
```

## 🧭 Usage (example)

```go
package main

import (
    "fmt"
    "github.com/aethiopicuschan/xbrl-go/pkg/xbrl"
)

func main() {
    doc, err := xbrl.ParseFile("sample.xbrl")
    if err != nil {
        panic(err)
    }

    // Print facts
    for _, f := range doc.Facts() {
        fmt.Printf("%s = %s (%s)\n", f.Name(), f.Value(), f.ContextRef())
    }
}
```

And you can see [example](./example) for more detailed usage.

CLI example:

```sh
xbrl facts sample.xbrl
```

## 🔧 Project structure

```
.
├── example # Example code demonstrating library usage
├── cmd
│   └── xbrl-go # CLI entrypoint
└── pkg/
    └── xbrl # Public library code
```

## 🤝 Contributing

Contributions are welcome!
Please read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a PR.
