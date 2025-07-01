package utils

import "testing"

func TestIsValidMoon(t *testing.T) {
	testCases := []struct {
		Name     string
		Input    string
		Expected bool
	}{
		{
			Name:     "Valid_Moon_Number",
			Input:    "49927398716",
			Expected: true,
		},
		{
			Name:     "Invalid_Moon_Number",
			Input:    "49927398717",
			Expected: false,
		},
		{
			Name:     "Valid_Moon_Number_With_Even_Length",
			Input:    "1234567812345670",
			Expected: true,
		},
		{
			Name:     "Invalid_Moon_Number_With_Even_Length",
			Input:    "1234567812345678",
			Expected: false,
		},
		{
			Name:     "Invalid_Number_With_Characters",
			Input:    "4992739x8716",
			Expected: false,
		},
		{
			Name:     "Valid_Moon_Number_Single_Digit",
			Input:    "0",
			Expected: true,
		},
		{
			Name:     "Invalid_Moon_Number_Single_Digit",
			Input:    "5",
			Expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			result := IsValidMoon(tc.Input)
			if result != tc.Expected {
				t.Errorf("Expected %v, but got %v", tc.Expected, result)
			}
		})
	}
}
