package api

import (
	"testing"
	
	"github.com/shanurcsenitap/irisk8s/internal/k8s"
)

func TestIsValidKubernetesName(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		expectedValid  bool
		expectedReason string
	}{
		// Valid cases
		{"Valid simple name", "myservice", true, ""},
		{"Valid with hyphen", "my-service", true, ""},
		{"Valid with numbers", "service123", true, ""},
		{"Valid combined", "my-service123", true, ""},
		
		// Invalid cases
		{"Empty name", "", false, "Name cannot be empty"},
		{"Too long", "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijkl", false, "Name must be 63 characters or less"},
		{"Uppercase letters", "MyService", false, "Name must consist of lower case alphanumeric characters or '-', start with an alphabetic character, and end with an alphanumeric character"},
		{"Underscore", "my_service", false, "Name must consist of lower case alphanumeric characters or '-', start with an alphabetic character, and end with an alphanumeric character"},
		{"Special chars", "my@service", false, "Name must consist of lower case alphanumeric characters or '-', start with an alphabetic character, and end with an alphanumeric character"},
		{"Starting with hyphen", "-myservice", false, "Name must consist of lower case alphanumeric characters or '-', start with an alphabetic character, and end with an alphanumeric character"},
		{"Ending with hyphen", "myservice-", false, "Name must consist of lower case alphanumeric characters or '-', start with an alphabetic character, and end with an alphanumeric character"},
		{"With dots", "my.service", false, "Name must consist of lower case alphanumeric characters or '-', start with an alphabetic character, and end with an alphanumeric character"},
		{"Starting with number", "1234-user", false, "Name must consist of lower case alphanumeric characters or '-', start with an alphabetic character, and end with an alphanumeric character"},
		{"Purely numeric", "1234", false, "Name must consist of lower case alphanumeric characters or '-', start with an alphabetic character, and end with an alphanumeric character"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			valid, reason := k8s.IsValidKubernetesName(tc.input)
			if valid != tc.expectedValid {
				t.Errorf("Expected validity %v, got %v for input %q", tc.expectedValid, valid, tc.input)
			}
			if reason != tc.expectedReason {
				t.Errorf("Expected reason %q, got %q for input %q", tc.expectedReason, reason, tc.input)
			}
		})
	}
}