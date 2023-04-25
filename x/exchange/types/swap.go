package types

func FindAllRoutes(denomMap map[string]map[string]uint64, fromDenom, toDenom string) (allRoutes [][]uint64) {
	var currentRoutes []uint64
	visited := map[uint64]struct{}{}
	var backtrack func(currentDenom string)
	// TODO: prevent stack overflow?
	backtrack = func(currentDenom string) {
		for denom, marketId := range denomMap[currentDenom] {
			if _, ok := visited[marketId]; !ok {
				if denom == toDenom {
					routes := make([]uint64, len(currentRoutes), len(currentRoutes)+1)
					copy(routes[:len(currentRoutes)], currentRoutes)
					routes = append(routes, marketId)
					allRoutes = append(allRoutes, routes)
				} else {
					visited[marketId] = struct{}{}
					currentRoutes = append(currentRoutes, marketId)
					backtrack(denom)
					currentRoutes = currentRoutes[:len(currentRoutes)-1]
					delete(visited, marketId)
				}
			}
		}
	}
	backtrack(fromDenom)
	return
}
