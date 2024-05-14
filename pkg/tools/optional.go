package tools

type genericLogger = func(...interface{})

func OrElse[T any](res T, e error, supplier func(e error) T) (r T) {
	r = res
	if e != nil {
		r = supplier(e)
	}

	return
}

// Lazy assertion
func Assert[T any](res T, e error) func(panicFunc genericLogger) T {
	return func(panicFunc genericLogger) T {
		return OrElse(res, e, func(e error) T {
			panicFunc(e)
			return res
		})
	}
}

// Lazy verification
func Should[T any](res T, e error) func(logFunc genericLogger) T {
	return func(logFunc genericLogger) T {
		return OrElse(res, e, func(e error) T {
			logFunc(e)
			return res
		})
	}
}
