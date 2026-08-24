package validator

import "testing"

func TestSanitizeText(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"<script>alert('xss')</script>Hello World", "alert('xss')Hello World"},
		{"<b>Bold Text</b>", "Bold Text"},
		{"  Clean text   ", "Clean text"},
		{"<a href=\"http://evil.com\">Click here</a>", "Click here"},
	}

	for _, c := range cases {
		out := SanitizeText(c.input)
		if out != c.expected {
			t.Errorf("SanitizeText(%q) = %q; expected %q", c.input, out, c.expected)
		}
	}
}

func TestValidateUUID(t *testing.T) {
	if !ValidateUUID("9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d") {
		t.Errorf("expected true for valid UUID")
	}
	if ValidateUUID("invalid-uuid-string") {
		t.Errorf("expected false for invalid UUID")
	}
}

func TestValidatePhoneNumber(t *testing.T) {
	if !ValidatePhoneNumber("+628123456789") {
		t.Errorf("expected true for +628123456789")
	}
	if !ValidatePhoneNumber("08123456789") {
		t.Errorf("expected true for 08123456789")
	}
	if ValidatePhoneNumber("abc") {
		t.Errorf("expected false for abc")
	}
}

