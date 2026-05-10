package attendance

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NotificationSender adalah interface untuk mengirim notifikasi.
type NotificationSender interface {
	SendNotification(ctx context.Context, userID, notifType, title, message string, metadata map[string]interface{}) error
}

// Service menangani logika bisnis untuk modul attendance.
type Service struct {
	repo      *Repository
	notifSvc  NotificationSender
}

// NewService membuat instance Service baru.
func NewService(db *pgxpool.Pool) *Service {
	return &Service{repo: NewRepository(db)}
}

// SetNotificationService menginjeksi notification service.
func (s *Service) SetNotificationService(ns NotificationSender) {
	s.notifSvc = ns
}

// ClockIn melakukan clock-in untuk user.
// Mengembalikan error jika sudah clock-in hari ini.
func (s *Service) ClockIn(ctx context.Context, userID, ipAddress, notes string) (*Attendance, error) {
	// Cek apakah sudah clock-in hari ini
	existing, err := s.repo.FindTodayByUserID(ctx, userID)
	if err == nil && existing != nil {
		return nil, errors.New("sudah melakukan clock-in hari ini")
	}

	att, err := s.repo.ClockIn(ctx, userID, ipAddress, notes)
	if err != nil {
		return nil, err
	}

	// Kirim konfirmasi clock-in
	if s.notifSvc != nil {
		_ = s.notifSvc.SendNotification(ctx, userID,
			"clock_in_confirmed",
			"Clock-in berhasil",
			fmt.Sprintf("Anda berhasil clock-in pada %s.", att.ClockIn.Format("15:04 WIB")),
			map[string]interface{}{"attendance_id": att.ID},
		)
	}

	return att, nil
}

// ClockOut melakukan clock-out untuk user.
// Mengembalikan error jika belum clock-in atau sudah clock-out.
func (s *Service) ClockOut(ctx context.Context, userID, notes string) (*Attendance, error) {
	// Cek absensi hari ini
	today, err := s.repo.FindTodayByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("belum melakukan clock-in hari ini")
	}

	if today.ClockOut != nil {
		return nil, errors.New("sudah melakukan clock-out hari ini")
	}

	return s.repo.ClockOut(ctx, today.ID, notes)
}

// GetToday mengambil status absensi hari ini untuk user.
func (s *Service) GetToday(ctx context.Context, userID string) (*Attendance, error) {
	att, err := s.repo.FindTodayByUserID(ctx, userID)
	if err != nil {
		// Belum ada absensi hari ini, kembalikan nil tanpa error
		return nil, nil
	}
	return att, nil
}

// GetByEmployee mengambil rekap absensi untuk satu karyawan.
func (s *Service) GetByEmployee(ctx context.Context, userID string, filter RekapFilter) ([]*Attendance, int, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 31 // Default satu bulan
	}
	filter.UserID = userID
	return s.repo.List(ctx, filter)
}

// Override mengupdate absensi oleh HR.
func (s *Service) Override(ctx context.Context, attendanceID, updatedBy string, req *OverrideRequest) (*Attendance, error) {
	// Validasi status
	if req.Status != "" {
		validStatuses := map[string]bool{
			"hadir": true, "terlambat": true, "izin": true, "sakit": true, "alpha": true,
		}
		if !validStatuses[req.Status] {
			return nil, fmt.Errorf("status tidak valid: %s", req.Status)
		}
	}

	return s.repo.Override(ctx, attendanceID, updatedBy, req)
}

// ExportCSV menghasilkan CSV dari data absensi.
func (s *Service) ExportCSV(ctx context.Context, startDate, endDate string) ([]byte, error) {
	attendances, err := s.repo.ListForExport(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("get data for export: %w", err)
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Header CSV
	if err := writer.Write([]string{
		"ID", "User ID", "Tanggal", "Clock In", "Clock Out", "Status", "Catatan", "IP Address",
	}); err != nil {
		return nil, fmt.Errorf("write csv header: %w", err)
	}

	// Data rows
	for _, att := range attendances {
		clockIn := ""
		if att.ClockIn != nil {
			clockIn = att.ClockIn.Format(time.RFC3339)
		}
		clockOut := ""
		if att.ClockOut != nil {
			clockOut = att.ClockOut.Format(time.RFC3339)
		}

		if err := writer.Write([]string{
			att.ID,
			att.UserID,
			att.Date.Format("2006-01-02"),
			clockIn,
			clockOut,
			att.Status,
			att.Notes,
			att.IPAddress,
		}); err != nil {
			return nil, fmt.Errorf("write csv row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("flush csv: %w", err)
	}

	return buf.Bytes(), nil
}
