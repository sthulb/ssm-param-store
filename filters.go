package store

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
)

// FilterOption allows you to specify an option to filter SSM Param Store keys by
type FilterOption int

// String returns a string representation of the FilterOption
func (fo FilterOption) String() string {
	opts := []string{
		"BeginsWith",
		"Equals",
		"Recursive",
		"OneLevel",
	}

	if int(fo) < 0 || int(fo) >= len(opts) {
		return ""
	}

	return opts[fo]
}

const (
	FilterOptionBeginsWith FilterOption = 0
	FilterOptionEquals
	FilterOptionRecursive
	FilterOptionOneLevel
)

// FilterFn type
type FilterFn func(f *[]*Filter)

// Filter holds information about a filter
type Filter struct {
	Name   string
	Option FilterOption
	Value  string
}

// AWSFilter converts our internal filter type to one recognised by the AWS SDK
func (f *Filter) AWSFilter() *ssm.ParameterStringFilter {
	return &ssm.ParameterStringFilter{
		Key:    aws.String(f.Name),
		Option: aws.String(f.Option.String()),
		Values: []*string{aws.String(f.Value)},
	}
}

// NewFilters creates a new array of filters, applying each one as they're passed in
func NewFilters(filterFns ...FilterFn) []*Filter {
	filters := []*Filter{}
	for _, f := range filterFns {
		f(&filters)
	}

	return filters
}

// FilterKMSKey filters based on KMS Key ID
func FilterKMSKey(value string, option FilterOption) FilterFn {
	return func(f *[]*Filter) {
		filter := &Filter{
			Name:   "KeyId",
			Option: option,
			Value:  value,
		}

		*f = append(*f, filter)
	}
}

// FilterType filters based on the Param Store type, StringList, String
func FilterType(value string, option FilterOption) FilterFn {
	return func(f *[]*Filter) {
		filter := &Filter{
			Name:   "Type",
			Option: option,
			Value:  value,
		}

		*f = append(*f, filter)
	}
}
