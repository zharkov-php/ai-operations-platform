package dependencies

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Postgres struct{ Pool *pgxpool.Pool }

func (p Postgres) Ping(ctx context.Context) error { return p.Pool.Ping(ctx) }

type Redis struct{ Client *redis.Client }

func (r Redis) Ping(ctx context.Context) error { return r.Client.Ping(ctx).Err() }
