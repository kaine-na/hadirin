package notification

import (
	"encoding/json"
	"time"
)

// NotificationType mendefinisikan tipe-tipe notifikasi yang tersedia.
type NotificationType string

const (
	TypeClockInReminder  NotificationType = "clock_in_reminder"
	TypeClockInConfirmed NotificationType = "clock_in_confirmed"
	TypeLeaveApproved    NotificationType = "leave_approved"
	TypeLeaveRejected    NotificationType = "leave_rejected"
	TypeDocApproved      NotificationType = "doc_approved"
	TypeDocRejected      NotificationType = "doc_rejected"
	TypeLeaveReminder    NotificationType = "leave_reminder"    // reminder ke manager
	TypeDocReminder      NotificationType = "doc_reminder"      // reminder ke HR
)

// Notification merepresentasikan satu notifikasi di database.
type Notification struct {
	ID        string           `json:"id"`
	UserID    string           `json:"user_id"`
	Type      NotificationType `json:"type"`
	Title     string           `json:"title"`
	Message   string           `json:"message"`
	IsRead    bool             `json:"is_read"`
	Metadata  json.RawMessage  `json:"metadata,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
}

// CreateNotificationInput adalah input untuk membuat notifikasi baru.
type CreateNotificationInput struct {
	UserID   string
	Type     NotificationType
	Title    string
	Message  string
	Metadata map[string]interface{}
}

// ListFilter adalah filter untuk mengambil daftar notifikasi.
type ListFilter struct {
	UserID   string
	UnreadOnly bool
	Page     int
	PageSize int
}

// UnreadCountResponse adalah response untuk endpoint unread-count.
type UnreadCountResponse struct {
	Count int `json:"count"`
}
