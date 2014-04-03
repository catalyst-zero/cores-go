package client

func assertNotNil(object interface{}) {
	if object == nil {
		panic("'nil' not allowed as argument.")
	}
}
