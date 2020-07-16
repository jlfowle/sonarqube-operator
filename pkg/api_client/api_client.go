package api_client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

type APIProvider interface {
	New(URL string) APIReader
}

type APIReader interface {
	Ping() error
	Status() (*Status, error)
	Upgrades() (*Upgrades, error)
}

type APIClient struct {
	URL    string
	Client *http.Client
}

func (r *APIClient) New(URL string) APIReader {
	var netTransport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	return &APIClient{
		URL: URL,
		Client: &http.Client{
			Timeout:   time.Second * 10,
			Transport: netTransport,
		},
	}
}

func (r *APIClient) Ping() error {
	res, err := r.get("system", "ping")
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("non 200 error code returned")
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if string(body) != "pong" {
		return fmt.Errorf("incorrect response got %s expected pond", string(body))
	}

	return nil
}

func (r *APIClient) Status() (*Status, error) {
	output := &Status{}
	res, err := r.get("system", "status")
	if err != nil {
		return output, err
	}
	if res.StatusCode != 200 {
		return output, fmt.Errorf("non 200 error code returned")
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return output, err
	}

	err = json.Unmarshal(body, output)
	if err != nil {
		return output, err
	}

	return output, nil
}

func (r *APIClient) Upgrades() (*Upgrades, error) {
	output := &Upgrades{}
	res, err := r.get("system", "upgrades")
	if err != nil {
		return output, err
	}
	if res.StatusCode != 200 {
		return output, fmt.Errorf("non 200 error code returned")
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return output, err
	}

	err = json.Unmarshal(body, output)
	if err != nil {
		return output, err
	}

	return output, nil
}

func (r *APIClient) get(domain, object string) (*http.Response, error) {
	url := fmt.Sprintf("%s/api/%s/%s", r.URL, domain, object)
	return r.Client.Get(url)
}
