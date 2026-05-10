package compliance

import (
	"time"
)

// Religion adalah agama karyawan untuk menentukan hari raya THR.
type Religion string

const (
	ReligionIslam    Religion = "islam"
	ReligionKristen  Religion = "kristen"
	ReligionKatolik  Religion = "katolik"
	ReligionHindu    Religion = "hindu"
	ReligionBuddha   Religion = "buddha"
	ReligionKonghucu Religion = "konghucu"
)

// HolidayInfo menyimpan informasi hari raya untuk THR.
type HolidayInfo struct {
	Religion    Religion
	Name        string
	Date        time.Time
	DeadlineH7  time.Time // H-7 sebelum hari raya
}

// GetHolidays mengembalikan daftar hari raya untuk tahun tertentu.
// Tanggal hari raya Islam (Lebaran) bersifat perkiraan dan harus diupdate tiap tahun.
func GetHolidays(year int) []HolidayInfo {
	holidays := []HolidayInfo{}

	// Lebaran (Idul Fitri) - perkiraan, berubah tiap tahun
	// 2026: sekitar 20 Maret 2026 (perkiraan)
	lebaranDates := map[int]time.Time{
		2024: time.Date(2024, 4, 10, 0, 0, 0, 0, time.UTC),
		2025: time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC),
		2026: time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
		2027: time.Date(2027, 3, 10, 0, 0, 0, 0, time.UTC),
	}
	if lebaran, ok := lebaranDates[year]; ok {
		holidays = append(holidays, HolidayInfo{
			Religion:   ReligionIslam,
			Name:       "Idul Fitri",
			Date:       lebaran,
			DeadlineH7: lebaran.AddDate(0, 0, -7),
		})
	}

	// Natal - 25 Desember (tetap)
	natal := time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC)
	holidays = append(holidays, HolidayInfo{
		Religion:   ReligionKristen,
		Name:       "Natal",
		Date:       natal,
		DeadlineH7: natal.AddDate(0, 0, -7),
	})
	holidays = append(holidays, HolidayInfo{
		Religion:   ReligionKatolik,
		Name:       "Natal",
		Date:       natal,
		DeadlineH7: natal.AddDate(0, 0, -7),
	})

	// Nyepi - berubah tiap tahun (kalender Saka)
	nyepiDates := map[int]time.Time{
		2024: time.Date(2024, 3, 11, 0, 0, 0, 0, time.UTC),
		2025: time.Date(2025, 3, 29, 0, 0, 0, 0, time.UTC),
		2026: time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC),
		2027: time.Date(2027, 3, 8, 0, 0, 0, 0, time.UTC),
	}
	if nyepi, ok := nyepiDates[year]; ok {
		holidays = append(holidays, HolidayInfo{
			Religion:   ReligionHindu,
			Name:       "Nyepi",
			Date:       nyepi,
			DeadlineH7: nyepi.AddDate(0, 0, -7),
		})
	}

	// Waisak - berubah tiap tahun
	waisakDates := map[int]time.Time{
		2024: time.Date(2024, 5, 23, 0, 0, 0, 0, time.UTC),
		2025: time.Date(2025, 5, 12, 0, 0, 0, 0, time.UTC),
		2026: time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC),
		2027: time.Date(2027, 5, 21, 0, 0, 0, 0, time.UTC),
	}
	if waisak, ok := waisakDates[year]; ok {
		holidays = append(holidays, HolidayInfo{
			Religion:   ReligionBuddha,
			Name:       "Waisak",
			Date:       waisak,
			DeadlineH7: waisak.AddDate(0, 0, -7),
		})
	}

	// Imlek - berubah tiap tahun
	imlekDates := map[int]time.Time{
		2024: time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC),
		2025: time.Date(2025, 1, 29, 0, 0, 0, 0, time.UTC),
		2026: time.Date(2026, 2, 17, 0, 0, 0, 0, time.UTC),
		2027: time.Date(2027, 2, 6, 0, 0, 0, 0, time.UTC),
	}
	if imlek, ok := imlekDates[year]; ok {
		holidays = append(holidays, HolidayInfo{
			Religion:   ReligionKonghucu,
			Name:       "Imlek",
			Date:       imlek,
			DeadlineH7: imlek.AddDate(0, 0, -7),
		})
	}

	return holidays
}

// GetHolidayForReligion mengembalikan hari raya untuk agama tertentu di tahun tertentu.
func GetHolidayForReligion(religion Religion, year int) *HolidayInfo {
	holidays := GetHolidays(year)
	for i, h := range holidays {
		if h.Religion == religion {
			return &holidays[i]
		}
	}
	return nil
}

// THRInput adalah input untuk kalkulasi THR.
type THRInput struct {
	BaseSalary      int64    // gaji pokok (bukan gaji bruto)
	ServiceMonths   int      // masa kerja dalam bulan
	Religion        Religion // agama karyawan
	Year            int      // tahun THR
}

// THRResult adalah hasil kalkulasi THR.
type THRResult struct {
	BaseSalary      int64    `json:"base_salary"`
	ServiceMonths   int      `json:"service_months"`
	Religion        Religion `json:"religion"`
	HolidayName     string   `json:"holiday_name"`
	HolidayDate     *time.Time `json:"holiday_date,omitempty"`
	DeadlineH7      *time.Time `json:"deadline_h7,omitempty"`
	DaysUntilH7     int      `json:"days_until_h7"`     // hari tersisa sampai deadline H-7
	DaysUntilHoliday int     `json:"days_until_holiday"` // hari tersisa sampai hari raya
	THRAmount       int64    `json:"thr_amount"`
	IsFullAmount    bool     `json:"is_full_amount"`    // true jika masa kerja >= 12 bulan
	ProRataRatio    float64  `json:"pro_rata_ratio"`    // rasio pro-rata (0-1)
	IsEligible      bool     `json:"is_eligible"`       // false jika masa kerja < 1 bulan
}

// CalculateTHR menghitung THR untuk satu karyawan.
func CalculateTHR(input THRInput) THRResult {
	result := THRResult{
		BaseSalary:    input.BaseSalary,
		ServiceMonths: input.ServiceMonths,
		Religion:      input.Religion,
	}

	// Cari info hari raya
	holiday := GetHolidayForReligion(input.Religion, input.Year)
	if holiday != nil {
		result.HolidayName = holiday.Name
		result.HolidayDate = &holiday.Date
		result.DeadlineH7 = &holiday.DeadlineH7

		// Hitung hari tersisa
		now := time.Now()
		if holiday.DeadlineH7.After(now) {
			result.DaysUntilH7 = int(holiday.DeadlineH7.Sub(now).Hours() / 24)
		}
		if holiday.Date.After(now) {
			result.DaysUntilHoliday = int(holiday.Date.Sub(now).Hours() / 24)
		}
	}

	// Karyawan dengan masa kerja < 1 bulan tidak berhak THR
	if input.ServiceMonths < 1 {
		result.IsEligible = false
		result.THRAmount = 0
		return result
	}

	result.IsEligible = true

	if input.ServiceMonths >= 12 {
		// Masa kerja >= 12 bulan: 1x gaji pokok
		result.IsFullAmount = true
		result.ProRataRatio = 1.0
		result.THRAmount = input.BaseSalary
	} else {
		// Masa kerja 1-11 bulan: pro-rata
		result.IsFullAmount = false
		result.ProRataRatio = float64(input.ServiceMonths) / 12.0
		result.THRAmount = roundToInt(float64(input.BaseSalary) * result.ProRataRatio)
	}

	return result
}

// CalculateTHRBatch menghitung THR untuk banyak karyawan sekaligus.
func CalculateTHRBatch(inputs []THRInput) []THRResult {
	results := make([]THRResult, len(inputs))
	for i, input := range inputs {
		results[i] = CalculateTHR(input)
	}
	return results
}
