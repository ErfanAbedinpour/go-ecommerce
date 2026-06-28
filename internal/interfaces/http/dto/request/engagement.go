package request

// WishlistAddRequest is the request payload to add an item to wishlist.
type WishlistAddRequest struct {
	ProductID string `json:"product_id" validate:"required,uuid"`
}

// ProductReviewSubmitRequest is the storefront request payload to submit a review.
type ProductReviewSubmitRequest struct {
	AuthorName string `json:"author_name" validate:"required,max=255"`
	Rating     int    `json:"rating" validate:"required,min=1,max=5"`
	Title      string `json:"title" validate:"max=255"`
	Content    string `json:"content" validate:"required"`
}

// ProductReviewStatusUpdateRequest is the admin request to update review status.
type ProductReviewStatusUpdateRequest struct {
	Status string `json:"status" validate:"required,oneof=pending approved rejected"`
}

// ProductQuestionAskRequest is the storefront request to ask a question.
type ProductQuestionAskRequest struct {
	AskerName  string `json:"asker_name" validate:"required,max=255"`
	AskerEmail string `json:"asker_email" validate:"omitempty,email,max=255"`
	Question   string `json:"question" validate:"required"`
}

// ProductQuestionAnswerRequest is the admin request to answer a question.
type ProductQuestionAnswerRequest struct {
	Answer string `json:"answer" validate:"required"`
}

// ContactMessageSubmitRequest is the storefront request to submit a contact form.
type ContactMessageSubmitRequest struct {
	Name    string `json:"name" validate:"required,max=255"`
	Email   string `json:"email" validate:"required,email,max=255"`
	Phone   string `json:"phone" validate:"omitempty,max=50"`
	Subject string `json:"subject" validate:"omitempty,max=500"`
	Message string `json:"message" validate:"required"`
	Source  string `json:"source" validate:"omitempty,oneof=homepage about contact_page"`
}

// ContactMessageStatusUpdateRequest is the admin request to update contact message status.
type ContactMessageStatusUpdateRequest struct {
	Status string `json:"status" validate:"required,oneof=unread read archived"`
}
