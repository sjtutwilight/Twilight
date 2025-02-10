package action

var actionRegistry = map[string]Action{}

func RegisterAction(name string, a Action) {
	actionRegistry[name] = a
}

func GetAction(name string) (Action, bool) {
	a, ok := actionRegistry[name]
	return a, ok
}
