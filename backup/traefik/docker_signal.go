// 📄backup/traefik/docker_signal.go

package traefik

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

func GetTraefikContainerID() (string, error) {
	return queryContainerID("org.opencontainers.image.title=Traefik")
}

func queryContainerID(label string) (string, error) {
	client := dockerClient()
	filter := url.QueryEscape(fmt.Sprintf(`{"label":["%s"]}`, label))
	req, err := http.NewRequest("GET", "http://unix/containers/json?filters="+filter, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result []struct {
		ID string `json:"Id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result) == 0 {
		return "", fmt.Errorf("no container with label %s", label)
	}
	return result[0].ID, nil
}

func SendUSR1(id string) error {
	client := dockerClient()
	url := fmt.Sprintf("http://unix/containers/%s/kill?signal=USR1", id)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
}

func dockerClient() *http.Client {
	return &http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", "/var/run/docker.sock")
			},
		},
	}
}
