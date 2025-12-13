package middlewares

import (
	"context"
	"fmt"
	"localapps-server/constants"
	"localapps-server/utils"
	"net/http"
	"net/http/httputil"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

var idleTimeouts = make(map[string]time.Duration)

func AppProxy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var appEnvironmentVars []string

		if r.Host == strings.Split(utils.ServerConfig.AccessUrl, "://")[1] {
			next.ServeHTTP(w, r)
			return
		}

		appId := strings.Split(r.Host, ".")[0]

		appData, err := utils.GetAppData(appId)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		cli, _ := client.NewClientWithOpts(client.FromEnv)

		_, err = cli.Ping(context.Background())
		if err != nil {
			fmt.Fprintf(w, "Failed to connect to Docker engine: %s", err)
			return
		}

		var currentPartName string
		var fallbackPartName string

		for name, path := range appData.Parts {
			if path == "" {
				fallbackPartName = name
			}

			if strings.Split(r.URL.Path, "/")[1] == path {
				currentPartName = name
				break
			} else {
				currentPartName = fallbackPartName
			}
		}

		containersByLabels, _ := cli.ContainerList(context.Background(), container.ListOptions{
			Filters: filters.NewArgs(
				filters.Arg("label", "LOCALAPPS_APP_ID="+appId),
				filters.Arg("label", "LOCALAPPS_APP_PART="+currentPartName),
			),
		})

		var freePort int
		var containerAddress string

		idleTimeoutToSet := 20 * time.Second

		if len(containersByLabels) > 0 {
			if constants.IsDebugBuild {
				fmt.Printf("Reseting timeout for app '%s': %v -> %v\n", appId, idleTimeouts[appId], idleTimeoutToSet)
			}

			idleTimeouts[appId] = idleTimeoutToSet

			if constants.IsRunningInContainer() {
				containerInspect, _ := cli.ContainerInspect(context.Background(), containersByLabels[0].ID)
				containerAddress = strings.TrimPrefix(containerInspect.Name, "/")
				freePort = 80
			} else {
				portIndex := slices.IndexFunc(containersByLabels[0].Ports, func(port container.Port) bool {
					return port.PrivatePort == 80
				})

				containerPort := containersByLabels[0].Ports[portIndex].PublicPort
				containerAddress = "localhost"
				freePort = int(containerPort)
			}
		} else {
			config := container.Config{
				Image: "localapps/apps/" + appId + "/" + currentPartName,
				Env:   append(appEnvironmentVars, "PORT=80"),
				Labels: map[string]string{
					"LOCALAPPS_APP_ID":   appId,
					"LOCALAPPS_APP_PART": currentPartName,
				},
			}

			hostConfig := container.HostConfig{
				AutoRemove:  true,
				Mounts:      []mount.Mount{{Type: mount.TypeVolume, Source: "localapps-storage-" + appId, Target: "/storage"}},
				NetworkMode: "localapps-network",
			}

			if !constants.IsRunningInContainer() {
				config.ExposedPorts = nat.PortSet{"80": struct{}{}}
				hostConfig.PortBindings = nat.PortMap{
					"80": {
						{
							HostIP:   "0.0.0.0",
							HostPort: strconv.Itoa(freePort),
						},
					},
				}
			}

			appNameWithPart := appId + "/" + currentPartName
			createdContainer, _ := cli.ContainerCreate(context.Background(), &config, &hostConfig, nil, nil, "")

			idleTimeouts[appId] = idleTimeoutToSet

			if constants.IsDebugBuild {
				fmt.Printf("Creating container for app '%s' with idle timeout of %v\n", appId, idleTimeouts[appId])
			}

			fmt.Printf("[app:%s] Got a http request while stopped - creating container with %v ilde timeout\n", appNameWithPart, idleTimeouts[appId])

			if err := cli.ContainerStart(context.Background(), createdContainer.ID, container.StartOptions{}); err != nil {
				fmt.Fprintf(w, "Failed to start app \"%s\": %s", appId, err)
				return
			}

			if constants.IsRunningInContainer() {
				containerInspect, _ := cli.ContainerInspect(context.Background(), createdContainer.ID)
				containerAddress = strings.TrimPrefix(containerInspect.Name, "/")
				freePort = 80
			} else {
				containerInspect, _ := cli.ContainerInspect(context.Background(), createdContainer.ID)
				containerPort := containerInspect.NetworkSettings.Ports["80/tcp"][0].HostPort
				containerAddress = "localhost"
				freePort, _ = strconv.Atoi(containerPort)
			}

			go func() {
				originalIdleTimeout := idleTimeouts[appId]

				for idleTimeouts[appId] > 0 {
					time.Sleep(time.Second)

					if constants.IsDebugBuild {
						fmt.Printf("Updating timeout for '%s': %v -> %v\n", appId, idleTimeouts[appId], idleTimeouts[appId]-time.Second)
					}

					idleTimeouts[appId] = idleTimeouts[appId] - time.Second
				}

				fmt.Printf("[app:%s] Idle timeout exceeded (%v) - stopping container\n", appNameWithPart, originalIdleTimeout)
				cli.ContainerStop(context.Background(), createdContainer.ID, container.StopOptions{})

				if constants.IsDebugBuild {
					fmt.Printf("Removing '%s' from idleTimeouts, new length: %v -> %v\n", appId, len(idleTimeouts), len(idleTimeouts)-1)
				}

				delete(idleTimeouts, appId)
			}()
		}

		// Wait for the app to be ready
		containerAccessPoint := fmt.Sprintf("http://%s:%d", containerAddress, freePort)
		for {
			_, err := http.Get(containerAccessPoint)
			if err == nil {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}

		appUrl, _ := url.Parse(containerAccessPoint)
		httputil.NewSingleHostReverseProxy(appUrl).ServeHTTP(w, r)
	})
}
