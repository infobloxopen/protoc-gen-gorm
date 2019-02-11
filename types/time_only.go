package types

func ParseTime(value string) (string, error) {
	return value, nil
}

func ParseValue(value string) (*TimeOnly, error) {
	return &TimeOnly{Value: value}, nil
}