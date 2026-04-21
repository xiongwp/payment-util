package money

import "testing"

func TestMinorToStorage(t *testing.T) {
	cases := []struct {
		name    string
		minor   int64
		code    string
		storage int64
	}{
		{"PHP 100 cents = 1 PHP", 100, "PHP", 10000},
		{"PHP 10000 cents = 100 PHP", 10000, "PHP", 1000000},
		{"JPY 100 yen", 100, "JPY", 10000},
		{"KWD 1000 fils = 1 dinar", 1000, "KWD", 100000},
		{"USD 525 cents", 525, "USD", 52500},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := MinorToStorage(c.minor, c.code)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if got != c.storage {
				t.Errorf("got %d, want %d", got, c.storage)
			}
			// 回路：StorageToMinor 应该还原
			back, err := StorageToMinor(got, c.code)
			if err != nil || back != c.minor {
				t.Errorf("round trip: got %d err=%v, want %d", back, err, c.minor)
			}
		})
	}
}

func TestUnknownCurrency(t *testing.T) {
	if _, err := MinorToStorage(100, "XYZ"); err == nil {
		t.Errorf("expected error for unknown currency")
	}
}

func TestMinorUnitsPerMajor(t *testing.T) {
	cases := map[string]int64{
		"PHP": 100,
		"JPY": 1,
		"KWD": 1000,
	}
	for code, want := range cases {
		got, err := MinorUnitsPerMajor(code)
		if err != nil {
			t.Fatalf("%s err: %v", code, err)
		}
		if got != want {
			t.Errorf("%s: got %d, want %d", code, got, want)
		}
	}
}
