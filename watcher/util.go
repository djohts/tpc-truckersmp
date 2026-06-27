package watcher

import "strings"

type AssignedVehiclesBlock struct {
	Id                string
	Vehicle           string
	VehiclePlacement  string
	Trailer           string
	TrailerPlacements []string
	TrailerAttached   bool
}

func getAssignedVehiclesBlock(lines []string) AssignedVehiclesBlock {
	inPlayerBlock := false
	inVehiclesBlock := false
	vehiclesBlockId := ""

	block := AssignedVehiclesBlock{}

loop:
	for _, line := range lines {
		switch {
		case inPlayerBlock && line == "}":
			inPlayerBlock = false

		case inVehiclesBlock && line == "}":
			inVehiclesBlock = false
			break loop

		case !inPlayerBlock && strings.HasPrefix(line, "player : "):
			inPlayerBlock = true

		case inPlayerBlock && strings.HasPrefix(line, " assigned_vehicles: "):
			vehiclesBlockId = strings.Split(line, ": ")[1]
			block.Id = vehiclesBlockId

		case !inVehiclesBlock && strings.HasPrefix(line, "player_vehicles : "+vehiclesBlockId):
			inVehiclesBlock = true

		case inVehiclesBlock:
			switch {
			case strings.HasPrefix(line, " vehicle: "):
				if nameless := strings.Split(line, ": ")[1]; nameless != "null" {
					block.Vehicle = nameless
				}

			case strings.HasPrefix(line, " stored_vehicle_placement: "):
				block.VehiclePlacement = strings.Split(line, ": ")[1]

			case strings.HasPrefix(line, " trailer: "):
				if nameless := strings.Split(line, ": ")[1]; nameless != "null" {
					block.Trailer = nameless
				}

			case strings.HasPrefix(line, " stored_trailer_placements["):
				block.TrailerPlacements = append(block.TrailerPlacements, strings.Split(line, ": ")[1])

			case strings.HasPrefix(line, " stored_trailer_attached: "):
				block.TrailerAttached = strings.Split(line, ": ")[1] == "true"
			}
		}
	}

	return block
}
