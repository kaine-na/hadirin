package attendance

import "time"

// Attendance adalah representasi data absensi dari database.
type Attendance struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Date      time.Time  `json:"date"`
	ClockIn   *time.Time `json:"clock_in,omitempty"`
	ClockOut  *time.Time `json:"clock_out,omitempty"`
	Status    string     `json:"status"`
	Notes     string     `json:"notes,omitempty"`
	IPAddress string     `json:"ip_address,omitempty"`
	CreatedBy string     `json:"created_by,omitempty"`
	UpdatedBy string     `json:"updated_by,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// ClockInRequest adalah body request untuk clock-in.
type ClockInRequest struct {
	Notes string `json:"notes"`
}

// ClockOutRequest adalah body request untuk clock-out.
type ClockOutRequest struct {
	Notes string `json:"notes"`
}

// OverrideRequest adalah body request untuk HR override absensi.
type OverrideRequest struct {
	Status   string `json:"status"`
	Notes    string `json:"notes"`
	ClockIn  string `json:"clock_in"`  // ISO 8601
	ClockOut string `json:"clock_out"` // ISO 8601
}

// RekapFilter adalah parameter filter untuk rekap absensi.
type RekapFilter struct {
	UserID    string
	StartDate string
	EndDate   string
	Page      int
	PageSize  int
}

// DailySummary adalah ringkasan absensi per hari untuk rekap.
type DailySummary struct {
	Date     string `json:"date"`
	Status   string `json:"status"`
	ClockIn  string `json:"clock_in,omitempty"`
	ClockOut string `json:"clock_out,omitempty"`
	Notes    string `json:"notes,omitempty"`
}
