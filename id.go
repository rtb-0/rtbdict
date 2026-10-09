package rtbdict

// ID is a numeric category id.
type ID interface {
	~int | ~int16 | ~int64 | uint | ~uint16 | ~uint64
}
