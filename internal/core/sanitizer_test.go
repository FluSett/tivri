package core

import "testing"

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"user@example.com", true},
		{"john.doe+tag@sub.domain.co.uk", true},
		{"user;name@domain.com", false},
		{"user@domain;com", false},
		{"user@domain", false},
		{"@domain.com", false},
		{"user@", false},
		{"user @domain.com", false},
		{"user<name>@domain.com", false},
		{"", false},
	}

	for _, tt := range tests {
		got := IsValidEmail(tt.email)
		if got != tt.valid {
			t.Errorf("IsValidEmail(%q) = %v, want %v", tt.email, got, tt.valid)
		}
	}
}
