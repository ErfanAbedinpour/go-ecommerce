package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	domain "app/internal/domain/cart"
	"app/internal/infrastructure/persistence/models"
)

// CartRepository implements cart.Repository using PostgreSQL.
type CartRepository struct {
	db *gorm.DB
}

// NewCartRepository creates a PostgreSQL-backed cart repository.
func NewCartRepository(db *gorm.DB) *CartRepository {
	return &CartRepository{db: db}
}

func (r *CartRepository) Get(ctx context.Context, owner domain.Owner) (*domain.Cart, error) {
	cartModel, err := r.findCart(ctx, owner)
	if err != nil {
		return nil, err
	}
	if cartModel == nil {
		return &domain.Cart{Items: []domain.Item{}}, nil
	}

	var itemModels []models.CartItemModel
	if err := r.db.WithContext(ctx).
		Where("cart_id = ?", cartModel.ID).
		Order("added_at ASC").
		Find(&itemModels).Error; err != nil {
		return nil, err
	}

	items := make([]domain.Item, len(itemModels))
	for i, m := range itemModels {
		items[i] = domain.Item{
			ProductID: m.ProductID,
			SkuID:     m.SkuID,
			Quantity:  m.Quantity,
			AddedAt:   m.AddedAt,
		}
	}

	return &domain.Cart{
		Items:     items,
		UpdatedAt: cartModel.UpdatedAt,
	}, nil
}

func (r *CartRepository) Save(ctx context.Context, owner domain.Owner, cart *domain.Cart) error {
	if cart == nil {
		cart = &domain.Cart{Items: []domain.Item{}}
	}
	cart.UpdatedAt = time.Now().UTC()

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		cartModel, err := r.findOrCreateCartTx(tx, owner)
		if err != nil {
			return err
		}

		if err := tx.Where("cart_id = ?", cartModel.ID).Delete(&models.CartItemModel{}).Error; err != nil {
			return err
		}

		if len(cart.Items) > 0 {
			itemModels := make([]models.CartItemModel, len(cart.Items))
			for i, item := range cart.Items {
				itemModels[i] = models.CartItemModel{
					ID:        uuid.New(),
					CartID:    cartModel.ID,
					ProductID: item.ProductID,
					SkuID:     item.SkuID,
					Quantity:  item.Quantity,
					AddedAt:   item.AddedAt,
				}
			}
			if err := tx.Create(&itemModels).Error; err != nil {
				return err
			}
		}

		return tx.Model(cartModel).Update("updated_at", cart.UpdatedAt).Error
	})
}

func (r *CartRepository) Delete(ctx context.Context, owner domain.Owner) error {
	cartModel, err := r.findCart(ctx, owner)
	if err != nil {
		return err
	}
	if cartModel == nil {
		return nil
	}
	return r.db.WithContext(ctx).Delete(cartModel).Error
}

func (r *CartRepository) findCart(ctx context.Context, owner domain.Owner) (*models.CartModel, error) {
	var cart models.CartModel
	q := r.db.WithContext(ctx).Model(&models.CartModel{})
	if owner.UserID != nil {
		q = q.Where("user_id = ?", *owner.UserID)
	} else {
		q = q.Where("guest_token = ?", owner.GuestToken)
	}
	if err := q.First(&cart).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &cart, nil
}

func (r *CartRepository) findOrCreateCartTx(tx *gorm.DB, owner domain.Owner) (*models.CartModel, error) {
	var cart models.CartModel
	q := tx.Model(&models.CartModel{})
	if owner.UserID != nil {
		q = q.Where("user_id = ?", *owner.UserID)
	} else {
		q = q.Where("guest_token = ?", owner.GuestToken)
	}

	err := q.First(&cart).Error
	if err == nil {
		return &cart, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	cart = models.CartModel{ID: uuid.New()}
	if owner.UserID != nil {
		cart.UserID = owner.UserID
	} else {
		token := owner.GuestToken
		cart.GuestToken = &token
	}
	if err := tx.Create(&cart).Error; err != nil {
		return nil, err
	}
	return &cart, nil
}

var _ domain.Repository = (*CartRepository)(nil)
