package condition

var conditionRegistry = map[string]Condition{}

func RegisterCondition(name string, c Condition) {
	conditionRegistry[name] = c
}
func GetCondition(name string) (Condition, bool) {
	c, ok := conditionRegistry[name]
	return c, ok
}
