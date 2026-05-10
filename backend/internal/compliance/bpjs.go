package compliance

import (
	"math"
	"time"
)

// BPJSRates menyimpan tarif BPJS yang berlaku.
type BPJSRates struct {
	// BPJS Kesehatan
	KesCompanyRate  float64 // 4%
	KesEmployeeRate float64 // 1%
	KesMaxSalary    int64   // Rp 12.000.000

	// BPJS TK - JHT (Jaminan Hari Tua)
	JHTCompanyRate  float64 // 3.7%
	JHTEmployeeRate float64 // 2%

	// BPJS TK - JP (Jaminan Pensiun)
	JPCompanyRate  float64 // 2%
	JPEmployeeRate float64 // 1%
	JPMaxSalary    int64   // Rp 9.559.600

	// BPJS TK - JKK (Jaminan Kecelakaan Kerja) - risiko standar
	JKKCompanyRate float64 // 0.24%

	// BPJS TK - JKM (Jaminan Kematian)
	JKMCompanyRate float64 // 0.3%
}

// DefaultBPJSRates mengembalikan tarif BPJS yang berlaku per 2026.
func DefaultBPJSRates() BPJSRates {
	return BPJSRates{
		KesCompanyRate:  0.04,
		KesEmployeeRate: 0.01,
		KesMaxSalary:    12_000_000,

		JHTCompanyRate:  0.037,
		JHTEmployeeRate: 0.02,

		JPCompanyRate:  0.02,
		JPEmployeeRate: 0.01,
		JPMaxSalary:    9_559_600,

		JKKCompanyRate: 0.0024,
		JKMCompanyRate: 0.003,
	}
}

// BPJSResult menyimpan hasil kalkulasi BPJS untuk satu karyawan.
type BPJSResult struct {
	GrossSalary int64 `json:"gross_salary"`

	// BPJS Kesehatan
	KesBaseSalary   int64   `json:"kes_base_salary"`   // gaji yang dipakai (max 12jt)
	KesCompany      int64   `json:"kes_company"`       // tanggungan perusahaan
	KesEmployee     int64   `json:"kes_employee"`      // tanggungan karyawan
	KesTotal        int64   `json:"kes_total"`         // total iuran
	KesCompanyRate  float64 `json:"kes_company_rate"`
	KesEmployeeRate float64 `json:"kes_employee_rate"`

	// BPJS TK - JHT
	JHTCompany      int64   `json:"jht_company"`
	JHTEmployee     int64   `json:"jht_employee"`
	JHTTotal        int64   `json:"jht_total"`
	JHTCompanyRate  float64 `json:"jht_company_rate"`
	JHTEmployeeRate float64 `json:"jht_employee_rate"`

	// BPJS TK - JP
	JPBaseSalary   int64   `json:"jp_base_salary"`   // gaji yang dipakai (max 9.559.600)
	JPCompany      int64   `json:"jp_company"`
	JPEmployee     int64   `json:"jp_employee"`
	JPTotal        int64   `json:"jp_total"`
	JPCompanyRate  float64 `json:"jp_company_rate"`
	JPEmployeeRate float64 `json:"jp_employee_rate"`

	// BPJS TK - JKK
	JKKCompany     int64   `json:"jkk_company"`
	JKKCompanyRate float64 `json:"jkk_company_rate"`

	// BPJS TK - JKM
	JKMCompany     int64   `json:"jkm_company"`
	JKMCompanyRate float64 `json:"jkm_company_rate"`

	// Ringkasan
	TotalCompanyContribution  int64 `json:"total_company_contribution"`
	TotalEmployeeContribution int64 `json:"total_employee_contribution"`
	TotalContribution         int64 `json:"total_contribution"`
	TakeHomePay               int64 `json:"take_home_pay"` // gaji - potongan karyawan
}

// CalculateBPJS menghitung iuran BPJS untuk satu karyawan.
// grossSalary adalah gaji bruto bulanan dalam rupiah.
func CalculateBPJS(grossSalary int64, rates BPJSRates) BPJSResult {
	result := BPJSResult{
		GrossSalary:     grossSalary,
		KesCompanyRate:  rates.KesCompanyRate,
		KesEmployeeRate: rates.KesEmployeeRate,
		JHTCompanyRate:  rates.JHTCompanyRate,
		JHTEmployeeRate: rates.JHTEmployeeRate,
		JPCompanyRate:   rates.JPCompanyRate,
		JPEmployeeRate:  rates.JPEmployeeRate,
		JKKCompanyRate:  rates.JKKCompanyRate,
		JKMCompanyRate:  rates.JKMCompanyRate,
	}

	// BPJS Kesehatan: gaji dibatasi max Rp 12 juta
	kesSalary := grossSalary
	if kesSalary > rates.KesMaxSalary {
		kesSalary = rates.KesMaxSalary
	}
	result.KesBaseSalary = kesSalary
	result.KesCompany = roundToInt(float64(kesSalary) * rates.KesCompanyRate)
	result.KesEmployee = roundToInt(float64(kesSalary) * rates.KesEmployeeRate)
	result.KesTotal = result.KesCompany + result.KesEmployee

	// BPJS TK - JHT: tidak ada batas gaji
	result.JHTCompany = roundToInt(float64(grossSalary) * rates.JHTCompanyRate)
	result.JHTEmployee = roundToInt(float64(grossSalary) * rates.JHTEmployeeRate)
	result.JHTTotal = result.JHTCompany + result.JHTEmployee

	// BPJS TK - JP: gaji dibatasi max Rp 9.559.600
	jpSalary := grossSalary
	if jpSalary > rates.JPMaxSalary {
		jpSalary = rates.JPMaxSalary
	}
	result.JPBaseSalary = jpSalary
	result.JPCompany = roundToInt(float64(jpSalary) * rates.JPCompanyRate)
	result.JPEmployee = roundToInt(float64(jpSalary) * rates.JPEmployeeRate)
	result.JPTotal = result.JPCompany + result.JPEmployee

	// BPJS TK - JKK: hanya perusahaan
	result.JKKCompany = roundToInt(float64(grossSalary) * rates.JKKCompanyRate)

	// BPJS TK - JKM: hanya perusahaan
	result.JKMCompany = roundToInt(float64(grossSalary) * rates.JKMCompanyRate)

	// Ringkasan
	result.TotalCompanyContribution = result.KesCompany + result.JHTCompany +
		result.JPCompany + result.JKKCompany + result.JKMCompany
	result.TotalEmployeeContribution = result.KesEmployee + result.JHTEmployee + result.JPEmployee
	result.TotalContribution = result.TotalCompanyContribution + result.TotalEmployeeContribution
	result.TakeHomePay = grossSalary - result.TotalEmployeeContribution

	return result
}

// CalculateBPJSForPeriod menghitung BPJS untuk periode tertentu.
// Berguna untuk audit trail dan penyimpanan ke database.
func CalculateBPJSForPeriod(grossSalary int64, period time.Time) BPJSResult {
	// Tarif bisa berubah per periode; untuk sekarang gunakan tarif default 2026
	_ = period
	return CalculateBPJS(grossSalary, DefaultBPJSRates())
}

// roundToInt membulatkan float ke int64 (pembulatan biasa).
func roundToInt(v float64) int64 {
	return int64(math.Round(v))
}
