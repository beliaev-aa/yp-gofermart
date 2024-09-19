package utils

func IsValidMoon(number string) bool {
	var sum int
	alt := false
	numDigits := len(number)

	for i := numDigits - 1; i >= 0; i-- {
		n := int(number[i] - '0')
		if n < 0 || n > 9 {
			return false
		}
		if alt {
			n *= 2
			if n > 9 {
				n = (n % 10) + 1
			}
		}
		sum += n
		alt = !alt
	}
	return sum%10 == 0
}
