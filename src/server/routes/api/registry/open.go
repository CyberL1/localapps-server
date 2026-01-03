package registryApi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"localapps-server/constants"
	"localapps-server/types"
	"localapps-server/utils"
	"net/http"
	"runtime"
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func openRegistry(w http.ResponseWriter, r *http.Request) {
	cli, _ := client.NewClientWithOpts(client.FromEnv)

	_, err := cli.Ping(context.Background())
	if err != nil {
		fmt.Println("Failed to connect to Docker daemon. Is it running?")
		return
	}

	registryImages, _ := cli.ImageList(context.Background(), image.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("reference", constants.UnregistryImage),
		),
	})

	if len(registryImages) < 1 {
		pull, err := cli.ImagePull(context.Background(), constants.UnregistryImage, image.PullOptions{})
		if err != nil {
			response := types.ApiError{
				Code:    constants.ErrorDockerEngine,
				Message: "Couldn't pull unregistry image",
				Error:   err,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}
		defer pull.Close()

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Transfer-Encoding", "chunked")

		buf := make([]byte, 32*1024)
		for {
			n, err := pull.Read(buf)
			if n > 0 {
				_, _ = w.Write(buf[:n])
				w.(http.Flusher).Flush()
			}
			if err != nil {
				if err == io.EOF {
					break
				}
				return
			}
		}
	}

	deploymentContainers, _ := cli.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("network", "localapps-network"),
			filters.Arg("name", "localapps-deployment-registry"),
		),
	})

	var freePort int

	if len(deploymentContainers) < 1 {
		fmt.Println("Opening unregistry container for app deployment")
		freePort, _ = utils.GetFreePort()

		var containerdSockPath string

		switch runtime.GOOS {
		case "darwin":
			containerdSockPath = "/run/docker/containerd/containerd.sock"
		default:
			containerdSockPath = "/run/containerd/containerd.sock"
		}

		unregistryConfig := container.Config{
			Image:        constants.UnregistryImage,
			ExposedPorts: nat.PortSet{"5000": struct{}{}},
			Env:          []string{"UNREGISTRY_CONTAINERD_SOCK=" + containerdSockPath},
		}

		unregistryHostConfig := container.HostConfig{
			AutoRemove:   true,
			Binds:        []string{containerdSockPath + ":" + containerdSockPath},
			PortBindings: nat.PortMap{"5000": {{HostIP: "0.0.0.0", HostPort: strconv.Itoa(freePort)}}},
			NetworkMode:  "localapps-network",
		}

		unregistryContainer, err := cli.ContainerCreate(context.Background(), &unregistryConfig, &unregistryHostConfig, nil, nil, "localapps-deployment-registry")
		if err != nil {
			fmt.Println("failed to create unregistry container:", err)
			return
		}

		if err := cli.ContainerStart(context.Background(), unregistryContainer.ID, container.StartOptions{}); err != nil {
			fmt.Println("Failed to start unregistry container:", err)
			return
		}
	} else {
		unregistryContainer := deploymentContainers[0]
		freePort = int(unregistryContainer.Ports[0].PublicPort)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"port": strconv.Itoa(freePort),
	})
}
