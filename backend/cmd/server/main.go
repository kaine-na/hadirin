package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"saas-karyawan/internal/ai"
	"saas-karyawan/internal/analytics"
	"saas-karyawan/internal/attendance"
	"saas-karyawan/internal/auth"
	"saas-karyawan/internal/compliance"
	"saas-karyawan/internal/database"
	"saas-karyawan/internal/document"
	"saas-karyawan/internal/employee"
	"saas-karyawan/internal/fraud"
	"saas-karyawan/internal/leave"
	"saas-karyawan/internal/notification"
	"saas-karyawan/pkg/config"
)

func main() {
	// Load konfigurasi
	cfg := config.Load()

	// Setup database connection
	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Gagal koneksi ke database: %v", err)
	}
	defer pool.Close()
	log.Println("Database terhubung")

	// Jalankan migrations
	migrationsDir := filepath.Join(".", "migrations")
	if err := database.RunMigrations(ctx, pool, migrationsDir); err != nil {
		log.Fatalf("Gagal menjalankan migrations: %v", err)
	}
	log.Println("Migrations selesai")

	// Buat direktori uploads jika belum ada
	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		log.Fatalf("Gagal membuat direktori uploads: %v", err)
	}

	// Inisialisasi services
	authSvc := auth.NewService(pool, cfg.JWTSecret, cfg.JWTExpiryHours)
	employeeSvc := employee.NewService(pool, cfg.UploadDir)
	attendanceSvc := attendance.NewService(pool)
	documentSvc := document.NewService(pool, cfg.UploadDir, cfg.MaxFileSizeMB)

	llmClient := ai.NewLLMClient(cfg.LLMBaseURL, cfg.LLMAPIKey, cfg.LLMModel, cfg.LLMTimeoutSeconds)
	aiSvc := ai.NewService(pool, llmClient, cfg.LLMModel)

	leaveSvc := leave.NewService(pool, llmClient)

	notificationSvc := notification.NewService(pool)
	notificationHandler := notification.NewHandler(notificationSvc)
	notificationWorker := notification.NewWorker(notificationSvc)

	// Inject notification service ke leave service (setelah keduanya dibuat)
	leaveSvc.SetNotificationService(notificationSvc)
	attendanceSvc.SetNotificationService(notificationSvc)

	// Inisialisasi analytics service dan handler
	analyticsSvc := analytics.NewService(pool, llmClient, cfg.LLMModel)
	analyticsHandler := analytics.NewHandler(analyticsSvc)

	// Inisialisasi compliance handler
	complianceHandler := compliance.NewHandler(pool, notificationSvc)

	// Inisialisasi fraud detection
	fraudRepo := fraud.NewRepository(pool)
	fraudGPSValidator := fraud.NewGPSValidator(pool)
	fraudLiveness := fraud.NewLivenessChecker(pool, cfg.UploadDir, llmClient)
	fraudAnomaly := fraud.NewAnomalyDetector(pool, llmClient)
	fraudHandler := fraud.NewHandler(fraudRepo, fraudGPSValidator, fraudLiveness, fraudAnomaly, cfg.UploadDir)

	// Inisialisasi handlers
	authHandler := auth.NewHandler(authSvc)
	loginRateLimiter := auth.NewLoginRateLimiter()
	employeeHandler := employee.NewHandler(employeeSvc)
	attendanceHandler := attendance.NewHandler(attendanceSvc)
	documentHandler := document.NewHandler(documentSvc)
	aiHandler := ai.NewHandler(aiSvc)
	leaveHandler := leave.NewHandler(leaveSvc)

	// Setup router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(corsMiddleware(cfg.CORSOrigins))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","time":"%s"}`, time.Now().Format(time.RFC3339))
	})

	// Serve uploaded files (dengan path yang aman)
	r.Handle("/uploads/*", http.StripPrefix("/uploads/",
		http.FileServer(http.Dir(cfg.UploadDir)),
	))

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Auth routes (public)
		r.Route("/auth", func(r chi.Router) {
			r.With(loginRateLimiter.Middleware).Post("/login", authHandler.Login)
			r.Post("/logout", authHandler.Logout)
			r.With(auth.RequireAuth(authSvc)).Get("/me", authHandler.Me)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireAuth(authSvc))

			// Employee routes
			r.Route("/employees", func(r chi.Router) {
				r.With(auth.RequireRole("super_admin", "hr_admin", "manager")).Get("/", employeeHandler.List)
				r.With(auth.RequireRole("super_admin", "hr_admin")).Post("/", employeeHandler.Create)
				r.Get("/{id}", employeeHandler.GetByID)
				r.With(auth.RequireRole("super_admin", "hr_admin")).Put("/{id}", employeeHandler.Update)
				r.With(auth.RequireRole("super_admin", "hr_admin")).Delete("/{id}", employeeHandler.Delete)
				r.Post("/{id}/photo", employeeHandler.UploadPhoto)
			})

			// Attendance routes
			r.Route("/attendance", func(r chi.Router) {
				r.Post("/clock-in", attendanceHandler.ClockIn)
				r.Post("/clock-out", attendanceHandler.ClockOut)
				r.Get("/me", attendanceHandler.GetMe)
				r.Get("/today", attendanceHandler.GetToday)
				r.With(auth.RequireRole("super_admin", "hr_admin", "manager")).Get("/{employee_id}", attendanceHandler.GetByEmployee)
				r.With(auth.RequireRole("super_admin", "hr_admin")).Put("/{id}", attendanceHandler.Override)
				r.With(auth.RequireRole("super_admin", "hr_admin")).Get("/export/csv", attendanceHandler.ExportCSV)
			})

			// Document routes
			r.Route("/documents", func(r chi.Router) {
				r.Post("/upload", documentHandler.Upload)
				r.Get("/", documentHandler.List)
				r.Get("/{id}", documentHandler.GetByID)
				r.Delete("/{id}", documentHandler.Delete)
				r.Get("/{id}/download", documentHandler.Download)
				r.With(auth.RequireRole("super_admin", "hr_admin", "manager")).Post("/{id}/comments", documentHandler.AddComment)
				r.Get("/{id}/comments", documentHandler.ListComments)
			})

			// AI routes (HR only)
			r.Route("/ai", func(r chi.Router) {
				r.Use(auth.RequireRole("super_admin", "hr_admin", "manager"))
				r.Post("/analyze/{employee_id}", aiHandler.Analyze)
				r.Get("/reports/{employee_id}", aiHandler.GetReports)
				r.Get("/report/{id}", aiHandler.GetReportByID)
			})

			// Leave routes
			r.Route("/leaves", func(r chi.Router) {
				r.Get("/types", leaveHandler.GetLeaveTypes)
				r.Post("/", leaveHandler.Create)
				r.Get("/", leaveHandler.List)
				r.Get("/balance", leaveHandler.GetMyBalance)
				r.With(auth.RequireRole("super_admin", "hr_admin", "manager")).Get("/balance/{user_id}", leaveHandler.GetBalanceByUserID)
				r.Get("/{id}", leaveHandler.GetByID)
				r.With(auth.RequireRole("super_admin", "hr_admin", "manager")).Put("/{id}/approve", leaveHandler.Approve)
				r.With(auth.RequireRole("super_admin", "hr_admin", "manager")).Put("/{id}/reject", leaveHandler.Reject)
				r.Put("/{id}/cancel", leaveHandler.Cancel)
				r.With(auth.RequireRole("super_admin", "hr_admin", "manager")).Get("/{id}/ai-recommendation", leaveHandler.GetAIRecommendation)
			})

			// Notification routes
				r.Route("/notifications", func(r chi.Router) {
					r.Get("/stream", notificationHandler.Stream)
					r.Get("/", notificationHandler.List)
					r.Get("/unread-count", notificationHandler.GetUnreadCount)
					r.Put("/read-all", notificationHandler.MarkAllAsRead)
					r.Put("/{id}/read", notificationHandler.MarkAsRead)
				})

				// Analytics routes (HR/Manager only)
				r.Route("/analytics", func(r chi.Router) {
					r.Use(auth.RequireRole("super_admin", "hr_admin", "manager"))
					r.Get("/attendance-summary", analyticsHandler.GetAttendanceSummary)
					r.Get("/department-stats", analyticsHandler.GetDepartmentStats)
					r.Get("/trend", analyticsHandler.GetTrend)
					r.Get("/top-late-employees", analyticsHandler.GetTopLateEmployees)
					r.Get("/executive-summary", analyticsHandler.GetExecutiveSummary)
				})

				// Reports routes (HR/Manager only)
				r.Route("/reports", func(r chi.Router) {
					r.Use(auth.RequireRole("super_admin", "hr_admin", "manager"))
					r.Get("/export-pdf", analyticsHandler.ExportPDF)
				})

				// Compliance routes (HR Admin only)
				r.Route("/compliance", func(r chi.Router) {
					r.Use(auth.RequireRole("super_admin", "hr_admin"))
					r.Get("/bpjs-calculation", complianceHandler.GetBPJSCalculation)
					r.Get("/pph21-calculation", complianceHandler.GetPPh21Calculation)
					r.Get("/thr-calculation", complianceHandler.GetTHRCalculation)
					r.Get("/checklist", complianceHandler.GetChecklist)
					r.Put("/checklist/{id}/done", complianceHandler.MarkChecklistDone)
					r.Get("/summary", complianceHandler.GetSummary)
				})

				// Fraud detection routes
				r.Route("/fraud", func(r chi.Router) {
					// Validasi clock-in (semua user terautentikasi)
					r.Post("/validate-clock-in", fraudHandler.ValidateClockIn)

					// Fraud logs (HR Admin only)
					r.With(auth.RequireRole("super_admin", "hr_admin")).Get("/logs", fraudHandler.ListFraudLogs)
					r.With(auth.RequireRole("super_admin", "hr_admin")).Get("/logs/{id}", fraudHandler.GetFraudLogByID)
					r.With(auth.RequireRole("super_admin", "hr_admin")).Put("/logs/{id}/dismiss", fraudHandler.DismissFraudLog)
					r.With(auth.RequireRole("super_admin", "hr_admin")).Put("/logs/{id}/confirm", fraudHandler.ConfirmFraudLog)
					r.With(auth.RequireRole("super_admin", "hr_admin")).Get("/summary", fraudHandler.GetFraudSummary)
				})
		})
	})

	// Jalankan background worker di goroutine terpisah
	workerCtx, cancelWorker := context.WithCancel(ctx)
	defer cancelWorker()
	go notificationWorker.Start(workerCtx)

	addr := ":" + cfg.Port
	log.Printf("Server berjalan di http://localhost%s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// corsMiddleware menangani CORS headers.
func corsMiddleware(allowedOrigins string) func(http.Handler) http.Handler {
	origins := strings.Split(allowedOrigins, ",")
	originMap := make(map[string]bool)
	for _, o := range origins {
		originMap[strings.TrimSpace(o)] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if originMap[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
