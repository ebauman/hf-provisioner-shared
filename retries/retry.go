package retries

import (
	"fmt"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"time"
)

type GenericRetry struct {
	Name              string      `json:"action"`
	MaxRetries        int         `json:"maxRetries"`
	Attempts          int         `json:"attempts"`
	BackoffTime       v1.Duration `json:"backoffTime"`
	LastAttemptTime   v1.Time     `json:"lastAttemptTime"`
	LastAttemptResult RetryResult `json:"lastAttemptResult"`
}

type RetryResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type GetAnnotations interface {
	GetAnnotations() map[string]string
}

type SetAnnotations interface {
	SetAnnotations(map[string]string)
}

type Retrier interface {
	GetAnnotations
	SetAnnotations
}

func New(name string, maxRetries int) GenericRetry {
	return GenericRetry{
		Name:       name,
		MaxRetries: maxRetries,
	}
}

func (r GenericRetry) Success(obj Retrier) {
	r.SetAttempt(obj, true, "")
}

func (r GenericRetry) Failure(obj Retrier) {
	r.SetAttempt(obj, false, "")
}

func (r GenericRetry) Successf(obj Retrier, message string, args ...any) {
	r.SetAttempt(obj, true, message, args...)
}

func (r GenericRetry) Failuref(obj Retrier, message string, args ...any) {
	r.SetAttempt(obj, false, message, args...)
}

func (r GenericRetry) SetAttempt(obj Retrier, success bool, message string, args ...any) {
	retryInstance, err := findOrCreateRetry(obj, r)
	if err != nil {
		return
	}

	retryInstance.Attempts++
	retryInstance.LastAttemptTime = v1.Now()
	retryInstance.LastAttemptResult = RetryResult{
		Success: success,
		Message: fmt.Sprintf(message, args...),
	}
}

func (r GenericRetry) ExceededRetries(obj Retrier) (exceeded bool, ok bool) {
	retryInstance, err := findOrCreateRetry(obj, r)

	if err != nil {
		logrus.Errorf("error executing exceededretries: %w", err)
		return false, false
	}

	if retryInstance.MaxRetries >= retryInstance.Attempts {
		return true, true
	}

	return false, true
}

func (r GenericRetry) CanRetry(obj Retrier) bool {
	// does this retry exist on the object?
	retryInstance, err := findOrCreateRetry(obj, r)
	if err != nil {
		logrus.Errorf("error executing canretry: %w", err)
		return false
	}

	if retryInstance.MaxRetries >= retryInstance.Attempts {
		return false
	}

	if r.LastAttemptTime.Time.Add(r.BackoffTime.Duration).Before(time.Now()) {
		// if the last time we attempted this, plus the backoff duration, has not yet passed
		return false
	}

	return true
}

func findOrCreateRetry(obj Retrier, r GenericRetry) (*GenericRetry, error) {
	annos := obj.GetAnnotations()

	// if the retry is not found, create it
	annotationJson, ok := annos[r.Name]
	if !ok {
		annoJson, err := json.Marshal(r)
		if err != nil {
			return nil, fmt.Errorf("error marshalling genericretry: %s", err)
		}

		if len(annos) == 0 {
			annos = map[string]string{}
		}
		annos[r.Name] = string(annoJson)

		obj.SetAnnotations(annos)

		return &r, nil
	}

	// the retry was found! return it after unmarshalling
	var returnRetry = &GenericRetry{}
	err := json.Unmarshal([]byte(annotationJson), returnRetry)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling genericretry: %w", err)
	}

	return returnRetry, nil
}
