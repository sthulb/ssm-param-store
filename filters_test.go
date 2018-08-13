package store

import (
	"fmt"
	"testing"
)

func TestFilterOption(t *testing.T) {
	t.Run("New Filter Option", func(t *testing.T) {
		fo := FilterOptionBeginsWith
		if fo.String() != "BeginsWith" {
			t.Fatalf("Unexpected value: %v", fo.String())
		}
	})

	t.Run("Invalid Filter Option", func(t *testing.T) {
		fo := FilterOption(4)
		if fo.String() != "" {
			t.Fatalf("Invalid filter option has value")
		}
	})
}

func TestFilters(t *testing.T) {
	t.Run("New Filters", func(t *testing.T) {
		filters := NewFilters(
			FilterKMSKey("TestKey", FilterOptionEquals),
			FilterType("TestType", FilterOptionEquals),
		)

		if len(filters) != 2 {
			t.Fatalf("Incorrect number of filters: %d", len(filters))
		}
	})

	t.Run("Filter", func(t *testing.T) {
		filterFn := func(f *[]*Filter) {
			filter := &Filter{
				Name:   "Test",
				Option: FilterOptionBeginsWith,
				Value:  "Test",
			}

			*f = append(*f, filter)
		}

		filters := []*Filter{}
		filterFn(&filters)

		if len(filters) != 1 {
			t.Fatalf("Too many filters in the slice, got: %d", len(filters))
		}

		ffs := filters[0]
		if ffs.Name != "Test" && ffs.Value != "Test" {
			t.Fatalf("Unknown filter found")
		}
	})

	t.Run("Filter AWS Filter Conversion", func(t *testing.T) {
		filters := []*Filter{}
		filterFn := FilterKMSKey("TestKey", FilterOptionEquals)

		filterFn(&filters)

		awsFilter := filters[0].AWSFilter()

		if fmt.Sprintf("%T", awsFilter) != "*ssm.ParameterStringFilter" {
			t.Fatalf("Incorrect type found: %v", fmt.Sprintf("%T", awsFilter))
		}

		if *awsFilter.Key != "KeyId" {
			t.Fatalf("Invalid key found: %v", *awsFilter.Key)
		}
	})
}
