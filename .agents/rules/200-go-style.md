# 200 - Go Style

Apply when editing Go code.

## Compatibility

- Keep code compatible with Go 1.21+.
- Do not use newer language or standard-library features unless the project intentionally raises the floor.
- `slices` and `maps` are allowed; Go 1.22+ loop forms and Go 1.24+ benchmark APIs are not.

## Code Style

- Use idiomatic Go and simple package boundaries.
- Prefer plain structs and functions over framework-style abstractions.
- Keep public API types in the root package stable and documented.
- Keep CLI formatting separate from analysis logic.
- Use `context.Context` for cancellable analysis work.
- Return errors as the last return value.
- Wrap errors with `fmt.Errorf("context: %w", err)`.
- Use `errors.Is` and `errors.As` where callers need classification.

## File Layout

Use this order for new files and touched regions:

1. Package declaration
2. Imports
3. Constants
4. Variables
5. Types
6. Constructors
7. Exported functions
8. Unexported functions
9. Methods grouped by receiver

## Naming

- Package names are short and lowercase.
- Exported names use `CamelCase`; unexported names use `camelCase`.
- Receiver names are short and consistent.
- Avoid abbreviations that make CLI/report fields unclear.

