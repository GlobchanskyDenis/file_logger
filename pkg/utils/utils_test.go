package utils

import (
	"testing"
	"time"
)

func TestCalcHour(t *testing.T) {
	testCases := []struct {
		name     string
		arg1     int
		arg2     int
		expected int
	}{
		{
			name:     "5, 2 ~ 4",
			arg1:     5,
			arg2:     2,
			expected: 4,
		},
		{
			name:     "23, 23 ~ 23",
			arg1:     23,
			arg2:     23,
			expected: 23,
		},
		{
			name:     "5, 24 ~ 0",
			arg1:     5,
			arg2:     24,
			expected: 0,
		},
		{
			name:     "22, 23 ~ 0",
			arg1:     22,
			arg2:     23,
			expected: 0,
		},
		{
			name:     "18, 6 ~ 18",
			arg1:     18,
			arg2:     6,
			expected: 18,
		},
		{
			name:     "17, 6 ~ 12",
			arg1:     17,
			arg2:     6,
			expected: 12,
		},
		{
			name:     "19, 6 ~ 18",
			arg1:     19,
			arg2:     6,
			expected: 18,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if result := CalcHour(tc.arg1, tc.arg2); result != tc.expected {
				t.Errorf("Fail: (%d, %d) expected %d got %d", tc.arg1, tc.arg2, tc.expected, result)
			}
		})
	}
}

func TestIsSameDateAndHour(t *testing.T) {
	t.Run("same", func(t *testing.T) {
		old, err := time.Parse("2006-01-02T15:04:05", "2022-09-30T15:00:00")
		if err != nil {
			t.Errorf("Error: %s", err)
			t.FailNow()
		}

		now, err := time.Parse("2006-01-02T15:04:05", "2022-09-30T17:00:00")
		if err != nil {
			t.Errorf("Error: %s", err)
			t.FailNow()
		}

		if result := IsSameDateAndHour(old, now, 6); result != true {
			t.Errorf("Fail: expected %#v got %#v", true, result)
		}
	})

	t.Run("not same", func(t *testing.T) {
		old, err := time.Parse("2006-01-02T15:04:05", "2022-09-30T15:00:00")
		if err != nil {
			t.Errorf("Error: %s", err)
			t.FailNow()
		}

		now, err := time.Parse("2006-01-02T15:04:05", "2022-09-30T17:00:00")
		if err != nil {
			t.Errorf("Error: %s", err)
			t.FailNow()
		}

		if result := IsSameDateAndHour(old, now, 2); result != false {
			t.Errorf("Fail: expected %#v got %#v", false, result)
		}
	})
}
