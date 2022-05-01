package path

func In(obj any, path string, v1Str string) (bool, error) {
	return opList(obj, path, v1Str, equal)
}
