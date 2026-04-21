// Package money 跨服务共享的币种精度 + minor↔storage 换算。
//
// 约定：
//   - minor units = ISO 最小货币单位（PHP cents / JPY yen / KWD fils）
//   - storage = minor_units × StorageScale（= 100，跨币种恒定）
//
// 例：
//   PHP 100.00 → minor_units=10000 → storage=1_000_000
//   JPY 100    → minor_units=100   → storage=10_000
//   KWD 1.000  → minor_units=1000  → storage=100_000
package money

import "fmt"

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
