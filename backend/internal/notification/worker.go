package notification

import (
	"context"
	"log"
	"time"
)

// Worker adalah background goroutine yang mengirim reminder otomatis.
type Worker struct {
	svc  *Service
	repo *Repository
}

// NewWorker membuat instance Worker baru.
func NewWorker(svc *Service) *Worker {
	return &Worker{
		svc:  svc,
		repo: svc.repo,
	}
}

// Start menjalankan background worker. Dipanggil dengan go worker.Start(ctx).
// Worker berhenti saat ctx di-cancel (misal saat server shutdown).
func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Println("[NotificationWorker] dimulai")

	for {
		select {
		case <-ctx.Done():
			log.Println("[NotificationWorker] berhenti")
			return
		case t := <-ticker.C:
			w.runChecks(ctx, t)
		}
	}
}

// runChecks menjalankan semua pengecekan reminder pada satu tick.
func (w *Worker) runChecks(ctx context.Context, now time.Time) {
	// Konversi ke WIB (UTC+7)
	wib := time.FixedZone("WIB", 7*60*60)
	localNow := now.In(wib)

	// Reminder clock-in: kirim jam 08:15 WIB
	if localNow.Hour() == 8 && localNow.Minute() == 15 {
		w.sendClockInReminders(ctx)
	}

	// Cek pending leave > 2 hari (jalankan setiap jam 09:00)
	if localNow.Hour() == 9 && localNow.Minute() == 0 {
		w.sendLeaveReminders(ctx)
	}

	// Cek pending dokumen > 3 hari (jalankan setiap jam 09:05)
	if localNow.Hour() == 9 && localNow.Minute() == 5 {
		w.sendDocReminders(ctx)
	}
}

// sendClockInReminders mengirim reminder ke karyawan yang belum clock-in.
func (w *Worker) sendClockInReminders(ctx context.Context) {
	userIDs, err := w.repo.GetUsersNotClockedIn(ctx)
	if err != nil {
		log.Printf("[NotificationWorker] gagal ambil users belum clock-in: %v", err)
		return
	}

	for _, userID := range userIDs {
		// Cek apakah reminder sudah dikirim hari ini
		sent, err := w.repo.HasReminderSentToday(ctx, userID, TypeClockInReminder)
		if err != nil || sent {
			continue
		}

		_, err = w.svc.Send(ctx, &CreateNotificationInput{
			UserID:  userID,
			Type:    TypeClockInReminder,
			Title:   "Jangan lupa clock-in!",
			Message: "Anda belum melakukan clock-in hari ini. Segera lakukan absensi.",
			Metadata: map[string]interface{}{
				"action": "clock_in",
			},
		})
		if err != nil {
			log.Printf("[NotificationWorker] gagal kirim clock-in reminder ke %s: %v", userID, err)
		}
	}

	log.Printf("[NotificationWorker] clock-in reminder dikirim ke %d karyawan", len(userIDs))
}

// sendLeaveReminders mengirim reminder ke manager untuk leave pending > 2 hari.
func (w *Worker) sendLeaveReminders(ctx context.Context) {
	leaves, err := w.repo.GetPendingLeavesOlderThan(ctx, 2)
	if err != nil {
		log.Printf("[NotificationWorker] gagal ambil pending leaves: %v", err)
		return
	}

	sent := 0
	for _, item := range leaves {
		managerID := item["manager_id"]
		leaveID := item["leave_id"]

		// Cek apakah reminder sudah dikirim hari ini untuk leave ini
		alreadySent, err := w.repo.HasReminderSentToday(ctx, managerID, TypeLeaveReminder)
		if err != nil || alreadySent {
			continue
		}

		_, err = w.svc.Send(ctx, &CreateNotificationInput{
			UserID:  managerID,
			Type:    TypeLeaveReminder,
			Title:   "Pengajuan cuti menunggu persetujuan",
			Message: "Ada pengajuan cuti yang sudah menunggu lebih dari 2 hari. Segera tinjau.",
			Metadata: map[string]interface{}{
				"leave_id": leaveID,
				"action":   "review_leave",
			},
		})
		if err != nil {
			log.Printf("[NotificationWorker] gagal kirim leave reminder ke manager %s: %v", managerID, err)
			continue
		}
		sent++
	}

	if sent > 0 {
		log.Printf("[NotificationWorker] leave reminder dikirim ke %d manager", sent)
	}
}

// sendDocReminders mengirim reminder ke HR untuk dokumen pending > 3 hari.
func (w *Worker) sendDocReminders(ctx context.Context) {
	docs, err := w.repo.GetPendingDocsOlderThan(ctx, 3)
	if err != nil {
		log.Printf("[NotificationWorker] gagal ambil pending docs: %v", err)
		return
	}

	sent := 0
	for _, item := range docs {
		hrID := item["hr_id"]
		docID := item["doc_id"]

		alreadySent, err := w.repo.HasReminderSentToday(ctx, hrID, TypeDocReminder)
		if err != nil || alreadySent {
			continue
		}

		_, err = w.svc.Send(ctx, &CreateNotificationInput{
			UserID:  hrID,
			Type:    TypeDocReminder,
			Title:   "Dokumen menunggu review",
			Message: "Ada dokumen yang sudah menunggu review lebih dari 3 hari. Segera tinjau.",
			Metadata: map[string]interface{}{
				"doc_id": docID,
				"action": "review_doc",
			},
		})
		if err != nil {
			log.Printf("[NotificationWorker] gagal kirim doc reminder ke HR %s: %v", hrID, err)
			continue
		}
		sent++
	}

	if sent > 0 {
		log.Printf("[NotificationWorker] doc reminder dikirim ke %d HR", sent)
	}
}
