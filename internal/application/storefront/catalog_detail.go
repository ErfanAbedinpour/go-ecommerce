package storefront

import (
	"context"

	"github.com/google/uuid"

	domainproduct "app/internal/domain/product"
	domainreview "app/internal/domain/productreview"
)

// ReviewsSummary is the public aggregate rating projection.
type ReviewsSummary struct {
	AverageRating float64       `json:"average_rating"`
	TotalCount    int64         `json:"total_count"`
	Distribution  map[int]int64 `json:"distribution"`
}

// ProductDetailOptions controls optional embedded fields on product detail.
type ProductDetailOptions struct {
	IncludeReviewsSummary bool
	IncludeWishlist       bool
	UserID                *uuid.UUID
}

// ProductDetailEnriched extends product detail with optional embedded data.
type ProductDetailEnriched struct {
	ProductDetail
	ReviewsSummary *ReviewsSummary `json:"reviews_summary,omitempty"`
	IsInWishlist   *bool           `json:"is_in_wishlist,omitempty"`
}

// ReviewsSummaryProvider loads aggregate review stats for a product.
type ReviewsSummaryProvider interface {
	GetSummary(ctx context.Context, productID uuid.UUID) (*domainreview.Summary, error)
}

// WishlistChecker checks whether a product is in a customer's wishlist.
type WishlistChecker interface {
	Check(ctx context.Context, userID uuid.UUID, productID uuid.UUID) (bool, error)
}

// GetProductEnriched returns product detail with optional embedded fields.
func (s *Service) GetProductEnriched(
	ctx context.Context,
	slugOrID string,
	opts ProductDetailOptions,
	reviews ReviewsSummaryProvider,
	wishlist WishlistChecker,
) (*ProductDetailEnriched, error) {
	detail, err := s.GetProduct(ctx, slugOrID)
	if err != nil {
		return nil, err
	}

	out := &ProductDetailEnriched{ProductDetail: *detail}

	if opts.IncludeReviewsSummary && reviews != nil {
		summary, err := reviews.GetSummary(ctx, detail.ID)
		if err != nil {
			return nil, err
		}
		out.ReviewsSummary = &ReviewsSummary{
			AverageRating: summary.AverageRating,
			TotalCount:    summary.TotalCount,
			Distribution:  summary.Distribution,
		}
	}

	if opts.IncludeWishlist {
		inWishlist := false
		out.IsInWishlist = &inWishlist
		if opts.UserID != nil && wishlist != nil {
			checked, err := wishlist.Check(ctx, *opts.UserID, detail.ID)
			if err != nil {
				return nil, err
			}
			*out.IsInWishlist = checked
		}
	}

	return out, nil
}

// ToProductCard converts a domain product to the shared storefront card DTO.
func ToProductCard(p *domainproduct.Product) ProductCard {
	return toProductCard(p)
}
