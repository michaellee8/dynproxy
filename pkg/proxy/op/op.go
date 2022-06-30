package op

type Operation struct {
	Type   OperationType
	Key    string
	Port   int
	Target string
}

func AddRuleOperation(key string) Operation {
	return Operation{
		Type: TypeAddRule,
		Key:  key,
	}
}

func RemoveRuleOperation(key string) Operation {
	return Operation{
		Type: TypeRemoveRule,
		Key:  key,
	}
}

func AddPortOperation(key string, port int) Operation {
	return Operation{
		Type: TypeAddPort,
		Key:  key,
		Port: port,
	}
}

func RemovePortOperation(key string, port int) Operation {
	return Operation{
		Type: TypeRemovePort,
		Key:  key,
		Port: port,
	}
}

func AddTargetOperation(key string, target string) Operation {
	return Operation{
		Type:   TypeAddTarget,
		Key:    key,
		Target: target,
	}
}

func RemoveTargetOperation(key string, target string) Operation {
	return Operation{
		Type:   TypeRemoveTarget,
		Key:    key,
		Target: target,
	}
}
