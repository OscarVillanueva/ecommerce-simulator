package dao

import "time"

type Ticket struct {
	TicketId string `json:"ticket_id"`
	Total float32 `json:"total"`
	CreatedAt time.Time `json:"created_at"`
}
