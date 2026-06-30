package productref

import (
	"context"

	"github.com/google/uuid"

	domainproduct "app/internal/domain/product"
)

// ResolveID returns a product UUID from a slug or UUID string.
func ResolveID(ctx context.Context, products domainproduct.Repository, ref string) (uuid.UUID, error) {
	if id, err := uuid.Parse(ref); err == nil {
		if _, err := products.FindByID(ctx, id); err != nil {
			return uuid.Nil, err
		}
		return id, nil
	}

	product, err := products.FindBySlug(ctx, ref)
	if err != nil {
		return uuid.Nil, err
	}
	return product.ID, nil
}
