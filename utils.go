package textmagic

const joinSep = ','

func joinUints(u []uint64) string {
	toStr := make([]byte, 0, 10*len(u))
	var digits [21]byte
	for n, num := range u {
		if n > 0 {
			toStr = append(toStr, joinSep)
		}
		if num == 0 {
			toStr = append(toStr, '0')
			continue
		}
		pos := 21
		for ; num > 0; num /= 10 {
			pos--
			digits[pos] = '0' + byte(num%10)
		}
		toStr = append(toStr, digits[pos:]...)
	}
	return string(toStr)
}

var maxInSlice = 100

func splitSlice(slice []uint64) [][]uint64 {
	toRet := make([][]uint64, 0, len(slice)/maxInSlice+1)
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

func stou(str string) (uint64, bool) {
	var num uint64
	for _, c := range str {
		if c > '9' || c < '0' {
			return 0, false
		}
		num *= 10
		num += uint64(c - '0')
	}
	return num, true
}
