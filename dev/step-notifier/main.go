package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	_ "log"
	"net/http"
	"os"
	"time"

	"github.com/sourcegraph/log"
)

var ErrRequestBody = fmt.Errorf("failed to read body from request")
var ErrJSONUnmarshall = fmt.Errorf("failed to unmarshal body")
var ErrInvalidToken = fmt.Errorf("buildkite token is invalid")
var ErrInvalidHeader = fmt.Errorf("Header of request is invalid")
var ErrUnwantedEvent = fmt.Errorf("Unwanted event received")

type BuildTrackingServer struct {
	logger  log.Logger
	store   *BuildStore
	bkToken string
	slack   *SlackClient
}

func NewBuildTrackingServer(logger log.Logger, buildkiteToken, slackToken string) (*BuildTrackingServer, error) {
	return &BuildTrackingServer{
		logger:  logger,
		store:   NewBuildStore(logger),
		bkToken: buildkiteToken,
		slack:   NewSlackClient(logger, slackToken),
	}, nil
}

func (s *BuildTrackingServer) processBuildkiteRequest(req *http.Request, token string) (*BuildEvent, error) {
	h, ok := req.Header["X-Buildkite-Token"]
	if !ok || len(h) == 0 {
		return nil, ErrInvalidToken
	} else if h[0] != token {
		return nil, ErrInvalidToken
	}

	h, ok = req.Header["X-Buildkite-Event"]
	if !ok || len(h) == 0 {
		return nil, ErrInvalidHeader
	}

	eventName := h[0]
	s.logger.Debug("receied event", log.String("eventName", eventName))

	var event BuildEvent
	err := readBody(s.logger, req, &event)
	if errors.Is(err, ErrRequestBody) {
		s.logger.Error("failed to read body of request", log.Error(err))
		return nil, ErrRequestBody
	} else if errors.Is(err, ErrJSONUnmarshall) {
		s.logger.Error("failed to unmarshall body of request", log.Error(err))
		return nil, ErrJSONUnmarshall
	}

	return &event, nil
}

// handleEvent handles an event received from the http listener. A event is valid when:
// - Has the correct headers from Buildkite
// - On of the following events
//   * job.finished
//   * build.finished
// - Has valid JSON
// Note that if we received an unwanted event ie. the event is not "job.finished" or "build.finished" we respond with a 200 OK regardless.
// Once all the conditions are met, the event is processed in a go routine with `processEvent`
func (s *BuildTrackingServer) handleEvent(w http.ResponseWriter, req *http.Request) {
	event, err := s.processBuildkiteRequest(req, s.bkToken)

	switch err {
	case ErrInvalidToken:
	case ErrInvalidHeader:
		w.WriteHeader(http.StatusBadRequest)
		return
	case ErrUnwantedEvent:
		w.WriteHeader(http.StatusOK)
		return
	case ErrRequestBody:
		w.WriteHeader(http.StatusBadRequest)
		return
	case ErrJSONUnmarshall:
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	s.logger.Info("processing event", log.String("eventName", event.Event), log.Int("buildNumber", event.BuildNumber()), log.String("JobName", event.JobName()))
	go s.processEvent(event)
	w.WriteHeader(http.StatusOK)
}

func readBody[T any](logger log.Logger, req *http.Request, target T) error {
	data, err := ioutil.ReadAll(req.Body)
	logger.Debug("read body of request", log.String("data", string(data)))
	if err != nil {
		logger.Error("failed to read request body", log.Error(err))
		return ErrRequestBody
	}

	err = json.Unmarshal(data, &target)
	if err != nil {
		logger.Error("failed to unmarshall request body", log.Error(err))
		return ErrJSONUnmarshall
	}

	return nil
}

// notifyIfFailed sends a notification over slack if the provided build has failed. If the build is successful not notifcation is sent
func (s *BuildTrackingServer) notifyIfFailed(build *Build) error {
	if build.HasFailed() {
		s.logger.Info("detected failed build - sending notification", log.Int("buildNumber", *build.Number))
		return s.slack.sendNotification(build)
	}

	s.logger.Info("build successful", log.Int("buildNumber", *build.Number))
	return nil
}

func (s *BuildTrackingServer) startOldBuildCleaner(every, window time.Duration) func() {
	ticker := time.NewTicker(every)
	done := make(chan interface{})

	go func() {
		for {
			select {
			case <-ticker.C:
				{
					oldBuilds := make([]int, 0)
					now := time.Now()
					for _, b := range s.store.AllFinishedBuilds() {
						finishedAt := *b.FinishedAt
						delta := now.Sub(finishedAt.Time)
						if delta >= window {
							oldBuilds = append(oldBuilds, *b.Number)
						}
					}
					s.logger.Info("deleting old builds", log.Int("oldBuildCount", len(oldBuilds)))
					s.store.DelByBuildNumber(oldBuilds...)
				}
			case <-done:
				{
					ticker.Stop()
					return
				}
			}
		}
	}()

	return func() { done <- nil }
}

// processEvent processes a BuildEvent received from Buildkite. If the event is for a `build.finished` event we get the
// full build which includes all recorded jobs for the build and send a notification.
// processEvent delegates the decision to actually send a notifcation
func (s *BuildTrackingServer) processEvent(event *BuildEvent) {
	// Build number is required!
	if event.Build.Number == nil {
		return
	}

	s.store.Add(event)
	if event.IsBuildFinished() {
		build := s.store.GetByBuildNumber(event.BuildNumber())
		if err := s.notifyIfFailed(build); err != nil {
			s.logger.Error("failed to send notification for build", log.Int("buildNumber", event.BuildNumber()), log.Error(err))
		}
	}
}

// Server starts the http server and listens for buildkite build events to be sent on the route "/buildkite"
func (s *BuildTrackingServer) Serve() error {
	http.HandleFunc("/buildkite", s.handleEvent)
	s.logger.Info("listening on :8080")
	return http.ListenAndServe(":8080", nil)
}

func main() {
	slackToken := os.Getenv("SLACK_TOKEN")
	if slackToken == "" {
		panic("SLACK_TOKEN cannot be empty")
	}
	buildkiteToken := os.Getenv("BK_WEBHOOK_TOKEN")
	if buildkiteToken == "" {
		panic("BK_WEBHOOK_TOKEN cannot be empty")
	}

	sync := log.Init(log.Resource{
		Name:      "BuildTracker",
		Namespace: "CI",
	})
	defer sync.Sync()

	logger := log.Scoped("server", "Server that tracks completed builds")
	server, err := NewBuildTrackingServer(logger, buildkiteToken, slackToken)
	if err != nil {
		log.Scoped("BuildTracker.main", "main entrypoint for BuildTracker").Fatal(
			"failed to create BuildTracking server", log.Error(err),
		)
	}

	stopFn := server.startOldBuildCleaner(5*time.Minute, 24*time.Hour)
	defer stopFn()
	if err := server.Serve(); err != nil {
		server.logger.Fatal("server exited with error", log.Error(err))
	}
}
