package units

import "strconv"

func FormatValue(v float64, precision int, scientific bool) string {
	if precision <= 0 {
		precision = 10
	}
	if scientific {
		return strconv.FormatFloat(v, 'e', precision-1, 64)
	}
	return strconv.FormatFloat(v, 'g', precision, 64)
}
