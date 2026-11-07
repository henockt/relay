package auth

import "testing"

func TestAllowedEmail(t *testing.T) {
	tests := []struct {
		email string
		want  bool
	}{
		{"someone@gmail.com", true},
		{"Someone@ICloud.com", true},
		{"  user@proton.me ", true},
		{"foo@mailinator.com", false},
		{"foo@gmail.com.evil.io", false},
		{"foo@sub.gmail.com", false},
		{"a@b@gmail.com", true}, // domain is what follows the last @
		{"gmail.com", false},
		{"@gmail.com", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := AllowedEmail(tt.email); got != tt.want {
			t.Errorf("AllowedEmail(%q) = %v, want %v", tt.email, got, tt.want)
		}
	}
}
