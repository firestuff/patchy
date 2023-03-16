package api

func IsCreate(obj *any, prev *any) bool {
	return obj != nil && prev == nil
}

func IsUpdate(obj *any, prev *any) bool {
	return obj != nil && prev != nil
}

func IsDelete(obj *any, prev *any) bool {
	return obj == nil && prev != nil
}
