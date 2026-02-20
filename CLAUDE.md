# CLAUDE.md

Instructions for AI agents working in this repository.

## Build & Test

```bash
go test ./...                          # unit tests only
go test ./... -tags=integration        # include integration tests
golangci-lint run ./...                # lint (must pass with 0 issues)
```

Every change must pass `golangci-lint run ./...` with zero issues before it is considered complete. Do not modify `.golangci.yml` to suppress lint findings — fix the code instead.

## Coding Conventions

- **File ordering**: Within each file, order declarations alphabetically: `const`, `type`, `var` blocks first, then functions/methods. The `NewTypeName` constructor must be the first function after its type definition; remaining methods follow alphabetically.
- **Switch/case ordering**: Order `case` clauses alphabetically within `switch` statements.
- **Constructors**: `NewTypeName(deps) *TypeName` — always return a pointer.
- **Error handling**: Wrap with context using `fmt.Errorf("operation: %w", err)`. Check errors immediately.
- **Imports**: Group as stdlib, then blank line, then internal packages. Use aliases when needed.
- **No dead code**: Eliminate unused code during refactoring. Do not keep dead functions, types, variables, or imports. If code is no longer referenced, delete it.
- **No external dependencies** beyond the standard library.
- **Reuse before creating**: Before introducing a new helper, utility, or abstraction, check whether an existing implementation in the repository already covers the need. Only add new code when no suitable reuse option exists.

## Testing Conventions

- **Naming**: `Test_<Unit>_With_<Condition>_Should_<Outcome>`
- **Structure**: Strict Arrange/Act/Assert with explicit `// Arrange`, `// Act`, `// Assert` comments.
- **Parallelism**: Every test starts with `t.Parallel()`.
- **Assertions**: Use `assert.That(t, "description", got, expected)` from the `assert` package.
- **Integration tests**: Use `//go:build integration` build tag on the first line.
