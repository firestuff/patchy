package path

func In(obj any, path string, matchStr string) (bool, error) {
	return opList(obj, path, matchStr, equal)
}
