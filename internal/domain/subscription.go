package domain

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID  `db:"id"`
	ServiceName string     `db:"service_name"`
	Price       int        `db:"price"`
	UserID      uuid.UUID  `db:"user_id"`
	StartDate   time.Time  `db:"start_date"`
	EndDate     *time.Time `db:"end_date"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

type UpdateSubscription struct {
	ServiceName *string    `db:"service_name"`
	Price       *int       `db:"price"`
	StartDate   *time.Time `db:"start_date"`
	EndDate     *time.Time `db:"end_date"`
}

type SubscriptionFilter struct {
	UserID      *uuid.UUID
	ServiceName *string
	Limit       int
	Offset      int
}

type CostRequest struct {
	UserID      *uuid.UUID
	ServiceName *string
}
