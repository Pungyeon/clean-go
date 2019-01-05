package utils

import "strconv"

const (
	TB = GB * 1000.0
	GB = MB * 1000.0
	MB = KB * 1000.0
	KB = 1000.0
)

func toFloatString(nbytes int64, divider float64) string {
	return strconv.FormatFloat(float64(nbytes)/divider, 'f', 2, 64)
}

func ToReadableSize(nbytes int64) string {
	switch {
	case nbytes > TB:
		return toFloatString(nbytes, TB) + " TB"
	case nbytes > GB:
		return toFloatString(nbytes, GB) + " GB"
	case nbytes > MB:
		return toFloatString(nbytes, MB) + " MB"
	case nbytes > KB:
		return toFloatString(nbytes, KB) + " KB"
	}
	return strconv.FormatInt(nbytes, 10) + " B"
}
