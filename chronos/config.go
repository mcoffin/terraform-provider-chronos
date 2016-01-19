package chronos

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
)

type Config struct {
	Url *url.URL
	UserInfo *url.Userinfo
}

var (
	ErrJobDoesNotExist = errors.New("the job does not exist")
)

func (c Config) GetCreateUrl() string {
	return c.Url.String() + "/scheduler/iso8601"
}

func (c Config) GetJobsUrl() string {
	return c.Url.String() + "/scheduler/jobs"
}

func (c Config) CreateRequest(method, u, contentType string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return nil, err
	}
	password, _ := c.UserInfo.Password()
	req.SetBasicAuth(c.UserInfo.Username(), password)

	if len(contentType) > 0 {
		req.Header.Add("Content-Type", contentType)
	}

	return req, nil
}

func (c Config) getJob(jobName string) (*job, error) {
	req, err := c.CreateRequest("GET", c.GetJobsUrl(), "", nil)
	if err != nil {
		return nil, err
	}

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(req.Body)
	var js []job
	err = decoder.Decode(&js)
	if err != nil {
		return nil, err
	}

	for _, j := range js {
		if j.Name == jobName {
			return  &j, nil
		}
	}

	return nil, ErrJobDoesNotExist
}
