package notification

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Service menangani logika bisnis untuk modul notification.
type Service struct {
	repo    *Repository
	hub     *SSEHub
}

// NewService membuat instance Service baru.
func NewService(db *pgxpool.Pool) *Service {
	return &Service{
		repo: NewRepository(db),
		hub:  NewSSEHub(),
	}
}

// GetHub mengembalikan SSEHub untuk digunakan oleh handler.
func (s *Service) GetHub() *SSEHub {
	return s.hub
}

// Send membuat notifikasi baru dan mem-push ke SSE client yang aktif.
func (s *Service) Send(ctx context.Context, input *CreateNotificationInput) (*Notification, error) {
	if input.UserID == "" {
		return nil, fmt.Errorf("user_id wajib diisi")
	}
	if input.Title == "" {
		return nil, fmt.Errorf("title wajib diisi")
	}
	if input.Message == "" {
		return nil, fmt.Errorf("message wajib diisi")
	}

	n, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	// Push ke SSE client yang sedang terhubung (non-blocking)
	s.hub.Broadcast(input.UserID, n)

	return n, nil
}

// List mengambil daftar notifikasi untuk user.
func (s *Service) List(ctx context.Context, userID string, filter ListFilter) ([]*Notification, int, error) {
	filter.UserID = userID
	return s.repo.List(ctx, filter)
}

// MarkAsRead menandai satu notifikasi sebagai sudah dibaca.
func (s *Service) MarkAsRead(ctx context.Context, id, userID string) error {
	return s.repo.MarkAsRead(ctx, id, userID)
}

// MarkAllAsRead menandai semua notifikasi user sebagai sudah dibaca.
func (s *Service) MarkAllAsRead(ctx context.Context, userID string) error {
	return s.repo.MarkAllAsRead(ctx, userID)
}

// GetUnreadCount mengambil jumlah notifikasi yang belum dibaca.
func (s *Service) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	return s.repo.GetUnreadCount(ctx, userID)
}

// SendNotification adalah adapter yang memenuhi interface NotificationSender
// yang digunakan oleh leave/service.go dan attendance/service.go.
func (s *Service) SendNotification(ctx context.Context, userID, notifType, title, message string, metadata map[string]interface{}) error {
	_, err := s.Send(ctx, &CreateNotificationInput{
		UserID:   userID,
		Type:     NotificationType(notifType),
		Title:    title,
		Message:  message,
		Metadata: metadata,
	})
	return err
}
