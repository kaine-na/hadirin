package compliance

import (
	"math"
	"time"
)

// TERCategory adalah kategori tarif efektif rata-rata PPh 21.
// Sesuai PMK 168/2023.
type TERCategory string

const (
	TERCategoryA TERCategory = "A" // TK/0
	TERCategoryB TERCategory = "B" // K/0, TK/1
	TERCategoryC TERCategory = "C" // K/1, K/2, TK/2, TK/3
)

// MaritalStatus status perkawinan untuk menentukan PTKP.
type MaritalStatus string

const (
	MaritalTK MaritalStatus = "TK" // Tidak Kawin
	MaritalK  MaritalStatus = "K"  // Kawin
)

// PTKPStatus menyimpan status PTKP karyawan.
type PTKPStatus struct {
	Marital      MaritalStatus
	Dependents   int // jumlah tanggungan (0-3)
}

// GetTERCategory menentukan kategori TER berdasarkan status PTKP.
func GetTERCategory(status PTKPStatus) TERCategory {
	switch status.Marital {
	case MaritalTK:
		if status.Dependents == 0 {
			return TERCategoryA
		}
		// TK/1, TK/2, TK/3
		if status.Dependents == 1 {
			return TERCategoryB
		}
		return TERCategoryC
	case MaritalK:
		if status.Dependents == 0 {
			return TERCategoryB
		}
		// K/1, K/2, K/3
		return TERCategoryC
	}
	return TERCategoryA
}

// terRateTableA adalah tabel tarif TER kategori A (TK/0) sesuai PMK 168/2023.
// Format: {batas_atas_penghasilan_bruto, tarif_persen}
// Penghasilan bruto adalah gaji bruto bulanan.
var terRateTableA = []struct {
	MaxIncome int64
	Rate      float64
}{
	{5_400_000, 0},
	{5_650_000, 0.25},
	{5_950_000, 0.50},
	{6_300_000, 0.75},
	{6_750_000, 1.00},
	{7_500_000, 1.25},
	{8_550_000, 1.50},
	{9_650_000, 1.75},
	{10_050_000, 2.00},
	{10_350_000, 2.25},
	{10_700_000, 2.50},
	{11_050_000, 3.00},
	{11_600_000, 3.00},
	{12_500_000, 3.00},
	{13_750_000, 5.00},
	{15_100_000, 5.00},
	{16_950_000, 7.50},
	{19_750_000, 7.50},
	{24_150_000, 10.00},
	{26_450_000, 10.00},
	{28_000_000, 12.50},
	{30_050_000, 12.50},
	{32_400_000, 15.00},
	{35_400_000, 15.00},
	{39_100_000, 17.50},
	{43_850_000, 17.50},
	{47_800_000, 20.00},
	{51_400_000, 20.00},
	{56_300_000, 22.50},
	{62_200_000, 25.00},
	{68_600_000, 25.00},
	{77_500_000, 30.00},
	{89_000_000, 30.00},
	{math.MaxInt64, 34.00},
}

// terRateTableB adalah tabel tarif TER kategori B (K/0, TK/1) sesuai PMK 168/2023.
var terRateTableB = []struct {
	MaxIncome int64
	Rate      float64
}{
	{6_200_000, 0},
	{6_500_000, 0.25},
	{6_850_000, 0.50},
	{7_300_000, 0.75},
	{9_200_000, 1.00},
	{10_750_000, 1.50},
	{11_250_000, 2.00},
	{11_600_000, 2.50},
	{12_600_000, 3.00},
	{13_600_000, 4.00},
	{14_950_000, 5.00},
	{16_400_000, 6.00},
	{18_450_000, 7.00},
	{21_850_000, 7.50},
	{26_000_000, 10.00},
	{27_700_000, 12.00},
	{29_350_000, 12.50},
	{31_450_000, 15.00},
	{33_950_000, 15.00},
	{37_100_000, 17.50},
	{41_100_000, 20.00},
	{45_800_000, 20.00},
	{49_500_000, 22.50},
	{53_800_000, 25.00},
	{58_500_000, 25.00},
	{64_000_000, 27.50},
	{71_000_000, 30.00},
	{80_000_000, 30.00},
	{93_000_000, 32.50},
	{math.MaxInt64, 34.00},
}

// terRateTableC adalah tabel tarif TER kategori C (K/1, K/2, TK/2, TK/3) sesuai PMK 168/2023.
var terRateTableC = []struct {
	MaxIncome int64
	Rate      float64
}{
	{6_600_000, 0},
	{6_950_000, 0.25},
	{7_350_000, 0.50},
	{7_800_000, 0.75},
	{8_850_000, 1.00},
	{9_800_000, 1.25},
	{10_950_000, 1.50},
	{11_200_000, 1.75},
	{12_050_000, 2.00},
	{12_950_000, 3.00},
	{14_150_000, 4.00},
	{15_550_000, 5.00},
	{17_050_000, 6.00},
	{19_500_000, 7.00},
	{22_700_000, 7.50},
	{26_600_000, 10.00},
	{28_100_000, 12.00},
	{30_100_000, 12.50},
	{32_600_000, 15.00},
	{35_400_000, 15.00},
	{38_900_000, 17.50},
	{43_000_000, 20.00},
	{47_400_000, 20.00},
	{51_200_000, 22.50},
	{55_800_000, 25.00},
	{60_400_000, 25.00},
	{66_700_000, 27.50},
	{74_500_000, 30.00},
	{86_000_000, 30.00},
	{math.MaxInt64, 34.00},
}

// getTERRate mengembalikan tarif TER (dalam persen) berdasarkan kategori dan penghasilan bruto.
func getTERRate(category TERCategory, grossMonthly int64) float64 {
	var table []struct {
		MaxIncome int64
		Rate      float64
	}

	switch category {
	case TERCategoryA:
		table = terRateTableA
	case TERCategoryB:
		table = terRateTableB
	case TERCategoryC:
		table = terRateTableC
	default:
		table = terRateTableA
	}

	for _, row := range table {
		if grossMonthly <= row.MaxIncome {
			return row.Rate
		}
	}
	return 34.0
}

// PPh21Input adalah input untuk kalkulasi PPh 21.
type PPh21Input struct {
	GrossMonthly int64        // gaji bruto bulanan
	PTKPStatus   PTKPStatus   // status PTKP
	Month        int          // bulan (1-12)
	Year         int          // tahun
	YTDGross     int64        // year-to-date gross (akumulasi bulan sebelumnya)
	YTDTax       int64        // year-to-date tax (akumulasi bulan sebelumnya)
}

// PPh21Result adalah hasil kalkulasi PPh 21.
type PPh21Result struct {
	GrossMonthly   int64       `json:"gross_monthly"`
	TERCategory    TERCategory `json:"ter_category"`
	TERRate        float64     `json:"ter_rate"`        // dalam persen
	PPh21Amount    int64       `json:"pph21_amount"`    // PPh 21 bulan ini
	IsDecember     bool        `json:"is_december"`
	// Untuk Desember: koreksi dengan tarif progresif
	YTDGross       int64       `json:"ytd_gross"`
	YTDTax         int64       `json:"ytd_tax"`
	AnnualGross    int64       `json:"annual_gross"`    // estimasi setahun
	AnnualTax      int64       `json:"annual_tax"`      // pajak setahun (progresif)
	DecemberTax    int64       `json:"december_tax"`    // koreksi Desember
	PTKP           int64       `json:"ptkp"`            // nilai PTKP setahun
	PKP            int64       `json:"pkp"`             // penghasilan kena pajak
}

// getPTKP mengembalikan nilai PTKP tahunan berdasarkan status.
// Sesuai PMK 168/2023 (PTKP 2016 masih berlaku).
func getPTKP(status PTKPStatus) int64 {
	base := int64(54_000_000) // TK/0
	if status.Marital == MaritalK {
		base += 4_500_000 // tambahan kawin
	}
	// Tambahan per tanggungan (max 3)
	deps := status.Dependents
	if deps > 3 {
		deps = 3
	}
	base += int64(deps) * 4_500_000
	return base
}

// calculateProgressiveTax menghitung pajak dengan tarif progresif (Pasal 17 UU PPh).
func calculateProgressiveTax(pkp int64) int64 {
	if pkp <= 0 {
		return 0
	}

	type bracket struct {
		limit int64
		rate  float64
	}
	brackets := []bracket{
		{60_000_000, 0.05},
		{250_000_000, 0.15},
		{500_000_000, 0.25},
		{5_000_000_000, 0.30},
		{math.MaxInt64, 0.35},
	}

	var tax int64
	prev := int64(0)
	for _, b := range brackets {
		if pkp <= prev {
			break
		}
		taxable := pkp
		if taxable > b.limit {
			taxable = b.limit
		}
		tax += roundToInt(float64(taxable-prev) * b.rate)
		prev = b.limit
	}
	return tax
}

// CalculatePPh21 menghitung PPh 21 dengan metode TER sesuai PMK 168/2023.
func CalculatePPh21(input PPh21Input) PPh21Result {
	category := GetTERCategory(input.PTKPStatus)
	terRate := getTERRate(category, input.GrossMonthly)

	result := PPh21Result{
		GrossMonthly: input.GrossMonthly,
		TERCategory:  category,
		TERRate:      terRate,
		YTDGross:     input.YTDGross,
		YTDTax:       input.YTDTax,
		PTKP:         getPTKP(input.PTKPStatus),
	}

	isDecember := input.Month == 12
	result.IsDecember = isDecember

	if !isDecember {
		// Bulan biasa: PPh 21 = gaji bruto x tarif TER
		result.PPh21Amount = roundToInt(float64(input.GrossMonthly) * terRate / 100)
		return result
	}

	// Desember: hitung ulang dengan tarif progresif untuk koreksi
	annualGross := input.YTDGross + input.GrossMonthly
	result.AnnualGross = annualGross

	// Biaya jabatan: 5% dari penghasilan bruto, max Rp 500.000/bulan = Rp 6.000.000/tahun
	biayaJabatan := int64(math.Min(float64(annualGross)*0.05, 6_000_000))

	pkp := annualGross - biayaJabatan - result.PTKP
	if pkp < 0 {
		pkp = 0
	}
	// PKP dibulatkan ke ribuan ke bawah
	pkp = (pkp / 1000) * 1000
	result.PKP = pkp

	annualTax := calculateProgressiveTax(pkp)
	result.AnnualTax = annualTax

	// PPh 21 Desember = pajak setahun - akumulasi pajak Jan-Nov
	decemberTax := annualTax - input.YTDTax
	if decemberTax < 0 {
		decemberTax = 0
	}
	result.DecemberTax = decemberTax
	result.PPh21Amount = decemberTax

	return result
}

// CalculatePPh21ForPeriod adalah wrapper dengan input periode time.Time.
func CalculatePPh21ForPeriod(grossSalary int64, status PTKPStatus, period time.Time, ytdGross, ytdTax int64) PPh21Result {
	return CalculatePPh21(PPh21Input{
		GrossMonthly: grossSalary,
		PTKPStatus:   status,
		Month:        int(period.Month()),
		Year:         period.Year(),
		YTDGross:     ytdGross,
		YTDTax:       ytdTax,
	})
}
