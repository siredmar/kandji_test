package filter

type Filter interface {
	// Eval should return true iff the input should be included.
	Eval(interface{}) (bool, error)
}
