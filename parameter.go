package store

import (
	"errors"
	"time"
)

var (
	// ErrExpiredParameter thrown when the value of a param has expired
	ErrExpiredParameter = errors.New("Parameter expired")

	// ErrNoRefreshFn for when the parameter has no refresh func, but is asked to refresh
	ErrNoRefreshFn = errors.New("No refresh function supplied")
)

// RefreshFn creates a callback for when a param value needs to be refreshed
type RefreshFn func(p *Parameter) (interface{}, error)

// ParameterOptionFn type allows options to be specified for the parameter, such as auto refresh or expiry
type ParameterOptionFn func(p *Parameter)

// ValueExpires tells a parameter to refresh based on an expiry duration
func ValueExpires(expiryTime time.Duration) ParameterOptionFn {
	return func(p *Parameter) {
		p.Expires = true
		p.Expiry = expiryTime
	}
}

// AutoRefreshValue will allow a parameter value to automatically refresh
func AutoRefreshValue(autoRefresh bool) ParameterOptionFn {
	return func(p *Parameter) {
		p.AutoRefresh = autoRefresh
		p.lastRefresh = time.Now()
	}
}

// Parameter data structure
type Parameter struct {
	Name  string
	Value interface{}

	Expires bool
	Expiry  time.Duration

	AutoRefresh bool
	RefreshFn   RefreshFn

	lastRefresh time.Time
}

// NewParameter creates a new parameter based on options
func NewParameter(name string, value interface{}, opts ...ParameterOptionFn) *Parameter {
	param := &Parameter{
		Expires: false,

		Name:  name,
		Value: value,
	}

	for _, o := range opts {
		if o == nil {
			break
		}

		o(param)
	}

	return param
}

func (p *Parameter) UpdateValue(v interface{}) {
	p.Value = v
	p.lastRefresh = time.Now()
}

func (p *Parameter) StringValue() (string, error) {
	if p.Expires {
		return "", p.RefreshValue()
	}

	value := p.Value.(string)

	return value, nil
}

// RefreshValue checks if a value is expired and refreshes it if required
func (p *Parameter) RefreshValue() error {
	lastSet := p.lastRefresh
	expiry := p.Expiry

	trueExpiry := lastSet.Add(expiry)
	if trueExpiry.After(time.Now()) {
		return nil
	}

	if p.AutoRefresh && p.RefreshFn == nil {
		return ErrNoRefreshFn
	}

	if p.AutoRefresh && p.RefreshFn != nil {
		_, err := p.RefreshFn(p)
		if err != nil {
			return err
		}

		p.lastRefresh = time.Now()

		return nil
	}

	return ErrExpiredParameter
}
