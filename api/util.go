package api

func IsCreate[T any](obj *T, prev *T) bool {
	return obj != nil && prev == nil
}

func IsUpdate[T any](obj *T, prev *T) bool {
	return obj != nil && prev != nil
}

func IsDelete[T any](obj *T, prev *T) bool {
	return obj == nil && prev != nil
}
