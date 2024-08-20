package textmagic

var maxInSlice = 100

func splitSlice(slice []string) [][]string {
	toRet := make([][]string, 0, len(slice)/maxInSlice+1)

	for len(slice) > maxInSlice {
		toRet = append(toRet, slice[:maxInSlice])
		slice = slice[maxInSlice:]
	}

	if len(slice) > 0 {
		toRet = append(toRet, slice)
	}

	return toRet
}

func utos(num uint64) string {
	if num == 0 {
		return "0"
	}

	var digits [21]byte

	pos := 21

	for ; num > 0; num /= 10 {
		pos--
		digits[pos] = '0' + byte(num%10)
	}

	return string(digits[pos:])
}
