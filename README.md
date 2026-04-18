# ousterhout-lint

[![CI](https://github.com/perzhul/ousterhout-lint/actions/workflows/ci.yml/badge.svg)](https://github.com/perzhul/ousterhout-lint/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/perzhul/ousterhout-lint/branch/main/graph/badge.svg)](https://codecov.io/gh/perzhul/ousterhout-lint)
[![Go Report Card](https://goreportcard.com/badge/github.com/perzhul/ousterhout-lint)](https://goreportcard.com/report/github.com/perzhul/ousterhout-lint)
[![Go Reference](https://pkg.go.dev/badge/github.com/perzhul/ousterhout-lint.svg)](https://pkg.go.dev/github.com/perzhul/ousterhout-lint)

A Go static analyzer that enforces John Ousterhout's **deep modules** principle from *A Philosophy of Software Design*.

Ousterhout's central thesis is that good modules hide significant complexity behind simple interfaces. This linter catches the *mechanically detectable* shape of shallow modules — trivial wrapper methods and parameters that flow straight through without adding value. Semantic depth (does this abstraction truly hide complexity?) is left to code review.

Scope is intentionally tight. The linter is *not* a general Ousterhout toolkit. It ships two passes, both targeting gaps that existing Go linters don't cover.

## Passes

| Pass | What it catches |
|---|---|
| `shallowmethod` | Methods whose entire body forwards the outer parameters to a single inner call. `func (s *Service) GetUser(id int) (*User, error) { return s.repo.GetUser(id) }` |
| `passthrough` | Individual parameters that flow into a downstream same-package call without any inspection, transformation, or branching. |

Existing tools (`revive`, `gocritic`, `godoc`, `funlen`) already cover parameter counts, exported doc presence, and naming conventions — those rules are deliberately *not* re-implemented here.

## Examples

### `shallowmethod` — flagged

```go
func (s *Service) GetUser(id int) (*User, error) {
    return s.repo.GetUser(id)
}
```

```
shallowmethod: GetUser is a trivial pass-through to s.repo.GetUser(...);
remove this layer or let it add real value
```

Either delete the method and let callers hit the repo directly, or let the wrapper do something real — error wrapping, auth, caching, metrics.

### `passthrough` — flagged

```go
func (s *Service) Handle(id int) error {
    log.Print("handling")
    return s.repo.Get(id)
}
```

```
passthrough: parameter "id" is forwarded to s.repo.Get(...) without
inspection; consider whether this layer adds value
```

`id` isn't inspected or transformed — it just rides through. Either do something with it, or hoist the logging and drop the wrapper.

### Not flagged

```go
// error wrapping adds real value
func (s *Service) GetUser(id int) (*User, error) {
    u, err := s.repo.GetUser(id)
    if err != nil {
        return nil, fmt.Errorf("service.GetUser: %w", err)
    }
    return u, nil
}

// adapter implementing an external interface — legitimate thin wrapper
func (a *IOAdapter) Write(p []byte) (int, error) {
    return a.inner.Write(p)
}

// context.Context is idiomatic plumbing, exempt from passthrough
func (s *Service) Fetch(ctx context.Context, id int) error {
    metrics.Inc("fetch")
    return s.downstream.Fetch(ctx, id)
}
```

## Install

```bash
go install github.com/perzhul/ousterhout-lint/cmd/ousterhout-lint@latest
```

Run on a module:

```bash
ousterhout-lint ./...
```

## golangci-lint plugin

Build the plugin and wire it into your project's `.golangci.yml`:

```bash
make plugin
cp ousterhout.so /path/to/your/project/
```

```yaml
# .golangci.yml in your project
linters-settings:
  custom:
    ousterhout:
      path: ./ousterhout.so
      description: Deep-module enforcer
      original-url: github.com/perzhul/ousterhout-lint
linters:
  enable:
    - ousterhout
```

### Plugin compatibility note

The `.so` plugin uses Go's `-buildmode=plugin`, which requires the plugin to be compiled with the **exact same Go toolchain version** as the golangci-lint binary that loads it. A mismatch produces a runtime load error at lint time. If you hit that, rebuild the plugin with the Go version reported by `golangci-lint version`.

golangci-lint's newer [Module Plugin System](https://golangci-lint.run/plugins/module-plugins/) sidesteps the toolchain-coupling problem by compiling the plugin into a custom `golangci-lint` binary via `golangci-lint custom`. Support for that distribution path is planned for a future release; contributions welcome.

## Escape hatches

Both passes respect standard suppression:

- `//nolint:shallowmethod` or `//nolint:passthrough` on the function's doc comment.
- A doc comment containing `implements <InterfaceName>` suppresses `shallowmethod` — adapter and bridge methods are legitimate thin wrappers.
- `shallowmethod` auto-skips when the receiver type implements a local interface with the method's name.
- Constructors (names starting with `New`) are exempt from `shallowmethod`.
- `passthrough` exempts `context.Context` parameters (idiomatic Go plumbing).
- `passthrough` skips forwarding to external-package functions — stdlib/library adapters are legitimate boundary-crossers, not shallow modules.

## What this linter is *not*

- It does not judge whether a module is *conceptually* deep. That's a human call.
- It does not enforce package structure, naming, comment quality, or parameter counts. Use `depguard`, `revive`, and `funlen` for those.
- It does not do cross-package whole-program analysis.

## License

MIT.
