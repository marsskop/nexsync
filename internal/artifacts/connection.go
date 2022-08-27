package artifacts

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

// Nexus is a useful struct for Nexus repositories
type Nexus struct {
	URL                            string
	Version                        int
	User, Pass, Repository, TmpDir string
}

/////////////////////////////////////////////////////////////////
// Request wrapper for Nexus repos
/////////////////////////////////////////////////////////////////
type requestJSONResponse struct {
	bytes   []byte
	strings string
}

func (n Nexus) request(url string, method string) (requestJSONResponse, error) {
	log.WithFields(log.Fields{"Nexus URL": n.URL, "METHOD": method, "URL": url, "User": n.User}).Debug("Request")
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return requestJSONResponse{}, err
	}
	// req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(n.User, n.Pass)
	bodyBytes, bodyString, errs := n.response(req)
	for _, err := range errs {
		if err != nil {
			return requestJSONResponse{}, err
		}
	}
	return requestJSONResponse{bodyBytes, bodyString}, nil
}

func (n Nexus) requestWithMultipartForm(url string, method string, body bytes.Buffer, w *multipart.Writer) (requestJSONResponse, error) {
	log.WithFields(log.Fields{"Nexus URL": n.URL, "METHOD": method, "URL": url, "User": n.User}).Debug("Request")
	req, err := http.NewRequest(method, url, &body)
	if err != nil {
		return requestJSONResponse{}, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.SetBasicAuth(n.User, n.Pass)
	bodyBytes, bodyString, errs := n.response(req)
	for _, err := range errs {
		if err != nil {
			return requestJSONResponse{}, err
		}
	}
	return requestJSONResponse{bodyBytes, bodyString}, nil
}

func (n Nexus) response(req *http.Request) (b []byte, s string, errs []error) {
	client := http.Client{
		Timeout: 60 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error { // propagate auth headers in redirects
			req.SetBasicAuth(n.User, n.Pass)
			return nil
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		errs = append(errs, err)
		return nil, "", errs
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			errs = append(errs, err)
		}
	}()

	bodyBytes, bodyString, err := n.responseBodyString(resp)
	if err != nil {
		errs = append(errs, err)
		return nil, "", errs
	}
	return bodyBytes, bodyString, errs
}

func (n Nexus) responseBodyString(resp *http.Response) ([]byte, string, error) {
	var bodyBytes []byte
	var bodyString string
	statusOK := resp.StatusCode >= 200 && resp.StatusCode < 300
	if statusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, "", err
		}
		bodyString = string(bodyBytes)
		if err != nil {
			return nil, "", err
		}
	} else {
		return nil, "", fmt.Errorf("responseCode: '%s', message: '%s', URL: '%s'", strconv.Itoa(resp.StatusCode), resp.Status, resp.Request.URL)
	}
	return bodyBytes, bodyString, nil
}