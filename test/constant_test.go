package pine_test

import (
	"fmt"
	"testing"
	"time"

	. "github.com/xpt-nl/pine"
)

func TestConstantInit(t *testing.T) {
	opts := SeriesOpts{
		Interval: 300,
		Max:      100,
	}

	now := time.Now()
	fivemin := now.Add(5 * time.Minute)
	constant := NewConstant(5.0)

	data := []OHLCV{
		{
			O: 14,
			H: 15,
			L: 13,
			C: 14,
			V: 131,
			S: now,
		},
		{
			O: 13,
			H: 18,
			L: 10,
			C: 15,
			V: 12,
			S: fivemin,
		},
	}
	s, err := NewSeries(data, opts)
	if err != nil {
		t.Fatal(fmt.Errorf("error init series: %w", err))
	}
	if err := s.AddIndicator("constant", constant); err != nil {
		t.Fatal(fmt.Errorf("expected constant to not error but errored: %w", err))
	}
	v := s.GetValueForInterval(now)
	if *(v.Indicators["constant"]) != 5.0 {
		t.Errorf("expected 5.0 but got %+v", *(v.Indicators["constant"]))
	}
	v = s.GetValueForInterval(fivemin)
	if *(v.Indicators["constant"]) != 5.0 {
		t.Errorf("expected 5.0 but got %+v", *(v.Indicators["constant"]))
	}
}
