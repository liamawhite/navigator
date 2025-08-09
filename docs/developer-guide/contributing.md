# Contributing

Thank you for your interest in contributing to Navigator! This guide will help you get started with contributing to the project.

## Development Setup

### Prerequisites

- [Nix](https://nixos.org/download.html) for development environment
- Git for version control

### Get Started

1. **Clone the repository**:
   ```bash
   git clone https://github.com/liamawhite/navigator.git
   cd navigator
   ```

2. **Enter development environment**:
   ```bash
   nix develop
   ```

3. **Build and test**:
   ```bash
   make build
   make test-unit
   ```

4. **Start locally**:
   ```bash
   make local
   ```

## Development Workflow

### Code Quality

Before submitting changes, ensure code quality:

```bash
# Format code (includes license headers)
make format

# Run linting
make lint

# Run all quality checks
make check
```

### Testing

```bash
# Run unit tests
make test-unit

# Run tests with verbose output
go test -v ./...

# Run tests for specific packages
go test ./manager/pkg/...
```

### Making Changes

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following the existing code patterns

3. **Add tests** for new functionality

4. **Update documentation** if needed

5. **Run quality checks**:
   ```bash
   make check
   ```

6. **Commit your changes**:
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   ```

## Code Guidelines

### Go Code Style

- Follow standard Go formatting (enforced by `gofmt`)
- Use meaningful variable and function names
- Add comments for exported functions and types
- Follow existing patterns in the codebase

### Commit Messages

Use conventional commit format:
- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation changes
- `refactor:` for code refactoring
- `test:` for adding tests
- `chore:` for maintenance tasks

### Project Structure

Follow the established project structure:
- `manager/` - Manager service code
- `edge/` - Edge service code
- `navctl/` - CLI tool code
- `pkg/` - Shared packages
- `api/` - Protocol buffer definitions
- `ui/` - React frontend
- `docs/` - Documentation

## API Changes

### Protocol Buffers

When modifying APIs:

1. **Update `.proto` files** in `api/`
2. **Regenerate code**:
   ```bash
   make generate
   ```
3. **Test changes** thoroughly
4. **Update documentation** if API contracts change

### Breaking Changes

- Avoid breaking changes in stable APIs
- If breaking changes are necessary, increment version appropriately
- Document breaking changes clearly in PR description

## Documentation

### Types of Documentation

1. **Code comments** - For developers reading the code
2. **User documentation** - In `docs/user-guide/`
3. **Developer documentation** - In `docs/developer-guide/`
4. **API documentation** - Generated from protobuf files

### Documentation Guidelines

- Keep documentation up-to-date with code changes
- Use clear, concise language
- Provide examples where helpful
- Link related documentation sections

## Submitting Changes

### Pull Requests

1. **Push your branch** to GitHub:
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create a pull request** with:
   - Clear title and description
   - Reference any related issues
   - Include testing information
   - List any breaking changes

3. **Address review feedback** promptly

4. **Keep your branch updated** with main:
   ```bash
   git fetch origin
   git rebase origin/main
   ```

### Review Process

- All code changes require review
- CI checks must pass
- Documentation should be updated for user-facing changes
- Breaking changes require special attention

## Getting Help

- **GitHub Issues** - For bug reports and feature requests
- **GitHub Discussions** - For questions and general discussion
- **Code Review** - Ask questions in pull request comments

## License

By contributing to Navigator, you agree that your contributions will be licensed under the Apache License 2.0.

## Recognition

Contributors are recognized in the project's release notes and GitHub contributor statistics. Thank you for helping make Navigator better!