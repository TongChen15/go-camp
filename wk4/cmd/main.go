package main

import "fmt"

func main() {
	c, err := initializeCluster("example")
	if err != nil {
		fmt.Errorf("initialize the cluster failed: %v", err)
	}
	c.Start()
}
