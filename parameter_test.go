package store

import (
	"testing"
	"time"
)

func TestParameter(t *testing.T) {
	// t.Run("New Parameter", func(t *testing.T) {
	// 	p := NewParameter("foo", nil)
	// 	if p == nil {
	// 		t.Fatalf("Unable to create a new parameter")
	// 	}

	// 	if fmt.Sprintf("%T", p) != "*store.Parameter" {
	// 		t.Fatalf("Returned the wrong type: %v", fmt.Sprintf("%T", p))
	// 	}
	// })

	// t.Run("New Parameter With Opts", func(t *testing.T) {
	// 	expiry := time.Second * 10

	// 	p := NewParameter("foo", nil, ValueExpires(expiry))

	// 	if p.Expires != true || expiry != p.Expiry {
	// 		t.Fatalf("Expiry time isn't correct: %v", p.Expires)
	// 	}
	// })

	// t.Run("Refresh Parameter", func(t *testing.T) {
	// 	expiry := time.Second * 1

	// 	p := NewParameter("foo", nil, ValueExpires(expiry), AutoRefreshValue(true))

	// 	if p.Expires != true || expiry != p.Expiry {
	// 		t.Fatalf("Expiry time isn't correct: %v", p.Expires)
	// 	}
	// })

	// t.Run("String Value", func(t *testing.T) {
	// 	value := "bar"
	// 	p := NewParameter("foo", value)

	// 	retValue, _ := p.StringValue()
	// 	if retValue != value {
	// 		t.Fatalf("Error getting the value")
	// 	}
	// })

	// t.Run("String Array Value", func(t *testing.T) {
	// 	value := []string{"foo", "bar"}
	// 	p := NewParameter("foo", value)

	// 	retValue, _ := p.StringListValue()
	// 	if !reflect.DeepEqual(retValue, value) {
	// 		t.Fatalf("Error getting the value")
	// 	}
	// })

	t.Run("Check Expiry in Future", func(t *testing.T) {
		now := time.Now()
		future := now.Add(time.Hour)

		p := NewParameter("foo", "test")
		p.Expires = true
		p.lastRefresh = future

		if err := p.RefreshValue(); err != nil {
			t.Fatal("Dates in the future should not error")
		}
	})

	t.Run("Check Expiry in Past", func(t *testing.T) {
		past, _ := time.Parse("Mon, 01/02/06, 03:04PM", "Thu, 05/19/11, 10:47PM")

		p := NewParameter("foo", "test")
		p.Expires = true
		p.lastRefresh = past

		if err := p.RefreshValue(); err == nil {
			t.Fatal("Dates in the past should error")
		}
	})

	t.Run("Check Autorefresh enabled, but with no func", func(t *testing.T) {
		past, _ := time.Parse("Mon, 01/02/06, 03:04PM", "Thu, 05/19/11, 10:47PM")

		p := NewParameter("foo", "test", AutoRefreshValue(true))
		p.Expires = true
		p.lastRefresh = past

		if err := p.RefreshValue(); err != ErrNoRefreshFn {
			t.Fatalf("Parameters that have no refresh function should fail when auto refresh is enabled: %v", err)
		}
	})

	t.Run("Check Autorefresh enabled, with func", func(t *testing.T) {
		past, _ := time.Parse("Mon, 01/02/06, 03:04PM", "Thu, 05/19/11, 10:47PM")

		p := NewParameter("foo", "test", AutoRefreshValue(true))
		p.RefreshFn = func(p *Parameter) (interface{}, error) {
			return nil, nil
		}

		p.Expires = true
		p.lastRefresh = past

		if err := p.RefreshValue(); err != nil {
			t.Fatal("Parameters that have no refresh function should fail when auto refresh is enabled")
		}
	})
}
