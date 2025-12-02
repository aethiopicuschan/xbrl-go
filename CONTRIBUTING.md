# Contributing to xbrl-go

Thanks for your interest in contributing!  
This project aims to be a high-quality and production-ready XBRL toolkit for Go.  
Any help — bug reports, discussions, code, docs — is welcome.

## 📌 Development workflow

1. Fork and clone the repository
2. Create a feature branch
3. Make changes with tests
4. Ensure CI passes locally
5. Open a Pull Request

Branch naming suggestion:

- feature/...
- fix/...
- refactor/...
- docs/...

---

## 🔎 Code quality and style

- Go 1.25+
- Prefer small and testable packages
- Public API must have documentation comments
- No breaking changes without discussion

## 🧰 Pre-commit hooks

This project uses [lefthook](https://lefthook.dev/) for enforcing quality checks.
Hooks run automatically on commit:

| Tool                                       | Purpose            |
| ------------------------------------------ | ------------------ |
| [staticcheck](https://staticcheck.dev/)    | Go static analysis |
| [typos](https://github.com/crate-ci/typos) | Spell checking     |

Install hooks:

```sh
lefthook install
```

You can manually run all hooks:

```sh
lefthook run pre-commit
```

## 🧪 Tests

- Contributions must include relevant tests where applicable
- Prefer table-driven tests
- For XBRL files used as fixtures, place them under `tests/data/`

To run tests:

```sh
go test ./...
```

## 🗨️ Discussions & Issues

- Questions / ideas → GitHub Discussions
- Bug reports → GitHub Issues
- Security issues → DO NOT open an issue. Email the maintainer instead.

---

Thank you again for contributing to xbrl-go 🚀
