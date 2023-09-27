package runtime

import opcuaruntime "harnsgateway/pkg/protocol/opcua/runtime"

type GroupVariables interface {
	~*opcuaruntime.Variable
}

func VariablesInGroupOf[T GroupVariables](variables []T, length int) [][]T {
	if len(variables) <= length {
		return [][]T{variables}
	}

	var group int
	if len(variables)%length == 0 {
		group = len(variables) / length
	} else {
		group = (len(variables) / length) + 1
	}
	groupVariables := make([][]T, group)

	start := 0
	for i := 0; i <= group; i++ {
		end := i * length
		if i != group {
			groupVariables = append(groupVariables, variables[start:end])
		} else {
			groupVariables = append(groupVariables, variables[start:])
		}
		start = i * length
	}
	return groupVariables
}
