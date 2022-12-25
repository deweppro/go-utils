package domain

var dot = byte('.')

func Level(s string, level int) string {
	max := len(s) - 1
	count, pos := 0, 0
	if s[max] == dot {
		max--
	}

	for i := max; i >= 0; i-- {
		if s[i] == dot {
			count++
			if count == level {
				pos = i + 1
				break
			}
		}
	}
	return s[pos:]
}
