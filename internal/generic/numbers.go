package generic

type Number interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64
}

type Integer interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64
}

type Unsigned interface {
	uint | uint8 | uint16 | uint32 | uint64
}

type Signed interface {
	int | int8 | int16 | int32 | int64
}

func Sum[V Number](values ...V) (sum V) {
	for _, n := range values {
		sum += n
	}

	return
}

func Max[V Number](values ...V) V {
	if len(values) == 0 {
		panic("must have at least one value")
	}

	current := values[0]

	for i := 1; i < len(values); i++ {
		if values[i] > current {
			current = values[i]
		}
	}

	return current
}

func Min[V Number](values ...V) V {
	if len(values) == 0 {
		panic("must have at least one value")
	}

	current := values[0]

	for i := 1; i < len(values); i++ {
		if values[i] < current {
			current = values[i]
		}
	}

	return current
}

func Abs[V Signed](value V) V {
	if value < 0 {
		return -value
	}

	return value
}
