# Contributing to adocli

Thanks for your interest in contributing! 🎉

## Getting Started

1. Fork the repo
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/adocli.git`
3. Create a branch: `git checkout -b feature/my-feature`
4. Make your changes
5. Run tests: `make test`
6. Run linter: `make lint`
7. Commit and push
8. Open a Pull Request

## Development Setup

```bash
# Prerequisites
# - Go 1.22+
# - golangci-lint (optional, for linting)

git clone https://github.com/gyurisc/adocli.git
cd adocli
make build
./bin/ado --help
```

## Code Style

- Follow standard Go conventions
- Run `make fmt` before committing
- Run `make vet` to catch common issues
- All exported functions need doc comments

## Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

- `feat: add work item search`
- `fix: handle empty response from API`
- `docs: update README examples`
- `chore: update dependencies`

## Reporting Bugs

Open an issue with:
- What you expected
- What happened
- Steps to reproduce
- `ado --version` output

## Feature Requests

Open an issue describing:
- The problem you're trying to solve
- Your proposed solution
- Any alternatives you considered

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
