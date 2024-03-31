package turnstile

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrCodeMissingInputSecret   = errors.New("missing-input-secret")
	ErrCodeInvalidInputSecret   = errors.New("invalid-input-secret")
	ErrCodeMissingInputResponse = errors.New("missing-input-response")
	ErrCodeInvalidInputResponse = errors.New("invalid-input-response")
	ErrCodeBadRequest           = errors.New("bad-request")
	ErrCodeTimeoutOrDuplicate   = errors.New("timeout-or-duplicate")
	ErrCodeInternalError        = errors.New("internal-error")
)

type Request struct {
	Secret string
	Token  string
	IP     string
}

type Response struct {
	Success   bool      `json:"success,omitempty"`
	Errors    []error   `json:"errors,omitempty"`
	Timestamp time.Time `json:"challenge_ts,omitempty"`
	Hostname  string    `json:"hostname,omitempty"`
}

type actualResponse struct {
	Success   bool     `json:"success"`
	ErrCodes  []string `json:"error-codes"`
	Timestamp string   `json:"challenge_ts"`
	Hostname  string   `json:"hostname"`
}

func (a actualResponse) ToClientResponse() (*Response, error) {
	t, err := time.Parse(time.RFC3339, a.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse challenge timestamp: %w", err)
	}

	errs := make([]error, len(a.ErrCodes))
	for idx, e := range a.ErrCodes {
		switch e {
		case ErrCodeMissingInputSecret.Error():
			errs[idx] = ErrCodeMissingInputSecret
		case ErrCodeInvalidInputSecret.Error():
			errs[idx] = ErrCodeInvalidInputSecret
		case ErrCodeMissingInputResponse.Error():
			errs[idx] = ErrCodeMissingInputResponse
		case ErrCodeInvalidInputResponse.Error():
			errs[idx] = ErrCodeInvalidInputResponse
		case ErrCodeBadRequest.Error():
			errs[idx] = ErrCodeBadRequest
		case ErrCodeTimeoutOrDuplicate.Error():
			errs[idx] = ErrCodeTimeoutOrDuplicate
		case ErrCodeInternalError.Error():
			errs[idx] = ErrCodeInternalError
		}
	}

	return &Response{
		Success:   a.Success,
		Errors:    errs,
		Timestamp: t,
		Hostname:  a.Hostname,
	}, nil
}
