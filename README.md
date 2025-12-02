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

## 🚀 Features (planned roadmap)

| Status | Feature                                                          |
| ------ | ---------------------------------------------------------------- |
| 🔜     | XBRL 2.1 core parsing (instances / taxonomy schemas / linkbases) |
| 🔜     | Calculation / Label / Presentation linkbase support              |
| 🔜     | Dimensions 1.0 support                                           |
| ⏳     | Inline XBRL (iXBRL) extraction                                   |
| ⏳     | XBRL validation (schema + linkbase consistency)                  |
| ⏳     | CLI utilities for validation / fact extraction                   |

> The library is under active development and not ready for production yet.

---

## 📦 Installation

```sh
go get github.com/aethiopicuschan/xbrl-go
```

`xbrl-go` is a Go library for parsing and handling XBRL (eXtensible Business Reporting Language) documents.

## Installation

```sh
go get -u github.com/aethiopicuschan/xbrl-go
```

## 🧭 Usage (example)

```go
package main

import (
    "fmt"
    "github.com/aethiopicuschan/xbrl-go"
)

func main() {
    doc, err := xbrl.ParseFile("sample.xbrl")
    if err != nil {
        panic(err)
    }

    // Print facts
    for _, f := range doc.Facts() {
        fmt.Printf("%s = %s (%s)\n", f.Name, f.Value, f.ContextRef)
    }
}
```

CLI example:

```sh
xbrl-go extract sample.xbrl --json
```

## 🔧 Project structure

```
.
├── cmd/              # CLI entrypoint
├── internal/         # Private implementation details
├── pkg/              # Public library code
│   ├── parser/       # Core parsers
│   ├── model/        # XBRL data structures
│   ├── linkbase/     # Linkbase handling
│   └── xbrl/         # High-level API
└── tests/            # Conformance / integration tests
```

## 🤝 Contributing

Contributions are welcome!
Please read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a PR.
