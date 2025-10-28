package dto

// SubscriptionOutput represents a subscription response.
// @Description SubscriptionOutput
type SubscriptionOutput struct {
	ID          string  `json:"id" example:"a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8"`
	ServiceName string  `json:"service_name" example:"Yandex Plus"`
	Price       int     `json:"price" example:"400"`
	UserID      string  `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string  `json:"start_date" example:"07-2025"`
	EndDate     *string `json:"end_date" example:"12-2025"`
	CreatedAt   string  `json:"created_at" example:"2025-04-05T10:00:00Z"`
	UpdatedAt   string  `json:"updated_at" example:"2025-04-05T10:00:00Z"`
}

// CreateSubscriptionRequest represents the request to create a subscription.
// @Description CreateSubscriptionRequest
type CreateSubscriptionRequest struct {
	ServiceName string  `json:"service_name" validate:"required" example:"Yandex Plus"`
	Price       int     `json:"price" validate:"required,min=0" example:"400"`
	UserID      string  `json:"user_id" validate:"required" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string  `json:"start_date" validate:"required" example:"07-2025"`
	EndDate     *string `json:"end_date" example:"12-2025"`
}

// UpdateSubscriptionRequest represents partial update fields.
// @Description UpdateSubscriptionRequest
type UpdateSubscriptionRequest struct {
	ServiceName *string `json:"service_name" example:"Spotify Premium"`
	Price       *int    `json:"price" validate:"omitempty,min=0" example:"599"`
	StartDate   *string `json:"start_date" example:"01-2026"`
	EndDate     *string `json:"end_date" example:"06-2026"`
}

type SubscriptionFilter struct {
	UserID      *string `form:"user_id"`
	ServiceName *string `form:"service_name"`
	Page        int     `form:"page"`
	PageSize    int     `form:"page_size"`
}

type CostRequest struct {
	UserID      *string `form:"user_id"`
	ServiceName *string `form:"service_name"`
	From        *string `form:"from"`
	To          *string `form:"to"`
}

// SubscriptionsOutput paginated response.
// @Description SubscriptionsOutput
type SubscriptionsOutput struct {
	Total         int                  `json:"total" example:"42"`
	Page          int                  `json:"page" example:"1"`
	PageSize      int                  `json:"page_size" example:"10"`
	HasNextPage   bool                 `json:"has_next_page" example:"true"`
	HasPrevPage   bool                 `json:"has_prev_page" example:"false"`
	Subscriptions []SubscriptionOutput `json:"subscriptions"`
}

func MakeSubscriptionsOutput(subscriptions []SubscriptionOutput, total int, page int, pageSize int) SubscriptionsOutput {
	return SubscriptionsOutput{
		Total:         total,
		Page:          page,
		PageSize:      pageSize,
		HasNextPage:   total > page*pageSize,
		HasPrevPage:   page > 1,
		Subscriptions: subscriptions,
	}
}
