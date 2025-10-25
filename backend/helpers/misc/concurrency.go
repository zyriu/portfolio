package misc

import (
	"context"

	"golang.org/x/sync/errgroup"
)

func Go[T any](g *errgroup.Group, ctx context.Context, fn func(context.Context) (T, error), dst *T) {
	g.Go(func() error {
		v, err := fn(ctx)
		if err != nil {
			return err
		}
		*dst = v
		return nil
	})
}
