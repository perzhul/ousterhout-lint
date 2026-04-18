package p

import (
	"context"
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
