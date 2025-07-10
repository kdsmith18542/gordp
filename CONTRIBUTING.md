# Contributing to GoRDP

Thank you for your interest in contributing to GoRDP! This document provides guidelines and information for contributors.

## Table of Contents

1. [Code of Conduct](#code-of-conduct)
2. [Getting Started](#getting-started)
3. [Development Setup](#development-setup)
4. [Making Changes](#making-changes)
5. [Testing](#testing)
6. [Code Style](#code-style)
7. [Documentation](#documentation)
8. [Submitting Changes](#submitting-changes)
9. [Release Process](#release-process)

## Code of Conduct

This project is committed to providing a welcoming and inclusive environment for all contributors. Please be respectful and considerate of others.

## Getting Started

### Prerequisites

- Go 1.18 or later
- Git
- Make (for build system)
- Docker (optional, for containerized development)

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/gordp.git
   cd gordp
   ```
3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/kdsmith18542/gordp.git
   ```

## Development Setup

### Quick Setup

```bash
# Setup development environment
make dev-setup

# Build the project
make build

# Run tests
make test
```

### Manual Setup

```bash
# Install dependencies
go mod download

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest

# Build
go build -o gordp .

# Run tests
go test -v ./...
```

## Making Changes

### Branch Strategy

1. Create a feature branch from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes and commit them:
   ```bash
   git add .
   git commit -m "feat: add new feature description"
   ```

3. Push your branch to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

### Commit Message Format

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

Examples:
```
feat: add keyboard input handling
fix: resolve connection timeout issue
docs: update API documentation
test: add integration tests for clipboard
```

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run specific test categories
go test -v -run TestKeyboardInput
go test -v -run TestMouseInput
go test -v -run TestClipboardFunctionality

# Run with coverage
make test-coverage

# Run benchmarks
make test-benchmark

# Run with race detection
make test-race
```

### Writing Tests

- Write tests for all new functionality
- Aim for high test coverage
- Use descriptive test names
- Test both success and failure cases
- Use table-driven tests when appropriate

Example:
```go
func TestKeyboardInput(t *testing.T) {
    tests := []struct {
        name     string
        key      rune
        modifier t128.ModifierKey
        wantErr  bool
    }{
        {"basic key press", 'a', t128.ModifierKey{}, false},
        {"with shift", 'A', t128.ModifierKey{Shift: true}, false},
        {"invalid key", 0, t128.ModifierKey{}, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := client.SendKeyPress(tt.key, tt.modifier)
            if (err != nil) != tt.wantErr {
                t.Errorf("SendKeyPress() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Integration Tests

```bash
# Run integration tests
cd tests/integration
go test -v -timeout=30s
```

## Code Style

### Go Conventions

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Follow Go naming conventions
- Use meaningful variable and function names

### Linting

```bash
# Run linter
make lint

# Format code
make fmt

# Check formatting
make fmt-check
```

### Code Quality

- Keep functions small and focused
- Use meaningful comments for complex logic
- Avoid code duplication
- Handle errors properly
- Use context for cancellation

## Documentation

### Updating Documentation

- Update README.md for user-facing changes
- Update API documentation in `docs/api.md`
- Add examples for new features
- Update inline code comments

### Documentation Standards

- Use clear, concise language
- Include code examples
- Keep documentation up to date with code changes
- Use proper markdown formatting

## Submitting Changes

### Pull Request Process

1. Ensure your code passes all tests:
   ```bash
   make ci
   ```

2. Update documentation as needed

3. Create a pull request on GitHub

4. Fill out the pull request template

5. Request review from maintainers

### Pull Request Guidelines

- Provide a clear description of changes
- Include tests for new functionality
- Update documentation if needed
- Link related issues
- Ensure CI checks pass

### Review Process

- All changes require review
- Address review comments promptly
- Maintainers may request changes
- Once approved, changes will be merged

## Release Process

### Versioning

We follow [Semantic Versioning](https://semver.org/):
- MAJOR.MINOR.PATCH
- MAJOR: Breaking changes
- MINOR: New features (backward compatible)
- PATCH: Bug fixes (backward compatible)

### Creating a Release

1. Update version in code
2. Create a git tag:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```
3. GitHub Actions will automatically:
   - Build binaries for all platforms
   - Create a GitHub release
   - Publish Docker images
   - Update package managers

### Release Checklist

- [ ] All tests pass
- [ ] Documentation is updated
- [ ] Version is updated
- [ ] Changelog is updated
- [ ] Release notes are prepared
- [ ] Tag is created and pushed

## Getting Help

### Questions and Issues

- Create an issue on GitHub for bugs
- Use discussions for questions
- Check existing issues and discussions
- Join our community channels

### Communication

- Be respectful and professional
- Provide clear, detailed information
- Include relevant code and error messages
- Follow up on issues and discussions

## Recognition

Contributors will be recognized in:
- GitHub contributors list
- Release notes
- Project documentation
- Community acknowledgments

Thank you for contributing to GoRDP! 