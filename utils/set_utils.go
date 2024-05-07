package utils

type Set interface {
	~int | ~uint | ~uint8 | ~int8 | ~int32 | ~uint32 | ~int64 | ~uint64 | ~string
}

// 交集
func Intersection[T Set](a, b []T) (c []T) {
	m := make(map[T]bool)

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		if _, ok := m[item]; ok {
			c = append(c, item)
		}
	}
	return
}

// 并集
func Union[T Set](a, b []T) []T {
	m := make(map[T]bool)

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		if _, ok := m[item]; !ok {
			a = append(a, item)
		}
	}
	return a
}

// 差集 Set Difference: A - B
func Difference[T Set](a, b []T) (diff []T) {
	m := make(map[T]bool)

	for _, item := range b {
		m[item] = true
	}

	for _, item := range a {
		if _, ok := m[item]; !ok {
			diff = append(diff, item)
		}
	}
	return
}
