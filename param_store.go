package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/aws-sdk-go/service/ssm"
)

type ParamStore struct {
	client *ssm.SSM
	path   string

	RequestRetries int
	RequestTimeout time.Duration

	// CacheTimeout int
}

func New() *ParamStore {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{MaxRetries: aws.Int(2)},
	}))

	svc := ssm.New(sess)

	ps := &ParamStore{
		client:         svc,
		RequestTimeout: time.Second,
	}

	return ps
}

func (ps *ParamStore) Retries(retries int) {
	ps.RequestRetries = retries
}

func (ps *ParamStore) Timeout(timeout time.Duration) {
	ps.RequestTimeout = timeout
}

func (ps *ParamStore) Param(key string) (*Parameter, error) {
	svc := ps.client

	ctx, cancel := context.WithTimeout(context.Background(), ps.RequestTimeout)
	defer cancel()

	out, err := svc.GetParameterWithContext(ctx, &ssm.GetParameterInput{
		Name:           aws.String(key),
		WithDecryption: aws.Bool(true),
	})

	if err != nil {
		return nil, err
	}

	param := NewParameter(
		*out.Parameter.Name,
		*out.Parameter.Value,
	)

	return param, nil
}

func (ps *ParamStore) ParamsByPath(p string, filters []*Filter) (interface{}, error) {
	svc := ps.client

	ctx, cancel := context.WithTimeout(context.Background(), ps.RequestTimeout)
	defer cancel()

	paramFilters := []*ssm.ParameterStringFilter{}
	for _, f := range filters {
		paramFilters = append(paramFilters, f.AWSFilter())
	}

	out, err := svc.GetParametersByPathWithContext(ctx, &ssm.GetParametersByPathInput{
		Path:           aws.String(p),
		WithDecryption: aws.Bool(true),
	})

	if err != nil {
		return nil, err
	}

	return out, nil
}

func Path(p ...string) string {
	return fmt.Sprintf("/%s", strings.Join(p, "/"))
}
