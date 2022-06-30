package op

type OperationType string

const (
	TypeAddTarget    OperationType = "AddTarget"
	TypeRemoveTarget OperationType = "RemoveTarget"
	TypeAddPort      OperationType = "AddPort"
	TypeRemovePort   OperationType = "RemovePort"
	TypeAddRule      OperationType = "AddApp"
	TypeRemoveRule   OperationType = "RemoveApp"
)
