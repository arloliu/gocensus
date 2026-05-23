# 300 - Testing

Apply before adding or changing tests.

## Organization

- Unit tests live next to the package under test.
- Prefer external test packages (`package foo_test`) when testing public behavior.
- Use internal package tests only when testing unexported helpers is the clearer option.
- Test fixtures should be created with `t.TempDir()` and `os.WriteFile`.

## Test Style

- Use the standard library testing package.
- Do not add assertion libraries unless there is a clear payoff.
- Keep tests behavior-focused and small.
- Use table tests only when there are multiple cases.
- Do not use sleeps for filesystem or CLI behavior; assert deterministic outputs.

## Coverage Expectations

Add or update tests for:

- File discovery and ignore behavior.
- Classification rules.
- Raw and effective line counting.
- Test/benchmark/example counting.
- Aggregation ratios.
- CLI flag parsing and output selection.
- Renderer output shape.

## Commands

```bash
go test ./...
make test
make check
```

