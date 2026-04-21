// Package money 跨服务共享的币种精度 + minor↔storage 换算 + 展示格式化。
//
// 约定：
//   - minor units = ISO 最小货币单位（PHP cents / JPY yen / KWD fils）
//   - storage = minor_units × StorageScale（= 100，跨币种恒定）
//
// 例：
//   PHP 100.00 → minor_units=10000   → storage=1_000_000   → Display="₱100.00"
//   USD 100.34 → minor_units=10034   → storage=1_003_400   → Display="$100.34"
//   JPY 100    → minor_units=100     → storage=10_000      → Display="¥100"
//   KWD 1.000  → minor_units=1000    → storage=100_000     → Display="KD1.000"
package money

import (
	"fmt"
	"strconv"
)

// StorageScale accounting-system 内部 storage 相对 ISO minor units 的放大倍数。
// 跨币种恒定，目的是给未来高精度（加密货币小额手续费）场景留余量。
const StorageScale int64 = 100

// precisionMinor ISO 4217 最小货币单位相对主单位的小数位。
var precisionMinor = map[string]int{
	"PHP": 2, "USD": 2, "EUR": 2, "GBP": 2, "HKD": 2,
	"SGD": 2, "AUD": 2, "CAD": 2, "CHF": 2, "MYR": 2,
	"THB": 2, "INR": 2, "TWD": 2, "CNY": 2,
	"VND": 0, "IDR": 0, "JPY": 0, "KRW": 0,
	"KWD": 3, "BHD": 3, "OMR": 3,
}

// currencySymbol 给 Display 用的前缀。没登记的币种 Symbol() 会回落到 ISO 码 + 空格。
//
// 短符号（$、¥、€、₱…）紧贴金额；多字母无独占符号的（CHF/MYR/IDR…）保留 ISO
// 码作前缀，下游想要 "CHF 100.00" 这种带空格风格的可以自己拼。
var currencySymbol = map[string]string{
	"USD": "$", "EUR": "€", "GBP": "£", "JPY": "¥", "CNY": "¥",
	"PHP": "₱", "THB": "฿", "INR": "₹", "VND": "₫", "KRW": "₩",
	"HKD": "HK$", "SGD": "S$", "AUD": "A$", "CAD": "C$", "TWD": "NT$",
	"CHF": "CHF", "MYR": "RM", "IDR": "Rp",
	"KWD": "KD", "BHD": "BD", "OMR": "OMR",
}

// IsSupported 币种是否已登记。
func IsSupported(code string) bool { _, ok := precisionMinor[code]; return ok }

// Precision 返回币种的 ISO 小数位。
func Precision(code string) (int, error) {
	p, ok := precisionMinor[code]
	if !ok {
		return 0, fmt.Errorf("money: unknown currency %q", code)
	}
	return p, nil
}

// MinorUnitsPerMajor 1 个主单位对应的最小单位数。PHP=100, JPY=1, KWD=1000。
func MinorUnitsPerMajor(code string) (int64, error) {
	p, err := Precision(code)
	if err != nil {
		return 0, err
	}
	n := int64(1)
	for i := 0; i < p; i++ {
		n *= 10
	}
	return n, nil
}

// MinorToStorage minor_units → storage。跨币种都是 ×StorageScale。
func MinorToStorage(minorUnits int64, code string) (int64, error) {
	if !IsSupported(code) {
		return 0, fmt.Errorf("money: unknown currency %q", code)
	}
	return minorUnits * StorageScale, nil
}

// StorageToMinor storage → minor_units。要求 storage % StorageScale == 0。
func StorageToMinor(storage int64, code string) (int64, error) {
	if !IsSupported(code) {
		return 0, fmt.Errorf("money: unknown currency %q", code)
	}
	if storage%StorageScale != 0 {
		return 0, fmt.Errorf("money: storage %d not divisible by scale %d", storage, StorageScale)
	}
	return storage / StorageScale, nil
}

// Symbol 返回币种用于展示的前缀。未登记的币种回落到 "<CODE> "（带尾随空格），
// 这样 Symbol(code)+amount 永远是合法可读字符串。
func Symbol(code string) string {
	if s, ok := currencySymbol[code]; ok {
		return s
	}
	return code + " "
}

// FormatMinor minor_units → 主单位字符串，按 ISO 小数位补零。
//   USD 10034 → "100.34"
//   JPY 100   → "100"
//   KWD 1500  → "1.500"
//   USD -25   → "-0.25"
func FormatMinor(minorUnits int64, code string) (string, error) {
	p, err := Precision(code)
	if err != nil {
		return "", err
	}
	sign := ""
	if minorUnits < 0 {
		sign = "-"
		minorUnits = -minorUnits
	}
	if p == 0 {
		return sign + strconv.FormatInt(minorUnits, 10), nil
	}
	div := int64(1)
	for i := 0; i < p; i++ {
		div *= 10
	}
	major := minorUnits / div
	frac := minorUnits % div
	return fmt.Sprintf("%s%d.%0*d", sign, major, p, frac), nil
}

// FormatStorage storage → 主单位字符串。等价于 StorageToMinor + FormatMinor。
//   USD 1_003_400 → "100.34"
func FormatStorage(storage int64, code string) (string, error) {
	m, err := StorageToMinor(storage, code)
	if err != nil {
		return "", err
	}
	return FormatMinor(m, code)
}

// Display storage → 带币种符号的人类可读金额。给 ledger/admin UI/收据用。
//   USD 1_003_400 → "$100.34"
//   JPY 10_000    → "¥100"
//   KWD 150_000   → "KD1.500"
//   未登记币种    → "XYZ 100.00"（Symbol fallback）
func Display(storage int64, code string) (string, error) {
	s, err := FormatStorage(storage, code)
	if err != nil {
		return "", err
	}
	return Symbol(code) + s, nil
}
