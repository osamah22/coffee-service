package auth

import "testing"

func TestValidatePasswordAllowsAnySixCharacterPassword(t *testing.T) {
	validPasswords := []string{
		"aaaaaa",
		"123456",
		"!!!!!!",
	}

	for _, password := range validPasswords {
		if err := validatePassword(password, "public"); err != nil {
			t.Fatalf("expected %q to be accepted, got %q", password, *err)
		}
	}
}

func TestValidatePasswordRejectsOnlyPasswordsUnderSixCharacters(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
	}{
		{name: "short string", value: "abcde"},
		{name: "missing", value: nil},
		{name: "wrong type", value: 123456},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validatePassword(tt.value, "public"); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
