package storefront

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"

	domaincategory "app/internal/domain/category"
	domainproduct "app/internal/domain/product"
	"app/internal/pkg/productref"
)

// BuildStoreListFilter parses storefront catalog query parameters.
func (s *Service) BuildStoreListFilter(ctx context.Context, q url.Values) (domainproduct.StoreListFilter, error) {
	filter := domainproduct.StoreListFilter{
		Query: strings.TrimSpace(q.Get("q")),
		Sort:  strings.TrimSpace(q.Get("sort")),
		Brand: strings.TrimSpace(q.Get("brand")),
	}

	if raw := q.Get("on_sale"); raw != "" {
		if parsed, err := strconv.ParseBool(raw); err == nil {
			filter.OnSale = &parsed
		}
	}
	if raw := q.Get("in_stock"); raw != "" {
		if parsed, err := strconv.ParseBool(raw); err == nil {
			filter.InStock = &parsed
		}
	}

	includeChildren := true
	if raw := q.Get("include_children"); raw != "" {
		if parsed, err := strconv.ParseBool(raw); err == nil {
			includeChildren = parsed
		}
	}
	filter.IncludeChildren = includeChildren

	if slug := strings.TrimSpace(q.Get("category_slug")); slug != "" {
		cat, err := s.categories.FindBySlug(ctx, slug)
		if err != nil {
			return filter, err
		}
		if includeChildren {
			ids, err := s.categorySubtreeIDs(ctx, cat.ID)
			if err != nil {
				return filter, err
			}
			filter.CategoryIDs = ids
		} else {
			filter.CategoryID = &cat.ID
		}
	} else if catID := strings.TrimSpace(q.Get("category_id")); catID != "" {
		id, err := uuid.Parse(catID)
		if err != nil {
			return filter, err
		}
		if includeChildren {
			ids, err := s.categorySubtreeIDs(ctx, id)
			if err != nil {
				return filter, err
			}
			filter.CategoryIDs = ids
		} else {
			filter.CategoryID = &id
		}
	}

	return filter, nil
}

func (s *Service) categorySubtreeIDs(ctx context.Context, rootID uuid.UUID) ([]uuid.UUID, error) {
	active := true
	items, err := s.categories.ListAll(ctx, domaincategory.ListFilter{IsActive: &active})
	if err != nil {
		return nil, err
	}

	children := make(map[uuid.UUID][]uuid.UUID)
	for _, item := range items {
		if item.ParentID != nil {
			children[*item.ParentID] = append(children[*item.ParentID], item.ID)
		}
	}

	ids := []uuid.UUID{rootID}
	queue := []uuid.UUID{rootID}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, childID := range children[current] {
			ids = append(ids, childID)
			queue = append(queue, childID)
		}
	}
	return ids, nil
}

// ResolveProductID resolves an active product slug or UUID.
func (s *Service) ResolveProductID(ctx context.Context, ref string) (uuid.UUID, error) {
	id, err := productref.ResolveID(ctx, s.products, ref)
	if err != nil {
		return uuid.Nil, err
	}
	product, err := s.products.FindByID(ctx, id)
	if err != nil {
		return uuid.Nil, err
	}
	if product.Status != domainproduct.StatusActive {
		return uuid.Nil, domainproduct.ErrNotFound
	}
	return id, nil
}
