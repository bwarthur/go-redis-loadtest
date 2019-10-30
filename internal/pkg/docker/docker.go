package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// StopNode stops a redis node based on the container name
func StopNode(masterNode string, cli *client.Client) {
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		fmt.Printf("Error listing containers, %v\n", err)
		return
	}

	if len(containers) == 0 {
		fmt.Printf("Could not find any containers\n")
		return
	}

	for _, c := range containers {
		if contains(c.Names, fmt.Sprintf("/%s", masterNode)) {
			fmt.Printf("Stopping container  %s => %s\n", masterNode, c.ID)
			cli.ContainerStop(context.Background(), c.ID, nil)

			return
		}
	}

	fmt.Printf("Cannot find container '%s' to stop\n", masterNode)
}

// StartNode starts a redis node based on the container name
func StartNode(masterNode string, cli *client.Client) {
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		fmt.Printf("Error listing containers, %v\n", err)
		return
	}

	if len(containers) == 0 {
		fmt.Printf("Could not find any containers\n")
		return
	}

	for _, c := range containers {
		if contains(c.Names, fmt.Sprintf("/%s", masterNode)) {
			fmt.Printf("Starting container  %s => %s\n", masterNode, c.ID)
			cli.ContainerStart(context.Background(), c.ID, types.ContainerStartOptions{})

			return
		}
	}

	fmt.Printf("Cannot find container '%s' to start\n", masterNode)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
