package p

import (
	"fmt"
	"io"
)

type User struct{ ID int }

type UserID int

type Repo struct{}

func (r *Repo) GetUser(id int) (*User, error)          { return &User{ID: id}, nil }
func (r *Repo) GetUserByID(id UserID) (*User, error)   { return nil, nil }
func (r *Repo) GetBoth(a int, b string) (*User, error) { return nil, nil }
func (r *Repo) Reorder(a, b int) int                   { return a - b }
func (r *Repo) Compute(prefix string, a int) int       { return a }
func (r *Repo) Plus(a int) int                         { return a + 1 }
func (r *Repo) Get(a int) int                          { return a }
func (r *Repo) Fetch(id string) ([]byte, error)        { return nil, nil }

type Service struct {
	repo   *Repo
	cached string
}

func (s *Service) GetUser(id int) (*User, error) { // want `shallowmethod: GetUser is a trivial pass-through`
	return s.repo.GetUser(id)
}

func (s *Service) GetBoth(a int, b string) (*User, error) { // want `shallowmethod: GetBoth is a trivial pass-through`
	return s.repo.GetBoth(a, b)
}

// Reordering is still pass-through — each arg is a param ident.
func (s *Service) Reorder(a, b int) int { // want `shallowmethod: Reorder is a trivial pass-through`
	return s.repo.Reorder(b, a)
}

func (s *Service) GetUserWrapped(id int) (*User, error) {
	u, err := s.repo.GetUser(id)
	if err != nil {
		return nil, fmt.Errorf("service.GetUser: %w", err)
	}
	return u, nil
}

func (s *Service) GetUserTyped(id int) (*User, error) {
	return s.repo.GetUserByID(UserID(id))
}

func (s *Service) Plus(a int) int {
	return s.repo.Plus(a + 1)
}

func (s *Service) WithField(a int) int {
	return s.repo.Compute(s.cached, a)
}

func (s *Service) WithConst(a int) int {
	return s.repo.Compute("fixed", a)
}

func (s *Service) WithBranch(a int) int {
	if a == 0 {
		return 0
	}
	return s.repo.Get(a)
}

func (s *Service) WithLog(a int) int {
	_ = a
	return s.repo.Get(a)
}

func NewService(r *Repo) *Service {
	return &Service{repo: r}
}

func (s *Service) Name() string {
	return s.describe()
}

func (s *Service) describe() string { return "svc" }

//nolint:shallowmethod
func (s *Service) Noisy(id int) (*User, error) {
	return s.repo.GetUser(id)
}

// FetchImpl implements Fetcher.
func (s *Service) FetchImpl(id string) ([]byte, error) {
	return s.repo.Fetch(id)
}

// Regression guard: "implement" in a sentence must not suppress.
//
// This layer doesn't really implement anything useful — it just wraps.
func (s *Service) NotImplementing(id int) int { // want `shallowmethod: NotImplementing is a trivial pass-through`
	return forward(id)
}

type Fetcher interface {
	Fetch(id string) ([]byte, error)
}

type Adapter struct {
	wrapped Fetcher
}

func (a *Adapter) Fetch(id string) ([]byte, error) {
	return a.wrapped.Fetch(id)
}

func Length(xs []int) int {
	return len(xs)
}

func ToInt(x float64) int {
	return int(x)
}

// --- variadic ---

func (r *Repo) Sum(vals ...int) int { return 0 }

func (s *Service) Sum(vals ...int) int { // want `shallowmethod: Sum is a trivial pass-through`
	return s.repo.Sum(vals...)
}

func (s *Service) SumOne(vals ...int) int {
	return s.repo.Sum(1)
}

// --- generics ---

type Box[T any] struct{ inner *Repo }

func (r *Repo) Generic(v int) int { return v }

func (b *Box[T]) Generic(v int) int { // want `shallowmethod: Generic is a trivial pass-through`
	return b.inner.Generic(v)
}

// --- value receiver implementing a local interface ---

type Writer interface {
	Write(data []byte) (int, error)
}

var globalWriter Writer

type ValueAdapter struct{}

func (v ValueAdapter) Write(data []byte) (int, error) {
	return globalWriter.Write(data)
}

// --- imported-interface adapters ---

type IOAdapter struct{ inner io.Writer }

func (a *IOAdapter) Write(p []byte) (int, error) {
	return a.inner.Write(p)
}

// Same shape but no matching interface — the unique name prevents any
// external interface from saving it.
func (a *IOAdapter) OusterhoutForward(p []byte) (int, error) { // want `shallowmethod: OusterhoutForward is a trivial pass-through`
	return a.inner.Write(p)
}

// Regression guard: "Newer" is not a constructor — only New/New[Upper].
func Newer(id int) int { // want `shallowmethod: Newer is a trivial pass-through`
	return forward(id)
}

func forward(id int) int { return id }
