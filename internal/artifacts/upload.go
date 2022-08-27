package artifacts

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Helper func to create handler for opening
func mustOpen(f string) *os.File {
	r, err := os.Open(f)
	if err != nil {
		panic(err)
	}
	return r
}

// Configure map with values for multipart form
func configureFormValues(assets []assetDict, component componentDict) map[string]io.Reader {
	values := map[string]io.Reader{}
	genPom := true
	for i, asset := range assets {
		// file
		baseKey := fmt.Sprintf("maven2.asset%v", i+1)
		values[baseKey] = mustOpen(asset.path)
		// extension field
		extensionKey := fmt.Sprintf("%s.extension", baseKey)
		values[extensionKey] = strings.NewReader(asset.extension)
		// classifier field if not empty
		if asset.classifier != "" {
			classifierKey := fmt.Sprintf("%s.classifier", baseKey)
			values[classifierKey] = strings.NewReader(asset.classifier)
		}
		// generate-pom field if pom
		if asset.extension == "pom" {
			values["maven2.generate-pom"] = strings.NewReader("false")
			genPom = false
		}
	}
	// generate pom if there is none
	if genPom {
		values["maven2.generate-pom"] = strings.NewReader("true")
		values["maven2.groupId"] = strings.NewReader(component.group)
		values["maven2.artifactId"] = strings.NewReader(component.name)
		values["maven2.Version"] = strings.NewReader(component.Version)
	}
	return values
}

// Helper func to prepare multi-part content
func prepareMultipartForm(values map[string]io.Reader) (b bytes.Buffer, w *multipart.Writer, err error) {
	w = multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return b, w, fmt.Errorf("failed to create form file: %v", err)
			}
		} else {
			if fw, err = w.CreateFormField(key); err != nil {
				return b, w, fmt.Errorf("failed to create form field: %v", err)
			}
		}
		if _, err := io.Copy(fw, r); err != nil {
			return b, w, fmt.Errorf("failed to copy: %v", err)
		}
	}
	w.Close()
	return b, w, nil
}

func (n Nexus) uploadComponent(assets []assetDict, component componentDict) error {
	values := configureFormValues(assets, component)
	body, multipartWriter, err := prepareMultipartForm(values)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/service/rest/v1/components?repository=%s", n.URL, n.Repository)
	resp, err := n.requestWithMultipartForm(url, "POST", body, multipartWriter)
	if err != nil {
		return err
	}
	json := resp.strings
	log.Debug(json)
	return nil
}

// An assetDict is a useful struct for uploading assets
type assetDict struct {
	path                  string
	classifier, extension string
}

// Create list of assetDicts to upload assets
func (n Nexus) configureUploadAssets(component componentDict, repository string) ([]assetDict, error) {
	var assets []assetDict
	dirPath := filepath.Join(n.TmpDir, repository, component.path)
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// get extension from path
		ext := filepath.Ext(path)[1:]
		// get classifier from path
		Version := strings.Split(dirPath, "/")[len(strings.Split(dirPath, "/"))-1]
		afterVersionIdx := strings.LastIndex(path, Version) + len(Version)
		afterVersion := path[afterVersionIdx:]
		var classifier string
		if strings.Contains(afterVersion, "-") {
			classifier = strings.TrimPrefix(strings.Split(afterVersion, ".")[0], "-")
		} else {
			classifier = ""
		}
		// add to assets
		if !info.IsDir() {
			assets = append(assets, assetDict{path: path, classifier: classifier, extension: ext})
		}
		return nil
	})
	if err != nil {
		return assets, err
	}
	return assets, nil
}

// Upload components listed in componentDict (works only for Nexus.Version = 3)
func (n Nexus) uploadComponents(components []componentDict, repository string) []error {
	// upload all jars (parse names to get classifiers) + pom together
	var errors []error
	for _, component := range components {
		assets, err := n.configureUploadAssets(component, repository)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		err = n.uploadComponent(assets, component)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to upload component %s: %s", component.name, err))
		}
	}
	log.Debug(errors)
	return errors
}
