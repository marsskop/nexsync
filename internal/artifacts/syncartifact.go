package artifacts

import (
	"strings"
	"fmt"
	"encoding/xml"
	"path/filepath"

	"golang.org/x/exp/slices"
	log "github.com/sirupsen/logrus"
	gojsonq "github.com/thedevsaddam/gojsonq/v2"
)

// Helper func to parse syncArtifact
func parseArtifact(syncArtifact string) (string, string) {
	artifactList := strings.Split(syncArtifact, "/")
	group := strings.ReplaceAll(artifactList[0], ".", "/")
	artifact := artifactList[1]
	return group, artifact
}

// get maven-metadata.xml
func (n Nexus) getMavenMetadata(syncArtifact string) (string, error) {
	group, artifact := parseArtifact(syncArtifact)
	var url string
	if n.Version == 3 {
		url = fmt.Sprintf("%s/repository/%s/%s/%s/maven-metadata.xml", n.URL, n.Repository, group, artifact)
	} else {
		url = fmt.Sprintf("%s/content/repositories/%s/%s/%s/maven-metadata.xml", n.URL, n.Repository, group, artifact)
	}
	resp, err := n.request(url, "GET")
	if err != nil {
		if strings.Contains(err.Error(), "responseCode: '404'") { // if 404, then maven-metadata.xml doesn't exist
			return "", nil
		}
		return "", err
	}
	xmlList := resp.strings
	return xmlList, nil
}

// A helpful struct to unmarshal maven-metadata.xml
type mavenMetadata struct {
	XMLName    xml.Name `xml:"metadata"`
	GroupId    string   `xml:"groupId"`
	ArtifactId string   `xml:"artifactId"`
	Version    []string `xml:"versioning>versions>Version"`
}

func unmarshalMetadata(metadata string) (string, string, []string) {
	x := &mavenMetadata{}
	err := xml.Unmarshal([]byte(metadata), x)
	if err != nil {
		return "", "", []string{}
	}
	name := x.ArtifactId
	group := x.GroupId
	versions := x.Version
	return name, group, versions
}

// Compare versions of artifact in xml
func compareVersions(listingVersionsFrom string, listingVersionsTo string, repoFrom *Nexus) ([]componentDict, error) {
	var diff []componentDict
	name, group, versionsFrom := unmarshalMetadata(listingVersionsFrom)
	var versionsTo []string
	if listingVersionsTo != "" {
		_, _, versionsTo = unmarshalMetadata(listingVersionsTo)
	} else {
		versionsTo = []string{}
	}
	for _, versionFrom := range versionsFrom { // iterate over versionsFrom
		if !slices.Contains(versionsTo, versionFrom) {
			path := filepath.Join(strings.Replace(group, ".", "/", -1), name, versionFrom)
			log.Debug(path)
			var downloadUrl string
			if repoFrom.Version == 3 {
				// http://proxy.s2b.tech/Nexus3/repository/dlg-maven/im/dlg/private-api-schema_2.13/0.0.8-42-gf9dbe9c-1.57.0/private-api-schema_2.13-0.0.8-42-gf9dbe9c-1.57.0.pom
				downloadUrl = fmt.Sprintf("%s/repository/%s/%s/%s/%s/%s-%s.jar", repoFrom.URL, repoFrom.Repository, strings.ReplaceAll(group, ".", "/"), name, versionFrom, name, versionFrom)
			} else {
				downloadUrl = fmt.Sprintf("%s/service/local/artifact/maven/content?g=%s&a=%s&v=%s&r=%s", repoFrom.URL, group, name, versionFrom, repoFrom.Repository)
			}
			diff = append(diff, componentDict{name: name, Version: versionFrom, group: group, path: path, assets: gojsonq.New().FromString(fmt.Sprintf("[ { \"downloadUrl\" : \"%s\" } ]", downloadUrl))})
		}
	}
	log.Debug(diff)
	return diff, nil
}

// Synchronize artifact
// Get list of componentDicts with diff versions
func GetDiffArtifact(repoFrom *Nexus, repoTo *Nexus, syncArtifact string) ([]componentDict, error) {
	var versionsDiff []componentDict
	listingVersionsFrom, err := repoFrom.getMavenMetadata(syncArtifact)
	if err != nil {
		return versionsDiff, err
	}
	listingVersionsTo, err := repoTo.getMavenMetadata(syncArtifact)
	if err != nil {
		return versionsDiff, err
	}
	versionsDiff, err = compareVersions(listingVersionsFrom, listingVersionsTo, repoFrom)
	if err != nil {
		return versionsDiff, err
	}
	return versionsDiff, nil
}

// Synchronize versions on artifact
func SyncDiffArtifact(diff []componentDict, repoFrom *Nexus, repoTo *Nexus, syncArtifact string) []error {
	var errors []error
	errsDownload := repoFrom.downloadComponents(diff)
	errsUpload := repoTo.uploadComponents(diff, repoFrom.Repository)
	errors = append(errsDownload, errsUpload...)
	return errors
}
