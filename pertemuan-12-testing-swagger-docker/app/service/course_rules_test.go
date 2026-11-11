package service

import "testing"

func TestWithinCreditLimit(t *testing.T) {
	tests := []struct {
		name   string
		taken  int
		adding int
		want   bool
	}{
		{name: "empty schedule", taken: 0, adding: 3, want: true},
		{name: "below limit", taken: 12, adding: 3, want: true},
		{name: "over by one", taken: 22, adding: 3, want: false},
		{name: "at limit cannot add", taken: 24, adding: 1, want: false},
		{name: "exactly reaches limit", taken: 21, adding: 3, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithinCreditLimit(tt.taken, tt.adding); got != tt.want {
				t.Errorf("WithinCreditLimit(%d, %d) = %v, want %v", tt.taken, tt.adding, got, tt.want)
			}
		})
	}
}

func TestCreditsAfterSwap(t *testing.T) {
	tests := []struct {
		name                    string
		taken, leaving, joining int
		want                    int
	}{
		{name: "same credits", taken: 18, leaving: 3, joining: 3, want: 18},
		{name: "more credits", taken: 18, leaving: 2, joining: 4, want: 20},
		{name: "fewer credits", taken: 18, leaving: 4, joining: 2, want: 16},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreditsAfterSwap(tt.taken, tt.leaving, tt.joining); got != tt.want {
				t.Errorf("CreditsAfterSwap(%d, %d, %d) = %d, want %d", tt.taken, tt.leaving, tt.joining, got, tt.want)
			}
		})
	}
}
