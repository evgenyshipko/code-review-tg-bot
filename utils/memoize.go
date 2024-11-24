package utils

import "fmt"

type GetProjectIdFunc func(projectName string) (int, error)

func Memoize(fn GetProjectIdFunc) GetProjectIdFunc {
	var memo = make(map[string]int, 10)

	return func(projectName string) (int, error) {

		fmt.Println("MEMOO", memo)

		if val, ok := memo[projectName]; ok {
			return val, nil
		}

		projectId, err := fn(projectName)
		if err != nil {
			return 0, err
		}
		memo[projectName] = projectId

		return projectId, nil
	}
}
