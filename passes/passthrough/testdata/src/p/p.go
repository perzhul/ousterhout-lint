package p

import (
	"context"
	"strconv"
	"testing"
)

type Repo struct{}

func (r *Repo) Get(id int) error                               { return nil }
func (r *Repo) GetCtx(ctx context.Context, id int) error       { return nil }
func (r *Repo) GetKey(key string, id int) error                { return nil }
func (r *Repo) Both(a, b int) error                            { return nil }

type Svc struct{ repo *Repo }

// REPORT: id flows straight into Get; the service adds no value to it.
func (s *Svc) Handle(id int) error { // want `passthrough: parameter "id"`
	_ = "side work"
	return s.repo.Get(id)
}

// REPORT: id is a passthrough even though ctx is exempt.
func (s *Svc) HandleCtx(ctx context.Context, id int) error { // want `passthrough: parameter "id"`
	_ = "side"
	return s.repo.GetCtx(ctx, id)
}

// OK: id is inspected (compared) in addition to being forwarded.
func (s *Svc) Inspected(id int) error {
	if id == 0 {
		return nil
	}
	return s.repo.Get(id)
}

// OK: id appears twice (other use besides the call).
func (s *Svc) Used(id int) error {
	_ = id
	return s.repo.Get(id)
}

// OK: id is transformed before forwarding.
func (s *Svc) Transformed(id int) error {
	_ = "side"
	return s.repo.Get(id + 1)
}

// OK: single-statement body is shallowmethod territory, not passthrough's.
func (s *Svc) Tiny(id int) error {
	return s.repo.Get(id)
}

// OK: context.Context is exempt — idiomatic Go plumbing.
func (s *Svc) ContextOnly(ctx context.Context) error {
	_ = "side"
	return s.repo.GetCtx(ctx, 0)
}

// OK: explicit nolint.
//
//nolint:passthrough
func (s *Svc) Silent(id int) error {
	_ = "side"
	return s.repo.Get(id)
}

// OK: param used across two call sites (broadcasting is legitimate).
func (s *Svc) Broadcast(id int) error {
	if err := s.repo.Get(id); err != nil {
		return err
	}
	return s.repo.Get(id)
}

// --- variadic ---

func (r *Repo) Sum(vals ...int) error { return nil }

// REPORT: variadic spread is still a direct forward.
func (s *Svc) HandleVariadic(vals ...int) error { // want `passthrough: parameter "vals"`
	_ = "side"
	return s.repo.Sum(vals...)
}

// --- generics ---

type Box[T any] struct{ repo *Repo }

func (r *Repo) HandleT(v int) error { return nil }

// REPORT: generic param forwarded to a same-package call.
func (b *Box[T]) Handle(v int) error { // want `passthrough: parameter "v"`
	_ = "side"
	return b.repo.HandleT(v)
}

// --- external-package forwarding is exempt ---

// OK: id forwarded to a stdlib call — cross-package forwarding is a
// conventional adapter pattern, not a shallow same-package wrapper.
func (s *Svc) Stdlib(id int) string {
	_ = "side"
	return strconv.Itoa(id)
}

// --- test-infrastructure plumbing is exempt ---

func assertT(t *testing.T)   { _ = t }
func assertTB(tb testing.TB) { _ = tb }

// OK: *testing.T is plumbing, like context.Context.
func HelperT(t *testing.T) {
	_ = "side"
	assertT(t)
}

// OK: testing.TB (interface) is also plumbing.
func HelperTB(tb testing.TB) {
	_ = "side"
	assertTB(tb)
}
