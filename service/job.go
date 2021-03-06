package service

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/NYTimes/gizmo/web"
	"github.com/NYTimes/video-captions-api/providers"
	log "github.com/Sirupsen/logrus"
	"github.com/satori/go.uuid"
)

// GetJobs returns all the Jobs associated with a ParentID
func (s *CaptionsService) GetJobs(r *http.Request) (int, interface{}, error) {
	parentID := web.Vars(r)["id"]
	jobs, err := s.client.GetJobs(parentID)
	if err != nil {
		return http.StatusNotFound, nil, err
	}
	return http.StatusOK, jobs, nil
}

// GetJob returns a Job given its ID
func (s *CaptionsService) GetJob(r *http.Request) (int, interface{}, error) {
	id := web.Vars(r)["id"]
	// TODO: on the 3play client, we should look at the errors field and check for not_found errors at least
	job, err := s.client.GetJob(id)
	if err != nil {
		return http.StatusNotFound, nil, err
	}
	return http.StatusOK, job, nil
}

// CreateJob create a Job
func (s *CaptionsService) CreateJob(r *http.Request) (int, interface{}, error) {
	requestLogger := logger.WithFields(log.Fields{
		"Handler": "CreateJob",
		"Method":  r.Method,
		"URI":     r.RequestURI,
	})
	var job providers.Job
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		requestLogger.Error("Could not read request body", err)
		return http.StatusBadRequest, nil, err
	}
	err = json.Unmarshal(data, &job)

	if err != nil {
		requestLogger.Error("Could not create job from request body", err)
		return http.StatusBadRequest, nil, errors.New("Malformed parameters")
	}

	mediaURL := job.MediaURL

	if mediaURL == "" {
		requestLogger.Error("Tried to create a job without a media url", err)
		return http.StatusBadRequest, nil, errors.New("Please provide a media_url")
	}

	job.ParentID = job.ID
	job.ID = uuid.NewV4().String()
	job, err = s.client.DispatchJob(job)
	if err != nil {
		return http.StatusInternalServerError, err, err
	}

	return http.StatusCreated, job, nil
}
