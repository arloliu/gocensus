# 000 - Agent Contract

Apply these rules for every task in this repository.

## Work From Evidence

- Read the relevant files before editing.
- State assumptions when they matter.
- Do not present guesses as facts.
- If a claim can be checked with source, tests, `go list`, `rg`, or git, check it.
- If verification is skipped, say what was skipped and why.

## Keep Scope Small

- Make the smallest change that solves the request.
- Do not add speculative features or broad refactors.
- Do not clean up unrelated code.
- Preserve user changes in the worktree.

## Verify Before Claiming

- Run the narrowest useful check after a change.
- Run `make check` before saying implementation work is complete.
- Separate product failures from local tooling or environment failures.
- Never say tests, lint, or build pass unless the command was run in the current turn.

