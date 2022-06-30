package op

type OperationSlice []Operation

func (s OperationSlice) Len() int {
	return len(s)
}

func (s OperationSlice) Less(i, j int) bool {
	return getOperationPriority(s[i]) < getOperationPriority(s[j])
}

func (s OperationSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func getOperationPriority(op Operation) int {
	switch op.Type {
	case TypeRemovePort:
		return 1
	case TypeRemoveTarget:
		return 2
	case TypeRemoveRule:
		return 3
	case TypeAddRule:
		return 4
	case TypeAddTarget:
		return 5
	case TypeAddPort:
		return 6
	default:
		panic("impossible case: no such OperationType")
	}
}
