package utils

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func ExecError(message string) {
	panic(message)
}
