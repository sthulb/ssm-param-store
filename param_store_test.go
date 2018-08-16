package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"

	"github.com/aws/aws-sdk-go/service/ssm"
)

var (
	past, _ = time.Parse("Mon, 01/02/06, 03:04PM", "Thu, 05/19/11, 10:47PM")
)

func TestParamStoreAttributeFuncs(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		ps := New()
		if ps == nil {
			t.Fatalf("Unable to create param store")
		}
	})

	t.Run("Retries", func(t *testing.T) {
		ps := New()
		ps.Retries(1)
		if ps.RequestRetries != 1 {
			t.Fatalf("Request Retries isn't 1")
		}
	})

	t.Run("Timeout", func(t *testing.T) {
		ps := New()
		ps.Timeout(time.Second)
		if ps.RequestTimeout != time.Second {
			t.Fatalf("Request Retries isn't 1")
		}
	})
}

func TestParam(t *testing.T) {
	t.Run("Get Param Success", func(t *testing.T) {
		var (
			name  = "foo"
			value = "bar"
		)

		mockClient := &mockSSMClient{}
		mockClient.paramFn = func(ctx context.Context, input *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
			param := &ssm.Parameter{
				Name:  aws.String(name),
				Value: aws.String(value),
			}

			out := &ssm.GetParameterOutput{
				Parameter: param,
			}

			return out, nil
		}

		ps := New()
		ps.client = mockClient

		out, _ := ps.Param(name)
		outValue, err := out.StringValue()
		if err != nil {
			t.Fatalf("Error returning value: %v", err)
		}

		if out.Name != name && outValue != value {
			t.Fatalf("Invalid parameter returned")
		}
	})

	t.Run("Get Param Error", func(t *testing.T) {
		var (
			name = "foo"
		)

		mockClient := &mockSSMClient{}
		mockClient.paramFn = func(ctx context.Context, input *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
			return nil, errors.New("Unable to get param")
		}

		ps := New()
		ps.client = mockClient

		_, err := ps.Param(name)

		if err == nil {
			t.Fatalf("We should have had an error")
		}
	})

	t.Run("Get Param Error Timeout", func(t *testing.T) {
		var (
			name = "foo"
		)

		mockClient := &mockSSMClient{}
		mockClient.paramFn = func(ctx context.Context, input *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		ps := New()
		ps.client = mockClient

		_, err := ps.Param(name)

		if err == nil {
			t.Fatalf("We should have had an error")
		}
	})

	t.Run("Get Param Refresh", func(t *testing.T) {
		var (
			name     = "foo"
			value    = "bar"
			newValue = "baz"
		)

		mockClient := &mockSSMClient{}
		mockClient.paramFn = func(ctx context.Context, input *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
			param := &ssm.Parameter{
				Name:  aws.String(name),
				Value: aws.String(value),
			}

			out := &ssm.GetParameterOutput{
				Parameter: param,
			}

			return out, nil
		}

		ps := New()
		ps.client = mockClient

		out, _ := ps.Param(name)
		out.lastRefresh = past

		mockClient.paramFn = func(ctx context.Context, input *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
			param := &ssm.Parameter{
				Name:  aws.String(name),
				Value: aws.String(newValue),
			}

			out := &ssm.GetParameterOutput{
				Parameter: param,
			}

			return out, nil
		}

		outValue, err := out.StringValue()
		if err != nil {
			t.Fatalf("Error returning value: %v", err)
		}

		if out.Name != name && outValue != newValue {
			t.Fatalf("Invalid parameter returned")
		}
	})
}

func TestParamsByPath(t *testing.T) {

	var (
		paramPath  = "/aws/ssm/param"
		parameters = []*ssm.Parameter{
			&ssm.Parameter{
				Name:  aws.String("/aws/ssm/param"),
				Value: aws.String("foo"),
			},
		}
	)

	t.Run("Success", func(t *testing.T) {
		mockClient := &mockSSMClient{}
		mockClient.paramsByPathFn = func(ctx context.Context, input *ssm.GetParametersByPathInput) (*ssm.GetParametersByPathOutput, error) {
			out := &ssm.GetParametersByPathOutput{
				Parameters: parameters,
			}

			return out, nil
		}

		ps := New()
		ps.client = mockClient

		out, err := ps.ParamsByPath(paramPath, nil)
		if err != nil {
			t.Fatalf("error: %v", err)
		}

		if len(out) != 1 {
			t.Fatalf("Not enough parameters returned")
		}

		if out[0].Name != paramPath {
			t.Fatalf("Wrong param path found")
		}
	})

	t.Run("Error", func(t *testing.T) {
		mockClient := &mockSSMClient{}
		mockClient.paramsByPathFn = func(ctx context.Context, input *ssm.GetParametersByPathInput) (*ssm.GetParametersByPathOutput, error) {
			return nil, errors.New("an error")
		}

		ps := New()
		ps.client = mockClient

		_, err := ps.ParamsByPath(paramPath, nil, nil)
		if err == nil {
			t.Fatalf("no error: %v", err)
		}
	})

	t.Run("Error No Params", func(t *testing.T) {
		mockClient := &mockSSMClient{}
		mockClient.paramsByPathFn = func(ctx context.Context, input *ssm.GetParametersByPathInput) (*ssm.GetParametersByPathOutput, error) {
			out := &ssm.GetParametersByPathOutput{}
			return out, nil
		}

		ps := New()
		ps.client = mockClient

		_, err := ps.ParamsByPath(paramPath, nil, nil)
		if err == nil {
			t.Fatalf("no error: %v", err)
		}
	})
}

func TestPath(t *testing.T) {
	if Path("foo", "bar") != "/foo/bar" {
		t.Fatalf("Incorrect path returned")
	}
}

type mockSSMClient struct {
	ssmiface.SSMAPI

	paramFn        func(ctx context.Context, input *ssm.GetParameterInput) (*ssm.GetParameterOutput, error)
	paramsByPathFn func(ctx context.Context, input *ssm.GetParametersByPathInput) (*ssm.GetParametersByPathOutput, error)
}

func (m *mockSSMClient) GetParameterWithContext(ctx aws.Context, input *ssm.GetParameterInput, opts ...request.Option) (*ssm.GetParameterOutput, error) {
	return m.paramFn(ctx, input)
}

func (m *mockSSMClient) GetParametersByPathWithContext(ctx aws.Context, input *ssm.GetParametersByPathInput, opts ...request.Option) (*ssm.GetParametersByPathOutput, error) {
	return m.paramsByPathFn(ctx, input)
}
