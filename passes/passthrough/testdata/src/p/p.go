package p

import (
	"context"
	"fmt"
	"strconv"
	"testing"
)

type Repo struct{}

func (r *Repo) Get(id int) error                         { return nil }
func (r *Repo) GetCtx(ctx context.Context, id int) error { return nil }
func (r *Repo) GetKey(key string, id int) error          { return nil }
func (r *Repo) Both(a, b int) error                      { return nil }

type Svc struct{ repo *Repo }

func (s *Svc) Handle(id int) error { // want `passthrough: parameter "id"`
	_ = "side work"
	return s.repo.Get(id)
}

func (s *Svc) HandleCtx(ctx context.Context, id int) error { // want `passthrough: parameter "id"`
	_ = "side"
	return s.repo.GetCtx(ctx, id)
}

func (s *Svc) Inspected(id int) error {
	if id == 0 {
		return nil
	}
	return s.repo.Get(id)
}

func (s *Svc) Used(id int) error {
	_ = id
	return s.repo.Get(id)
}

func (s *Svc) Transformed(id int) error {
	_ = "side"
	return s.repo.Get(id + 1)
}

// Single-statement body is shallowmethod's job.
func (s *Svc) Tiny(id int) error {
	return s.repo.Get(id)
}

func (s *Svc) ContextOnly(ctx context.Context) error {
	_ = "side"
	return s.repo.GetCtx(ctx, 0)
}

//nolint:passthrough
func (s *Svc) Silent(id int) error {
	_ = "side"
	return s.repo.Get(id)
}

// Broadcasting the same value to multiple call sites is legitimate.
func (s *Svc) Broadcast(id int) error {
	if err := s.repo.Get(id); err != nil {
		return err
	}
	return s.repo.Get(id)
}

// --- variadic ---

func (r *Repo) Sum(vals ...int) error { return nil }

func (s *Svc) HandleVariadic(vals ...int) error { // want `passthrough: parameter "vals"`
	_ = "side"
	return s.repo.Sum(vals...)
}

// --- generics ---

type Box[T any] struct{ repo *Repo }

func (r *Repo) HandleT(v int) error { return nil }

func (b *Box[T]) Handle(v int) error { // want `passthrough: parameter "v"`
	_ = "side"
	return b.repo.HandleT(v)
}

// --- cross-package forwarding is exempt ---

func (s *Svc) Stdlib(id int) string {
	_ = "side"
	return strconv.Itoa(id)
}

// --- test-infra plumbing is exempt ---

func assertT(t *testing.T)   { _ = t }
func assertTB(tb testing.TB) { _ = tb }

func HelperT(t *testing.T) {
	_ = "side"
	assertT(t)
}

func HelperTB(tb testing.TB) {
	_ = "side"
	assertTB(tb)
}

// --- value-add markers exempt the function ---

// Error wrapping via fmt.Errorf — real value, not a shallow layer.
func (s *Svc) Wrapped(id int) error {
	err := s.repo.Get(id)
	if err != nil {
		return fmt.Errorf("svc: %w", err)
	}
	return nil
}

// For loop — function is iterating, not plumbing.
func (s *Svc) Loop(id int) error {
	for i := 0; i < 3; i++ {
		_ = i
	}
	return s.repo.Get(id)
}

// Switch — dispatching is real work.
func (s *Svc) Dispatch(id int) error {
	switch id {
	case 0:
		return nil
	}
	return s.repo.Get(id)
}

// Two calls in body, param forwarded once — the second call signals
// composition, so id is not "just plumbing".
func (s *Svc) Compose(id int) error {
	_ = s.repo.Get(42)
	return s.repo.Get(id)
}

// Constructor by naming convention — exempt like shallowmethod does.
type Wrapper struct{ inner *Repo }

func NewWrapper(r *Repo, id int) *Wrapper {
	_ = id
	return &Wrapper{inner: r}
}
