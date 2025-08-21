package main

import "fmt"

func DFS(graph map[int][]int, start int, visited map[int]bool) {
	if visited[start] {
		return
	}
	fmt.Printf("visited : %d \n", start)
	visited[start] = true
	for _, neighbor := range graph[start] {
		DFS(graph, neighbor, visited)

	}
}

func main() {
	graph := map[int][]int{
		0: {1, 2},
		1: {2},
		2: {0, 3},
		3: {3},
	}

	visited := make(map[int]bool)
	fmt.Print(" --- DFS start -- \n")
	DFS(graph, 2, visited) // Start DFS from node 2
	fmt.Println()
}
