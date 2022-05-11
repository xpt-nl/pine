package pine_test

import (
	"fmt"
	"math"
	"testing"
	"time"

	. "github.com/xpt-nl/pine"
)

func TestStdDev(t *testing.T) {
	itvl := 300
	opts := SeriesOpts{
		Interval: itvl,
		Max:      32,
	}
	name := "stddev"
	sdTests := struct {
		candles  []OHLCV
		expected []*Interval
	}{
		candles: []OHLCV{
			{C: 52.22},
			{C: 52.78},
			{C: 53.02},
			{C: 53.67},
			{C: 53.67},
			{C: 53.74},
			{C: 53.45},
			{C: 53.72},
			{C: 53.39},
			{C: 52.51},
			{C: 52.32},
			{C: 51.45},
			{C: 51.60},
			{C: 52.43},
			{C: 52.47},
			{C: 52.91},
			{C: 52.07},
			{C: 53.12},
			{C: 52.77},
			{C: 52.73},
			{C: 52.09},
			{C: 53.19},
			{C: 53.73},
			{C: 53.87},
			{C: 53.85},
			{C: 53.88},
			{C: 54.08},
			{C: 54.14},
			{C: 54.50},
			{C: 54.30},
			{C: 54.40},
			{C: 54.16},
		},
		expected: []*Interval{
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			{Value: 0.523018},
			{Value: 0.505411},
			{Value: 0.730122},
			{Value: 0.857364},
			{Value: 0.833642},
			{Value: 0.788707},
			{Value: 0.716251},
			{Value: 0.675498},
			{Value: 0.584679},
			{Value: 0.507870},
			{Value: 0.518353},
			{Value: 0.526061},
			{Value: 0.480964},
			{Value: 0.490176},
			{Value: 0.578439},
			{Value: 0.622905},
			{Value: 0.670093},
			{Value: 0.622025},
			{Value: 0.661064},
			{Value: 0.690358},
			{Value: 0.651152},
			{Value: 0.360466},
			{Value: 0.242959},
		},
	}

	prettybad := 0.005
	now := time.Now()
	for idx := range sdTests.candles {
		t := now.Add(time.Duration(idx*itvl) * time.Second)
		sdTests.candles[idx].S = t
	}

	s, err := NewSeries(sdTests.candles, opts)
	if err != nil {
		t.Fatal(err)
	}
	close := NewOHLCProp(OHLCPropClose)
	stddev := NewStdDev(close, 10)
	if err := s.AddIndicator(name, stddev); err != nil {
		t.Fatal(err)
	}
	for idx, exp := range sdTests.expected {
		tim := now.Add(time.Duration(idx*itvl) * time.Second)
		v := s.GetValueForInterval(tim)
		if v == nil {
			t.Fatal(fmt.Errorf("interval should not be nil: %w", err))
		}
		if exp == nil && v.Indicators[name] == nil {
			continue // ok
		}
		if exp == nil && v.Indicators[name] != nil {
			t.Errorf("expected v to be nil but got %+v at idx: %d", v, idx)
		}
		if v.Indicators[name] == nil {
			t.Errorf("expected indicator to have value but got none at idx %d", idx)
		} else if math.Abs(exp.Value-*v.Indicators[name])/exp.Value > prettybad {
			t.Errorf("expected %+v but got %+v for idx: %d", exp.Value, *v.Indicators[name], idx)
		}
	}
}
