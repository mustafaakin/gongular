package gongular

import (
	"math"
	"reflect"
)

// TODO: Do for float, int and uint
func compareAndReturnIntAndRanges(val, lower, upper int64) (bool, int64, int64) {
	result := val >= lower && val <= upper
	return result, lower, upper
}

func compareAndReturnUIntAndRanges(val, lower, upper uint64) (bool, uint64, uint64) {
	result := val >= lower && val <= upper
	return result, lower, upper
}

func compareAndReturnFloatAndRanges(val, lower, upper float64) (bool, float64, float64) {
	result := val >= lower && val <= upper
	return result, lower, upper
}

func checkIntRange(kind reflect.Kind, val int64) (bool, int64, int64) {
	switch kind {
	case reflect.Int8:
		return compareAndReturnIntAndRanges(val, math.MinInt8, math.MaxInt8)
	case reflect.Int16:
		return compareAndReturnIntAndRanges(val, math.MinInt16, math.MaxInt16)
	case reflect.Int32, reflect.Int:
		return compareAndReturnIntAndRanges(val, math.MinInt32, math.MaxInt32)
	case reflect.Int64:
		return compareAndReturnIntAndRanges(val, math.MinInt64, math.MaxInt64)
	}
	// Should not be here
	return false, math.MinInt64, math.MaxInt64
}

func checkUIntRange(kind reflect.Kind, val uint64) (bool, uint64, uint64) {
	switch kind {
	case reflect.Uint8:
		return compareAndReturnUIntAndRanges(val, 0, math.MaxUint8)
	case reflect.Uint16:
		return compareAndReturnUIntAndRanges(val, 0, math.MaxUint16)
	case reflect.Uint32, reflect.Uint:
		return compareAndReturnUIntAndRanges(val, 0, math.MaxUint32)
	case reflect.Uint64:
		return compareAndReturnUIntAndRanges(val, 0, math.MaxUint64)
	}
	// Should not be here
	return false, 0, math.MaxUint64
}

func checkFloatRange(kind reflect.Kind, val float64) (bool, float64, float64) {
	switch kind {
	case reflect.Float32:
		// TODO: Validate this, is it really true
		return compareAndReturnFloatAndRanges(val, -math.MaxFloat32-1, math.MaxFloat32)
	case reflect.Float64:
		return compareAndReturnFloatAndRanges(val, -math.MaxFloat64-1, math.MaxFloat64)
	}
	// Should not be here
	return false, 0, math.MaxFloat64
}
