package response

import (
	"time"

	domaincontact "app/internal/domain/contact"
	domainquestion "app/internal/domain/productquestion"
	domainreview "app/internal/domain/productreview"
	domainwishlist "app/internal/domain/wishlist"
	"app/pkg/pagination"
)

// ── Wishlist Responses ──────────────────────────────────────────

type WishlistProductSummaryResponse struct {
	Name      string   `json:"name"`
	Slug      string   `json:"slug"`
	Price     float64  `json:"price"`
	SalePrice *float64 `json:"sale_price,omitempty"`
	ImageURL  string   `json:"image_url,omitempty"`
	IsInStock bool     `json:"is_in_stock"`
}

type WishlistItemResponse struct {
	ID        string                         `json:"id"`
	ProductID string                         `json:"product_id"`
	CreatedAt string                         `json:"created_at"`
	Product   WishlistProductSummaryResponse `json:"product"`
}

type WishlistListResponse struct {
	Data []WishlistItemResponse `json:"data"`
	Meta pagination.Meta        `json:"meta"`
}

func ToWishlistItemResponse(item domainwishlist.ListItem) WishlistItemResponse {
	return WishlistItemResponse{
		ID:        item.ID.String(),
		ProductID: item.ProductID.String(),
		CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339),
		Product: WishlistProductSummaryResponse{
			Name:      item.Product.Name,
			Slug:      item.Product.Slug,
			Price:     item.Product.Price,
			SalePrice: item.Product.SalePrice,
			ImageURL:  item.Product.ImageURL,
			IsInStock: item.Product.IsInStock,
		},
	}
}

func ToWishlistListResponse(paginated pagination.Paginated[domainwishlist.ListItem]) WishlistListResponse {
	data := make([]WishlistItemResponse, len(paginated.Data))
	for i, item := range paginated.Data {
		data[i] = ToWishlistItemResponse(item)
	}
	return WishlistListResponse{
		Data: data,
		Meta: paginated.Meta,
	}
}

// ── Product Review Responses ─────────────────────────────────────

type ProductReviewResponse struct {
	ID         string     `json:"id"`
	ProductID  string     `json:"product_id"`
	CustomerID *string    `json:"customer_id,omitempty"`
	AuthorName string     `json:"author_name"`
	Rating     int        `json:"rating"`
	Title      string     `json:"title,omitempty"`
	Content    string     `json:"content"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
}

type ProductReviewSummaryResponse struct {
	AverageRating float64       `json:"average_rating"`
	TotalCount    int64         `json:"total_count"`
	Distribution  map[int]int64 `json:"distribution"`
}

type ProductReviewListResponse struct {
	Data []ProductReviewResponse `json:"data"`
	Meta pagination.Meta         `json:"meta"`
}

type AdminProductReviewListItemResponse struct {
	ProductReviewResponse
	ProductName string `json:"product_name"`
}

type AdminProductReviewListResponse struct {
	Data []AdminProductReviewListItemResponse `json:"data"`
	Meta pagination.Meta                      `json:"meta"`
}

func ToProductReviewResponse(rev domainreview.Review) ProductReviewResponse {
	var custID *string
	if rev.CustomerID != nil {
		idStr := rev.CustomerID.String()
		custID = &idStr
	}
	return ProductReviewResponse{
		ID:         rev.ID.String(),
		ProductID:  rev.ProductID.String(),
		CustomerID: custID,
		AuthorName: rev.AuthorName,
		Rating:     rev.Rating,
		Title:      rev.Title,
		Content:    rev.Content,
		Status:     string(rev.Status),
		CreatedAt:  rev.CreatedAt,
	}
}

func ToProductReviewListResponse(paginated pagination.Paginated[domainreview.Review]) ProductReviewListResponse {
	data := make([]ProductReviewResponse, len(paginated.Data))
	for i, r := range paginated.Data {
		data[i] = ToProductReviewResponse(r)
	}
	return ProductReviewListResponse{
		Data: data,
		Meta: paginated.Meta,
	}
}

func ToProductReviewSummaryResponse(s *domainreview.Summary) *ProductReviewSummaryResponse {
	return &ProductReviewSummaryResponse{
		AverageRating: s.AverageRating,
		TotalCount:    s.TotalCount,
		Distribution:  s.Distribution,
	}
}

func ToAdminProductReviewListItemResponse(item domainreview.AdminListItem) AdminProductReviewListItemResponse {
	return AdminProductReviewListItemResponse{
		ProductReviewResponse: ToProductReviewResponse(item.Review),
		ProductName:           item.ProductName,
	}
}

func ToAdminProductReviewListResponse(paginated pagination.Paginated[domainreview.AdminListItem]) AdminProductReviewListResponse {
	data := make([]AdminProductReviewListItemResponse, len(paginated.Data))
	for i, r := range paginated.Data {
		data[i] = ToAdminProductReviewListItemResponse(r)
	}
	return AdminProductReviewListResponse{
		Data: data,
		Meta: paginated.Meta,
	}
}

// ── Product Question Responses ───────────────────────────────────

type ProductQuestionResponse struct {
	ID         string     `json:"id"`
	ProductID  string     `json:"product_id"`
	AskerName  string     `json:"asker_name"`
	Question   string     `json:"question"`
	Answer     string     `json:"answer,omitempty"`
	AnsweredAt *time.Time `json:"answered_at,omitempty"`
	AnsweredBy *string    `json:"answered_by,omitempty"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
}

type ProductQuestionListResponse struct {
	Data []ProductQuestionResponse `json:"data"`
	Meta pagination.Meta           `json:"meta"`
}

type AdminProductQuestionListItemResponse struct {
	ProductQuestionResponse
	AskerEmail  string `json:"asker_email,omitempty"`
	ProductName string `json:"product_name"`
}

type AdminProductQuestionListResponse struct {
	Data []AdminProductQuestionListItemResponse `json:"data"`
	Meta pagination.Meta                        `json:"meta"`
}

func ToProductQuestionResponse(q domainquestion.Question) ProductQuestionResponse {
	var ansBy *string
	if q.AnsweredBy != nil {
		str := q.AnsweredBy.String()
		ansBy = &str
	}
	return ProductQuestionResponse{
		ID:         q.ID.String(),
		ProductID:  q.ProductID.String(),
		AskerName:  q.AskerName,
		Question:   q.Question,
		Answer:     q.Answer,
		AnsweredAt: q.AnsweredAt,
		AnsweredBy: ansBy,
		Status:     string(q.Status),
		CreatedAt:  q.CreatedAt,
	}
}

func ToProductQuestionListResponse(paginated pagination.Paginated[domainquestion.Question]) ProductQuestionListResponse {
	data := make([]ProductQuestionResponse, len(paginated.Data))
	for i, q := range paginated.Data {
		data[i] = ToProductQuestionResponse(q)
	}
	return ProductQuestionListResponse{
		Data: data,
		Meta: paginated.Meta,
	}
}

func ToAdminProductQuestionListItemResponse(item domainquestion.AdminListItem) AdminProductQuestionListItemResponse {
	resp := ToProductQuestionResponse(item.Question)
	return AdminProductQuestionListItemResponse{
		ProductQuestionResponse: resp,
		AskerEmail:              item.AskerEmail,
		ProductName:             item.ProductName,
	}
}

func ToAdminProductQuestionListResponse(paginated pagination.Paginated[domainquestion.AdminListItem]) AdminProductQuestionListResponse {
	data := make([]AdminProductQuestionListItemResponse, len(paginated.Data))
	for i, q := range paginated.Data {
		data[i] = ToAdminProductQuestionListItemResponse(q)
	}
	return AdminProductQuestionListResponse{
		Data: data,
		Meta: paginated.Meta,
	}
}

// ── Contact Message Responses ────────────────────────────────────

type ContactMessageResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone,omitempty"`
	Subject   string    `json:"subject,omitempty"`
	Message   string    `json:"message"`
	Source    string    `json:"source"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type ContactMessageListResponse struct {
	Data []ContactMessageResponse `json:"data"`
	Meta pagination.Meta          `json:"meta"`
}

func ToContactMessageResponse(m domaincontact.Message) ContactMessageResponse {
	return ContactMessageResponse{
		ID:        m.ID.String(),
		Name:      m.Name,
		Email:     m.Email,
		Phone:     m.Phone,
		Subject:   m.Subject,
		Message:   m.Message,
		Source:    string(m.Source),
		Status:    string(m.Status),
		CreatedAt: m.CreatedAt,
	}
}

func ToContactMessageListResponse(paginated pagination.Paginated[domaincontact.Message]) ContactMessageListResponse {
	data := make([]ContactMessageResponse, len(paginated.Data))
	for i, m := range paginated.Data {
		data[i] = ToContactMessageResponse(m)
	}
	return ContactMessageListResponse{
		Data: data,
		Meta: paginated.Meta,
	}
}
