package op

type Operation struct {
	Type   OperationType
	Key    string
	Port   int
	Target string
}
