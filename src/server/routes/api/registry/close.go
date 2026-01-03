package registryApi

import (
	"context"
	"encoding/json"
	"fmt"
	"localapps-server/constants"
	"localapps-server/types"
	"net/http"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

func closeRegistry(w http.ResponseWriter, r *http.Request) {
	cli, _ := client.NewClientWithOpts(client.FromEnv)

	_, err := cli.Ping(context.Background())
	if err != nil {
		fmt.Println("Failed to connect to Docker daemon. Is it running?")
		return
	}

	deploymentContainers, _ := cli.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("network", "localapps-network"),
			filters.Arg("name", "localapps-deployment-registry"),
		),
	})

	w.Header().Set("Content-Type", "application/json")

	if len(deploymentContainers) < 1 {
		response := types.ApiError{
			Code:    constants.ErrorDeploymentRegistryClosed,
			Message: "There is no registry open",
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	fmt.Println("Closing unregistry container")

	cli.ContainerStop(context.Background(), deploymentContainers[0].ID, container.StopOptions{})
	w.WriteHeader(http.StatusNoContent)
}
