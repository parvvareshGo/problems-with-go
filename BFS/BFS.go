package main

import (
	"fmt"
)

type Graph struct {
	vertices int
	adjList map[int][]int
}

func newGraph(vertices int) *Graph{
	return &Graph{
		vertices: vertices,
        adjList:  make(map[int][]int),
	}
}


func (g *Graph) addEdge(v, w int){
	g.adjList[v] = append(g.adjList[v], w)
}



func (g *Graph) BFS(start int){
	visited := make(map[int]bool)
	queue := []int{start}
	visited[start] = true

	for len(queue) > 0 {
		vartex := queue[0]
		queue = queue[1:]
		fmt.Print(vartex, " ")
		
		for _, neighbor := range g.adjList[vartex]{
			if !visited[neighbor]{
				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}
}


func main() {
    g := newGraph(5)
    g.addEdge(0, 1)
    g.addEdge(0, 2)
    g.addEdge(1, 3)
    g.addEdge(2, 4)

    fmt.Print("BFS Traversal: ")
    g.BFS(0)
}

