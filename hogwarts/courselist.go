//go:build !solution

package hogwarts

func dfsTopSort(v string, visited map[string]struct{}, tempPath map[string]struct{}, res []string, adjList map[string][]string) []string {
	visited[v] = struct{}{}
	tempPath[v] = struct{}{}

	for _, n := range adjList[v] {
		if _, ok := tempPath[n]; ok {
			panic("cycle detected")
		}

		if _, ok := visited[n]; !ok {
			res = dfsTopSort(n, visited, tempPath, res, adjList)
		}
	}

	delete(tempPath, v)
	res = append(res, v)
	return res
}

func GetCourseList(prereqs map[string][]string) []string {
	adjList := make(map[string][]string, len(prereqs))

	for v, pList := range prereqs {
		for _, p := range pList {
			if _, ok := adjList[p]; !ok {
				adjList[p] = []string{v}
			} else {
				adjList[p] = append(adjList[p], v)
			}
		}
	}

	visited, tempPath := make(map[string]struct{}), make(map[string]struct{})
	res := []string{}
	for v := range adjList {
		if _, ok := visited[v]; !ok {
			res = dfsTopSort(v, visited, tempPath, res, adjList)
		}
	}

	for i, j := 0, len(res)-1; i < j; i, j = i+1, j-1 {
		res[i], res[j] = res[j], res[i]
	}

	return res
}
