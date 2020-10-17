package shorten

import (
	"context"
	"fmt"
	"time"

	"github.com/pavelmemory/jobtome/internal"
	"github.com/pavelmemory/jobtome/internal/storage"
)

type Entity struct {
	ID        int64
	URL       string
	Hash      string
	CreatedAt time.Time
}

type Repo struct{}

func (p Repo) Persist(ctx context.Context, run storage.Runner, entry Entity) (int64, error) {
	const query = `
		INSERT INTO shorten(id, url, hash, created_at) 
		VALUES (NULL, $1, $2, $3)`

	res := run.Exec(ctx, query, entry.URL, entry.Hash, time.Now().Unix())
	if err := storage.ConvertError(res.Err()); err != nil {
		return 0, fmt.Errorf("exec: %w", err)
	}

	return res.ID(), nil
}

func (p Repo) Retrieve(ctx context.Context, run storage.Runner, id int64) (Entity, error) {
	const query = `
		SELECT url, hash, created_at
		FROM shorten
		WHERE id = $1`

	entity := Entity{ID: id}
	var createdAt int64

	res := run.QuerySingle(ctx, query, id)
	if err := storage.ConvertError(res.Scan(&entity.URL, &entity.Hash, &createdAt)); err != nil {
		return Entity{}, fmt.Errorf("retrieve single: %w", err)
	}
	entity.CreatedAt = time.Unix(createdAt, 0)

	return entity, nil
}

// Pager is a lightweight paging abstraction that we could afford to use only with small databases like SQLLite.
type Pager struct {
	Limit  int64
	Offset int64
}

func (Repo) List(ctx context.Context, run storage.Runner, pager Pager) ([]Entity, error) {
	const query = `
		SELECT id, url, hash, created_at
		FROM shorten
		ORDER BY id
		LIMIT $1 
		OFFSET $2`

	var entities []Entity

	res, err := run.Query(ctx, query, pager.Limit, pager.Offset)
	if err := storage.ConvertError(err); err != nil {
		return nil, fmt.Errorf("retrieve multiple: %w", err)
	}
	defer res.Close() // TODO: proper handling of closing error

	for res.Next() {
		var entity Entity
		var createdAt int64
		err := res.Scan(&entity.ID, &entity.URL, &entity.Hash, &createdAt)
		if err := storage.ConvertError(err); err != nil {
			return nil, fmt.Errorf("scan retrieved: %w", err)
		}
		entity.CreatedAt = time.Unix(createdAt, 0)
		entities = append(entities, entity)
	}

	return entities, nil
}

func (Repo) Delete(ctx context.Context, run storage.Runner, id int64) error {
	const query = `DELETE FROM shorten WHERE id = $1`

	res := run.Exec(ctx, query, id)
	if err := storage.ConvertError(res.Err()); err != nil {
		return fmt.Errorf("exec delete: %w", err)
	}

	if res.Affected() == 1 {
		return nil
	}

	return internal.ErrNotFound
}

func (Repo) ByHash(ctx context.Context, run storage.Runner, hash string) (Entity, error) {
	const query = `
		SELECT id, url, created_at
		FROM shorten
		WHERE hash = $1`

	entity := Entity{Hash: hash}
	var createdAt int64

	res := run.QuerySingle(ctx, query, hash)
	if err := storage.ConvertError(res.Scan(&entity.ID, &entity.URL, &createdAt)); err != nil {
		return Entity{}, fmt.Errorf("retrieve single: %w", err)
	}
	entity.CreatedAt = time.Unix(createdAt, 0)

	return entity, nil
}
