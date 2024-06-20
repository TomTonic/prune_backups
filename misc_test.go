package main

import (
	"reflect"
	"testing"
	"time"
)

func Test_getAllButFirstMatchingPrefix(t *testing.T) {
	testCases := []struct {
		name   string
		from   []string
		prefix string
		want   []string
	}{
		{
			name:   "Test no matches",
			from:   []string{"cat", "cap", "car"},
			prefix: "df",
			want:   []string{},
		},
		{
			name:   "Test empty input",
			from:   []string{},
			prefix: "df",
			want:   []string{},
		},
		{
			name:   "Test empty prefix",
			from:   []string{"cat", "cap", "car"},
			prefix: "",
			want:   []string{"cap", "car"},
		},
		{
			name:   "Test 2 out of 4",
			from:   []string{"apple", "banana", "apricot", "grape"},
			prefix: "ap",
			want:   []string{"apricot"},
		},
		{
			name:   "Test all 3",
			from:   []string{"dog", "deer", "duck"},
			prefix: "d",
			want:   []string{"deer", "duck"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := getAllButFirstMatchingPrefix(tc.from, tc.prefix)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("getAllButFirstMatchingPrefix() = %v, want %v", got, tc.want)
			}
		})
	}
}

func Test_DateAdd(t *testing.T) {
	test_time := time.Date(2023, 3, 29, 20, 34, 58, 0, time.UTC)
	expect := time.Date(2023, 3, 1, 20, 34, 58, 0, time.UTC)
	got := test_time.AddDate(0, -1, 0)

	if got != expect {
		t.Errorf("DateAdd-Test: expected=%v, got=%v", expect, got)
	}

	test_time2 := time.Date(2023, 5, 31, 20, 34, 58, 0, time.UTC)
	expect2 := time.Date(2023, 5, 1, 20, 34, 58, 0, time.UTC)
	got2 := test_time2.AddDate(0, -1, 0)

	if got2 != expect2 {
		t.Errorf("DateAdd-Test: expected=%v, got=%v", expect2, got2)
	}
}

func Test_prevMonth(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name          string
		year          int
		month         int
		expectedYear  int
		expectedMonth int
	}{
		{"January to December", 2024, 1, 2023, 12},
		{"February to January", 2024, 2, 2024, 1},
		{"December to November", 2024, 12, 2024, 11},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the function with the current test case's year and month
			prevMonth(&tc.year, &tc.month)

			// Check if the year and month match the expected values
			if tc.year != tc.expectedYear || tc.month != tc.expectedMonth {
				t.Errorf("For %s, expected year:month to be %d:%d, but got %d:%d", tc.name, tc.expectedYear, tc.expectedMonth, tc.year, tc.month)
			}
		})
	}
}

func Test_toDateStr(t *testing.T) {
	tests := []struct {
		year   int
		month  int
		expect string
	}{
		{2024, 1, "2024-01"},
		{1912, 9, "1912-09"},
		{2024, 10, "2024-10"},
		{2024, 12, "2024-12"},
	}

	for _, test := range tests {
		result := toDateStr(test.year, test.month)
		if result != test.expect {
			t.Errorf("toDateStr(%d, %d) = %s; want %s", test.year, test.month, result, test.expect)
		}
	}
}

func Test_daysInMonth(t *testing.T) {
	tests := []struct {
		year  int
		month time.Month
		days  int
	}{
		{2024, time.April, 30},     // April
		{2024, time.June, 30},      // June
		{2024, time.September, 30}, // September
		{2024, time.November, 30},  // November
		{2024, time.January, 31},   // January
		{2024, time.March, 31},     // March
		{2024, time.May, 31},       // May
		{2024, time.July, 31},      // July
		{2024, time.August, 31},    // August
		{2024, time.October, 31},   // October
		{2024, time.December, 31},  // December
		{2024, time.February, 29},  // Leap year
		{2023, time.February, 28},  // Non-leap year
		{2000, time.February, 29},  // Leap year
		{2100, time.February, 28},  // Non-leap year
	}

	for _, test := range tests {
		if days := daysInMonth(test.year, test.month); days != test.days {
			t.Errorf("Year: %d, Month: %s, expected %d, got %d", test.year, test.month, test.days, days)
		}
	}
}

func Test_get15thOfMonthBefore(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "Jan 16, 2024",
			input:    time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2023, 12, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "End of February in non-leap year",
			input:    time.Date(2023, 2, 28, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "End of February in leap year",
			input:    time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "End of March in leap year",
			input:    time.Date(2024, 3, 30, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "End of March in non-leap year",
			input:    time.Date(2023, 3, 30, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2023, 2, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	// Run each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the function with the test case
			result := get15thOfMonthBefore(tc.input)

			// Check if the result is as expected
			if !result.Equal(tc.expected) {
				t.Errorf("Expected %v, but got %v", tc.expected, result)
			}
		})
	}
}

func Test_toDateStr3(t *testing.T) {
	type args struct {
		year  int
		month int
		day   int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test Case 1",
			args: args{year: 2024, month: 6, day: 16},
			want: "2024-06-16",
		},
		{
			name: "Test Case 2",
			args: args{year: 2024, month: 11, day: 5},
			want: "2024-11-05",
		},
		{
			name: "Test Case 3",
			args: args{year: 2024, month: 10, day: 15},
			want: "2024-10-15",
		},
		{
			name: "Test Case 4",
			args: args{year: 2024, month: 3, day: 5},
			want: "2024-03-05",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toDateStr3(tt.args.year, tt.args.month, tt.args.day); got != tt.want {
				t.Errorf("toDateStr3() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getUltimo(t *testing.T) {
	tests := []struct {
		year  int
		month time.Month
		want  time.Time
	}{
		{2024, time.January, time.Date(2024, time.January, 31, 0, 0, 0, 0, time.UTC)},
		{2024, time.February, time.Date(2024, time.February, 29, 0, 0, 0, 0, time.UTC)}, // Leap year
		{2023, time.February, time.Date(2023, time.February, 28, 0, 0, 0, 0, time.UTC)}, // Non-leap year
		{2024, time.March, time.Date(2024, time.March, 31, 0, 0, 0, 0, time.UTC)},
		{2024, time.April, time.Date(2024, time.April, 30, 0, 0, 0, 0, time.UTC)},
		{2024, time.December, time.Date(2024, time.December, 31, 0, 0, 0, 0, time.UTC)},
	}

	for _, tt := range tests {
		if got := getUltimo(tt.year, tt.month); !got.Equal(tt.want) {
			t.Errorf("getUltimo(%d, %d) = %v, want %v", tt.year, tt.month, got, tt.want)
		}
	}
}

func Test_getAnyMatchingAnyPrefixes(t *testing.T) {
	tests := []struct {
		search_in []string
		prefixes  []string
		want      bool
	}{
		{[]string{"apple", "banana", "cherry"}, []string{"a", "b"}, true},
		{[]string{"apple", "banana", "cherry"}, []string{"d", "e"}, false},
		{[]string{"apple", "banana", "cherry"}, []string{"ch", "ba"}, true},
		{[]string{}, []string{"a", "b"}, false},
		{[]string{"apple", "banana", "cherry"}, []string{}, false},
	}

	for _, tt := range tests {
		if got := getAnyMatchingAnyPrefixes(tt.search_in, tt.prefixes); got != tt.want {
			t.Errorf("getAnyMatchingAnyPrefixes(%v, %v) = %v, want %v", tt.search_in, tt.prefixes, got, tt.want)
		}
	}
}
