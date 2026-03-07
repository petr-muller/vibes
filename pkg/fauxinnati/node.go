package fauxinnati

import (
	"fmt"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/util/sets"
)

type NodeBuilder struct {
	queriedVersion semver.Version
	version        semver.Version
	channels       []string
	architecture   string

	client         Client
	getLatest      func(client Client, major, minor uint64) (semver.Version, error)
	digestResolver DigestResolver
}

type DigestResolver interface {
	getDigest(client Client, tag string) (string, error)
	getRepository() string
}

func latestPatchVersion(client Client, queriedVersion, version semver.Version, getLatest func(client Client, major, minor uint64) (semver.Version, error)) semver.Version {
	latest, err := getLatest(client, version.Major, version.Minor)
	if err != nil {
		logrus.WithError(err).WithField("version", version).WithField("major.minor", fmt.Sprintf("%d.%d", version.Minor, version.Minor)).Warning("Fail to find the latest version")
	} else if latest.LTE(queriedVersion) {
		logrus.WithField("queriedVersion", queriedVersion).WithField("latest", latest).Warning("The latest version is not greater than the queried version")
	} else if latest.LE(version) {
		logrus.WithField("version", version).WithField("latest", latest).Debug("Use the latest patch version")
		return latest
	}
	return version
}

func (b *NodeBuilder) Build() Node {
	version := latestPatchVersion(b.client, b.queriedVersion, b.version, b.getLatest)
	suffix := b.architecture
	switch b.architecture {
	case "":
		logrus.Debug("No architecture specified. Using default to resolve the image digest")
		suffix = "x86_64"
	case "amd64":
		suffix = "x86_64"
	case "arm64":
		suffix = "aarch64"
	}
	tag := fmt.Sprintf("%s-%s", version.String(), suffix)
	digest, err := b.digestResolver.getDigest(b.client, tag)
	if err != nil {
		digest = fmt.Sprintf("sha256:%064x", version.Major*1000000+version.Minor*1000+version.Patch)
	}
	metadata := map[string]string{
		"io.openshift.upgrades.graph.release.channels":    strings.Join(sets.List[string](sets.New[string](b.channels...)), ","),
		"io.openshift.upgrades.graph.release.manifestref": digest,
		"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", version.Major*1000+version.Minor*100+version.Patch),
	}
	if b.architecture == "multi" {
		metadata["release.openshift.io/architecture"] = b.architecture
	}
	return Node{
		Version:  version,
		Image:    fmt.Sprintf("%s@%s", b.digestResolver.getRepository(), digest),
		Metadata: metadata,
	}
}

func (b *NodeBuilder) WithChannels(channels []string) *NodeBuilder {
	b.channels = channels
	return b
}

func (b *NodeBuilder) WithArchitecture(architecture string) *NodeBuilder {
	b.architecture = architecture
	return b
}

func (b *NodeBuilder) WithVersion(version semver.Version) *NodeBuilder {
	b.version = version
	return b
}

func (b *NodeBuilder) WithClient(client Client) *NodeBuilder {
	b.client = client
	return b
}

func (b *NodeBuilder) withQueriedVersion(queriedVersion semver.Version) *NodeBuilder {
	b.queriedVersion = queriedVersion
	return b
}

func (b *NodeBuilder) withGetLatest(getLatest func(client Client, major, minor uint64) (semver.Version, error)) *NodeBuilder {
	b.getLatest = getLatest
	return b
}

func (b *NodeBuilder) WithDigestResolver(digestResolver DigestResolver) *NodeBuilder {
	b.digestResolver = digestResolver
	return b
}
