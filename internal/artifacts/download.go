package artifacts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

// Download file to specified n.TmpDir
func (n Nexus) downloadFileTmpDir(url string, component componentDict) (err error) {
	var relPath string
	if n.Version == 3 {
		relPath = fmt.Sprintf("%s/%s/%s", n.Repository, component.path, strings.Split(url, "/")[len(strings.Split(url, "/"))-1])
	} else {
		relPath = fmt.Sprintf("%s/%s/%s-%s.jar", n.Repository, component.path, component.name, component.Version)
	}
	fileName := filepath.Join(n.TmpDir, relPath)
	log.Debug(fileName)
	err = os.MkdirAll(filepath.Dir(fileName), os.ModePerm)
	out, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("failed to create file: %s", err)
	}
	defer func() {
		if errClose := out.Close(); errClose != nil {
			err = errClose
		}
	}()
	resp, err := n.request(url, "GET")
	if err != nil {
		return fmt.Errorf("failed to download file: %s", err)
	}
	_, err = out.WriteString(resp.strings)
	if err != nil {
		return fmt.Errorf("failed to write to file: %s", err)
	}
	return err
}

// List downloadUrls
func (n Nexus) configureDownloadAssets(component componentDict) ([]string, error) {
	var downloadUrls []string
	rawUrls, err := pluckStringSlice(component.assets, "downloadUrl")
	if err != nil {
		return downloadUrls, err
	}
	for _, url := range rawUrls {
		ext := filepath.Ext(url)[1:]
		if ext != "md5" && !strings.Contains(ext, "sha") {
			downloadUrls = append(downloadUrls, url)
		}
	}
	return downloadUrls, nil
}

// Download components listed in componentDict
func (n Nexus) downloadComponents(components []componentDict) []error {
	// download all jars (parse names to get classifiers) + pom (everything that is not sha*, md5)
	var wg sync.WaitGroup
	var errors []error
	for _, component := range components {
		assets, err := n.configureDownloadAssets(component)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		errChan := make(chan error, len(assets))
		for _, assetUrl := range assets {
			wg.Add(1)
			go func(downloadUrl string, errChan chan error) {
				defer wg.Done()
				log.Debug(downloadUrl)
				err := n.downloadFileTmpDir(downloadUrl, component)
				if err != nil {
					errChan <- err
				}
			}(assetUrl, errChan)
		}
		wg.Wait()
		close(errChan)
		for err := range errChan {
			errors = append(errors, err)
		}
	}
	log.Debug(errors)
	return errors
}
