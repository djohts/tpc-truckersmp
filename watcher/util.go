package watcher

import "strings"

func getAssignedVehiclesBlockId(lines []string) string {
	inPlayerBlock := false
	for _, line := range lines {
		switch {
		case inPlayerBlock && strings.HasPrefix(line, "player : "):
			return ""

		case inPlayerBlock && strings.HasPrefix(line, "}"):
			return ""

		case !inPlayerBlock && strings.HasPrefix(line, "player : "):
			inPlayerBlock = true

		case inPlayerBlock && strings.HasPrefix(line, " assigned_vehicles: "):
			return strings.Split(line, ": ")[1]
		}
	}
	return ""
}
