package shorten

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/pavelmemory/jobtome/internal"
	"github.com/pavelmemory/jobtome/internal/storage"
	"github.com/pavelmemory/jobtome/internal/storage/shorten"
)

const (
	hashLen = 7 // TODO: this should be configurable
)

type Entity struct {
	ID   int64
	URL  string
	Hash string
}

type Pager = shorten.Pager

//go:generate mockgen -source=service.go -destination mock.go -package shorten Storage

// Transactioner executes statements with/without explicitly open transaction.
type Transactioner interface {
	// WithTx executes provided callback inside of the transaction.
	// If callback returns an error the transaction will be rolled back, otherwise it will be committed.
	WithTx(context.Context, func(runner storage.Runner) error) error
	// WithoutTx executes provided callback without explicitly open transaction.
	WithoutTx(context.Context, func(runner storage.Runner) error) error
}

// Storage is a persistence storage for the shorten entity.
type Storage interface {
	// Persist saves the shorten and returns it's unique generated ID.
	Persist(ctx context.Context, run storage.Runner, shorten shorten.Entity) (int64, error)
	// Retrieve returns shorten by supplied 'id'.
	// If shorten doesn't exist it returns an error.
	Retrieve(ctx context.Context, run storage.Runner, id int64) (shorten.Entity, error)
	List(ctx context.Context, run storage.Runner, pager shorten.Pager) ([]shorten.Entity, error)
	// Delete deletes shorten entity by its identifier.
	Delete(ctx context.Context, runner storage.Runner, id int64) error
	// ByHash returns shorten by supplied 'hash'.
	ByHash(ctx context.Context, runner storage.Runner, hash string) (shorten.Entity, error)
}

// NewService returns initialized shorten service.
func NewService(tr Transactioner, storage Storage) *Service {
	return &Service{tr: tr, storage: storage}
}

// Service allows to CR_D shorten entity.
type Service struct {
	tr      Transactioner
	storage Storage
}

// Create creates a new shorten entity and returns back its unique ID.
func (s *Service) Create(ctx context.Context, short Entity) (int64, error) {
	if err := isNotBlank(short.URL, "url"); err != nil {
		return 0, err
	}

	if strings.TrimSpace(short.Hash) != "" {
		return 0, ValidationError{
			Cause:   internal.ErrBadInput,
			Details: map[string]interface{}{"hash": "not empty"},
		}
	}

	short.Hash = s.computeHash(short.URL)

	var id int64
	if err := s.tr.WithoutTx(ctx, func(runner storage.Runner) error {
		// check if such URL already exists
		existing, err := s.storage.ByHash(ctx, runner, short.Hash)
		if err != nil {
			if !errors.Is(err, internal.ErrNotFound) {
				return fmt.Errorf("lookup by hash %q: %w", short.Hash, err)
			}

			newShort := shorten.Entity{
				URL:       short.URL,
				Hash:      short.Hash,
				CreatedAt: time.Now(),
			}
			id, err = s.storage.Persist(ctx, runner, newShort)
			return err
		}

		id = existing.ID
		return nil
	}); err != nil {
		return 0, fmt.Errorf("persist short: %w", err)
	}

	return id, nil
}

func (s *Service) computeHash(long string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(long)))[:hashLen]
}

func (s *Service) Get(ctx context.Context, id int64) (Entity, error) {
	var short shorten.Entity
	if err := s.tr.WithoutTx(ctx, func(runner storage.Runner) (err error) {
		short, err = s.storage.Retrieve(ctx, runner, id)
		return err
	}); err != nil {
		return Entity{}, fmt.Errorf("retrieve shorten by id %q: %w", id, err)
	}

	return serviceEntity(short), nil
}

func (s *Service) List(ctx context.Context, pager Pager) ([]Entity, error) {
	if pager.Limit < 1 {
		return nil, ValidationError{
			Cause:   internal.ErrBadInput,
			Details: map[string]interface{}{"limit": "is lesser then 1"},
		}
	}

	if pager.Offset < 0 {
		return nil, ValidationError{
			Cause:   internal.ErrBadInput,
			Details: map[string]interface{}{"offset": "is negative"},
		}
	}

	var entities []Entity
	err := s.tr.WithoutTx(ctx, func(runner storage.Runner) error {
		shortens, err := s.storage.List(ctx, runner, pager)
		if err != nil {
			return err
		}

		entities = make([]Entity, len(shortens))
		for i, short := range shortens {
			entities[i] = serviceEntity(short)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("list shortens: %w", err)
	}

	return entities, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if err := s.tr.WithoutTx(ctx, func(runner storage.Runner) error {
		return s.storage.Delete(ctx, runner, id)
	}); err != nil {
		return fmt.Errorf("delete shorten %d: %w", id, err)
	}

	return nil
}

func (s *Service) Resolve(ctx context.Context, hash string) (string, error) {
	if err := isNotBlank(hash, "hash"); err != nil {
		return "", err
	}

	var url string
	err := s.tr.WithoutTx(ctx, func(runner storage.Runner) error {
		short, err := s.storage.ByHash(ctx, runner, hash)
		if err != nil {
			return err
		}

		url = short.URL

		return nil
	})
	if err != nil {
		return "", fmt.Errorf("retrieve shorten by hash %q: %w", hash, err)
	}

	return url, nil
}

// ValidationError encapsulates in it validation failure details.
type ValidationError struct {
	// Cause should be one of pre-defined standard errors.
	Cause error
	// Details any additional details about validation failure.
	Details map[string]interface{}
}

func (ve ValidationError) Error() string {
	d, _ := json.Marshal(struct {
		Cause   string                 `json:"cause"`
		Details map[string]interface{} `json:"details"`
	}{
		Cause:   ve.Cause.Error(),
		Details: ve.Details,
	})
	return string(d)
}

func (ve ValidationError) Is(err error) bool {
	return errors.Is(ve.Cause, err)
}

func isNotBlank(v, name string) error {
	if strings.TrimSpace(v) == "" {
		return ValidationError{
			Cause:   internal.ErrBadInput,
			Details: map[string]interface{}{name: "blank or empty"},
		}
	}

	return nil
}

func serviceEntity(u shorten.Entity) Entity {
	return Entity{
		ID:   u.ID,
		URL:  u.URL,
		Hash: u.Hash,
	}
}
