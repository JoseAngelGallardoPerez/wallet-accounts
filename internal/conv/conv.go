package conv

func Int64FromInterface(param interface{}) int64 {
	switch param := param.(type) {
	case int64:
		return param
	case float64:
		return int64(param)
	case float32:
		return int64(param)
	case int:
		return int64(param)
	case int8:
		return int64(param)
	case int16:
		return int64(param)
	case int32:
		return int64(param)
	}
	return 0
}
