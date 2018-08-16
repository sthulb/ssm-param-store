package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

var (
	ErrNoParameters = errors.New("No Parameters found")
)

type ParamStore struct {
	client ssmiface.SSMAPI
	path   string

	RequestRetries int
	RequestTimeout time.Duration
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

func (ps *ParamStore) Param(key string, opts ...ParameterOptionFn) (*Parameter, error) {
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
		opts...,
	)

	param.RefreshFn = func(p *Parameter) (interface{}, error) {
		refreshCtx, cancel := context.WithTimeout(context.Background(), ps.RequestTimeout)
		defer cancel()

		out, err := svc.GetParameterWithContext(refreshCtx, &ssm.GetParameterInput{
			Name:           aws.String(key),
			WithDecryption: aws.Bool(true),
		})

		if err != nil {
			return nil, err
		}

		p.UpdateValue(out.Parameter.Value)

		return p.Value, nil
	}

	return param, nil
}

func (ps *ParamStore) ParamsByPath(p string, filters []*Filter, opts ...ParameterOptionFn) ([]*Parameter, error) {
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

	if out.Parameters == nil || len(out.Parameters) == 0 {
		return nil, ErrNoParameters
	}

	params := []*Parameter{}
	for _, awsParam := range out.Parameters {
		param := NewParameter(*awsParam.Name, *awsParam.Value, nil)
		param.RefreshFn = func(p *Parameter) (interface{}, error) {
			refreshCtx, cancel := context.WithTimeout(context.Background(), ps.RequestTimeout)
			defer cancel()

			out, err := svc.GetParameterWithContext(refreshCtx, &ssm.GetParameterInput{
				Name:           aws.String(p.Name),
				WithDecryption: aws.Bool(true),
			})

			if err != nil {
				return nil, err
			}

			p.UpdateValue(out.Parameter.Value)

			return p.Value, nil
		}

		params = append(params, param)
	}

	return params, nil
}

func Path(p ...string) string {
	return fmt.Sprintf("/%s", strings.Join(p, "/"))
}
