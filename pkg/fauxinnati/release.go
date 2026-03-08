package fauxinnati

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/blang/semver/v4"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	kerrors "k8s.io/apimachinery/pkg/util/errors"
)

type prereleaseDigestResolver struct {
	cache Cache
}

func (r *prereleaseDigestResolver) getRepository() string {
	return "quay.io/openshift-release-dev/ocp-release"
}

func (r *prereleaseDigestResolver) getDigest(client Client, tag string) (string, error) {
	key := fmt.Sprintf("%s-%s", "getDigest", tag)
	if data, found := r.cache.Get(key); found {
		logrus.WithField("key", key).Debug("Found digest in cache")
		return data.(string), nil
	}
	logrus.WithField("key", key).Debug("Getting digest from quay.io ...")
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
	r.cache.Set(key, digest, cache.NoExpiration)
	return digest, nil
}

type latestCandidateGetter struct {
	cache Cache
}

func (g *latestCandidateGetter) latestCandidate(client Client, major, minor uint64) (semver.Version, error) {
	key := fmt.Sprintf("latestCandidate-%d.%d", major, minor)
	if data, found := g.cache.Get(key); found {
		logrus.WithField("key", key).Info("Found latest candidate in cache")
		return data.(semver.Version), nil
	}
	logrus.WithField("key", key).Debug("Getting latest candidate from github.com ...")
	var ret semver.Version
	versions, err := candidates(client, major, minor)
	if err != nil {
		return ret, err
	}
	if len(versions) == 0 {
		return ret, fmt.Errorf("no candidates found for version %d", major)
	}
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].LT(versions[j])
	})
	ret = versions[len(versions)-1]
	g.cache.Set(key, ret, 60*time.Minute)
	return ret, nil
}

func candidates(client Client, major, minor uint64) ([]semver.Version, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://raw.githubusercontent.com/openshift/cincinnati-graph-data/refs/heads/master/channels/candidate-%d.%d.yaml", major, minor), nil)
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d for url %s", res.StatusCode, req.URL)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var candidatesData CandidatesData
	if err := yaml.Unmarshal(data, &candidatesData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data (%s): %w", string(data), err)
	}
	var candidates []semver.Version
	var errs []error
	for _, v := range candidatesData.Versions {
		version, errP := semver.Parse(v)
		if errP != nil {
			errs = append(errs, errP)
		} else {
			candidates = append(candidates, version)
		}
	}
	return candidates, kerrors.NewAggregate(errs)
}

type CandidatesData struct {
	Versions []string `json:"versions"`
}
