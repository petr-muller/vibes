package fauxinnati

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

type prereleasePullSpecResolver struct {
}

func (r *prereleasePullSpecResolver) ResolvePullSpec(client Client, version string, arch string) (string, error) {
	const prefix = "quay.io/openshift-release-dev/ocp-release"

	suffix := arch
	switch arch {
	case "amd64":
		suffix = "x86_64"
	case "arm64":
		suffix = "aarch64"
	}

	tag := fmt.Sprintf("%s-%s", version, suffix)
	req, err := http.NewRequest(http.MethodHead, fmt.Sprintf("https://quay.io/v2/openshift-release-dev/ocp-release/manifests/%s", tag), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code %d for url %s", res.StatusCode, req.URL)
	}

	digest := res.Header.Get("docker-content-digest")
	if digest == "" {
		return "", fmt.Errorf("missing digest for url %s", req.URL)
	}
	logrus.WithField("tag", tag).WithField("digest", digest).Debug("Resolved pull spec")
	return fmt.Sprintf("%s@%s", prefix, digest), nil
}
