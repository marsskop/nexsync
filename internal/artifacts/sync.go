package artifacts

import (
	"fmt"
	"strings"
	"path/filepath"

	"golang.org/x/exp/slices"
	log "github.com/sirupsen/logrus"
	gojsonq "github.com/thedevsaddam/gojsonq/v2"
	validate "github.com/go-playground/validator/v10"
)

// Get continuation token from json response
func getContinuationToken(json string) (string, bool) {
	token, err := findString(gojsonq.New().FromString(json), "continuationToken")
	if err != nil || token == "" {
		return "", false
	}
	return token, true
}

// List components in Nexus repository
func (n Nexus) listComponents() (string, error) {
	var json string
	var jsonNext string
	var token string
	tokenQuery := ""
	next := true
	for next {
		// get next page
		resp, err := n.request(fmt.Sprintf("%s/service/rest/v1/components?repository=%s&%s", n.URL, n.Repository, tokenQuery), "GET")
		if err != nil {
			return "", err
		}
		jsonNext = resp.strings
		if json != "" { // hack to merge jsons of pages
			json = strings.TrimSuffix(json, fmt.Sprintf(" ],\n  \"continuationToken\" : \"%s\"\n}", token)) + ",\n" + strings.TrimPrefix(jsonNext, "{\n  \"items\" : [ ")
		} else {
			json = json + jsonNext
		}
		if err := validate.New().Var(json, "required,json"); err != nil {
			return "", fmt.Errorf("response  doesn't seem to be a json: '%v'", err)
		}
		token, next = getContinuationToken(jsonNext)
		tokenQuery = fmt.Sprintf("continuationToken=%s", token)
	}
	return json, nil
}

// Compare lists of components in different repos
func compareListings(listingFrom string, listingTo string) ([]componentDict, error) {
	// compare versions (should we compare shas?..)
	// return those that are in listingFrom and not in listingTo (+ maybe those in listingFrom that have a different SHA than in listingTo)
	var diff []componentDict
	for i := 0; i < gojsonq.New().FromString(listingFrom).From("items").Count(); i++ { // iterate over listingFrom
		// get name of component
		name, err := findString(gojsonq.New().FromString(listingFrom).From("items"), fmt.Sprintf("[%v].name", i))
		if err != nil {
			return diff, fmt.Errorf("failed to find string in json: %v", err)
		}
		// get versions of component from repoFrom
		versionsFrom, err := pluckStringSlice(gojsonq.New().FromString(listingFrom).From("items").Where("name", "=", name), "Version")
		if err != nil {
			return diff, fmt.Errorf("failed to pluck string from json: %v", err)
		}
		// try to find versions of component in repoTo
		versionsTo, err := pluckStringSlice(gojsonq.New().FromString(listingTo).From("items").Where("name", "=", name), "Version")
		if err != nil {
			return diff, fmt.Errorf("failed to pluck string from json: %v", err)
		}
		// get group to configure path
		groups, err := pluckStringSlice(gojsonq.New().FromString(listingFrom).From("items").Where("name", "=", name), "group")
		if err != nil {
			return diff, fmt.Errorf("failed to pluck string from json: %v", err)
		}
		group := groups[0]
		log.Debug(fmt.Sprintf("%v versions: from '%v' to '%v'", name, versionsFrom, versionsTo))
		for _, versionFrom := range versionsFrom { // iterate over versionsFrom
			if !slices.Contains(versionsTo, versionFrom) {
				path := filepath.Join(strings.Replace(group, ".", "/", -1), name, versionFrom)
				log.Debug(path)
				assetQuery := gojsonq.New().FromString(listingFrom).From(fmt.Sprintf("items.[%v].assets", i))
				diff = append(diff, componentDict{name: name, Version: versionFrom, group: group, path: path, assets: assetQuery})
			}
		}
	}
	return diff, nil
}


// Synchronize whole repo
// Get list of componentDict with diff between repoFrom and repoTo
func GetDiff(repoFrom *Nexus, repoTo *Nexus) ([]componentDict, error) {
	listingFrom, err := repoFrom.listComponents()
	if err != nil {
		return []componentDict{}, err
	}
	listingTo, err := repoTo.listComponents()
	if err != nil {
		return []componentDict{}, err
	}
	diff, err := compareListings(listingFrom, listingTo)
	if err != nil {
		return []componentDict{}, err
	}
	return diff, nil
}

// Synchronize repos
func SyncDiff(diff []componentDict, repoFrom *Nexus, repoTo *Nexus) []error {
	var errors []error
	errsDownload := repoFrom.downloadComponents(diff)
	errsUpload := repoTo.uploadComponents(diff, repoFrom.Repository)
	errors = append(errsDownload, errsUpload...)
	return errors
}