package testdata

func outer() {
	inner := func() {}
	inner()
}
