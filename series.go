package pine

import (
	"errors"
	"fmt"
	"math"
	"time"
)

type Series interface {
	AddIndicator(name string, i Indicator) error
	AddExec(v TPQ) error
	AddOHLCV(v OHLCV) error
	GetValueForInterval(t time.Time) *Interval
}

type Indicator interface {
	ApplyOpts(opts SeriesOpts) error
	GetValueForInterval(t time.Time) *Interval
	Update(v OHLCV) error
}

type Interval struct {
	StartTime  time.Time
	OHLCV      *OHLCV
	Value      float64
	Indicators map[string]*float64
}

type TPQ struct {
	Timestamp time.Time
	Px        float64
	Qty       float64
}

type OHLCV struct {
	O, H, L, C, V float64
	S             time.Time
}

// SeriesOpts is options required for creating Series
type SeriesOpts struct {
	// interval in seconds
	Interval int
	// max number of OHLC bars to keep
	Max int
	// instruction when there are no execs during interval
	EmptyInst EmptyInst
}

// EmptyInst is instruction when no values are set for the interval
type EmptyInst int

const (
	// EmptyInstUseLastClose uses the last close value for open, high, low, close but zero for volume
	EmptyInstUseLastClose EmptyInst = iota
	// EmptyInstIgnore ignores intervals if empty
	EmptyInstIgnore
	// EmptyInstUseZeros uses zeros for open, high, low, close, and volume
	EmptyInstUseZeros
)

// NewSeries generates new OHLCV serie
func NewSeries(ohlcv []OHLCV, opts SeriesOpts) (Series, error) {
	// Validate validates series opts and returns error if not good
	var err error
	if opts.Interval <= 0 {
		err = errors.New("`Interval` must be positive")
	} else if opts.Max <= 0 {
		err = errors.New("`Max` must be positive")
	}
	if err != nil {
		return nil, fmt.Errorf("error validating seriesopts: %w", err)
	}
	tm := make(map[time.Time]*OHLCV)
	s := &series{
		items:   make(map[string]Indicator),
		opts:    opts,
		timemap: tm,
		values:  make([]OHLCV, 0, opts.Max),
	}
	s.initValues(ohlcv)
	return s, nil
}

type series struct {
	items    map[string]Indicator
	lastExec TPQ
	lastOHLC *OHLCV
	opts     SeriesOpts
	values   []OHLCV
	timemap  map[time.Time]*OHLCV
}

func (s *series) initValues(values []OHLCV) {
	for _, v := range values {
		s.insertInterval(v)
	}
}

func (s *series) insertInterval(v OHLCV) {
	t := s.getLastIntervalFromTime(v.S)
	v.S = t
	_, ok := s.timemap[t]
	if !ok {
		s.values = append(s.values, v)
		s.timemap[t] = &v
		s.lastOHLC = &v
	}
}

func (s *series) updateIndicators(v OHLCV) error {
	for _, ind := range s.items {
		if err := ind.Update(v); err != nil {
			return fmt.Errorf("error updating indicator: %w", err)
		}
	}
	return nil
}

func (s *series) getLastIntervalFromTime(t time.Time) time.Time {
	year, month, day := t.UTC().Date()
	st := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	m := s.getMultiplierDiff(t, st)
	return st.Add(time.Duration(m*s.opts.Interval) * time.Second)
}

func (s *series) getMultiplierDiff(t time.Time, st time.Time) int {
	diff := t.Sub(st).Seconds()
	return int(diff / float64(s.opts.Interval))
}

func (s *series) getOHLCV(t time.Time) *OHLCV {
	return s.timemap[t]
}

// series_add_exec.go

func (s *series) AddExec(v TPQ) error {
	start := s.getLastIntervalFromTime(v.Timestamp)
	if s.lastOHLC == nil {
		if err := s.createNewOHLCV(v, start); err != nil {
			return fmt.Errorf("error creating new ohlcv: %w", err)
		}
	} else if s.lastOHLC.S.Equal(start) {
		if err := s.updateLastOHLCV(v); err != nil {
			return fmt.Errorf("error creating new ohlcv: %w", err)
		}
	} else if start.Sub(s.lastOHLC.S).Seconds() > 0 {
		// calculate how many intervals are missing
		if err := s.updateAndFillGaps(v, start); err != nil {
			return fmt.Errorf("error updating and filling gaps: %w", err)
		}
	}
	s.lastExec = v
	return nil
}

func NewOHLCVWithSamePx(px, qty float64, t time.Time) OHLCV {
	return OHLCV{px, px, px, px, qty, t}
}

func (s *series) createNewOHLCV(v TPQ, start time.Time) error {
	// create first one
	ohlcv := NewOHLCVWithSamePx(v.Px, v.Qty, start)
	s.insertInterval(ohlcv)
	if err := s.updateIndicators(ohlcv); err != nil {
		return fmt.Errorf("error updating indicator: %w", err)
	}
	return nil
}

func (s *series) updateLastOHLCV(v TPQ) error {
	itvl := s.lastOHLC
	itvl.C = v.Px
	itvl.V += v.Qty
	if v.Px > itvl.H {
		itvl.H = v.Px
	} else if v.Px < itvl.L {
		itvl.L = v.Px
	}
	if err := s.updateIndicators(*itvl); err != nil {
		return fmt.Errorf("error updating indicator: %w", err)
	}
	return nil
}

func (s *series) updateAndFillGaps(v TPQ, start time.Time) error {
	switch s.opts.EmptyInst {
	case EmptyInstUseLastClose:
		// figure out how many
		m := s.getMultiplierDiff(v.Timestamp, s.lastOHLC.S)
		usenew := int(math.Max(float64(m-1), float64(0)))
		for i := 0; i < m; i++ {
			var px, qty float64
			if i == usenew {
				px = v.Px
				qty = v.Qty
			} else {
				px = s.lastOHLC.C
				qty = 0
			}
			newt := s.lastOHLC.S.Add(time.Duration(s.opts.Interval) * time.Second)
			ohlcv := NewOHLCVWithSamePx(px, qty, newt)
			s.insertInterval(ohlcv)
			s.updateIndicators(ohlcv)
		}
	case EmptyInstUseZeros:
		// figure out how many
		m := s.getMultiplierDiff(v.Timestamp, s.lastOHLC.S)
		usenew := int(math.Max(float64(m-1), float64(0)))
		for i := 0; i < m; i++ {
			var px, qty float64
			if i == usenew {
				px = v.Px
				qty = v.Qty
			} else {
				px = 0
				qty = 0
			}
			newt := s.lastOHLC.S.Add(time.Duration(s.opts.Interval) * time.Second)
			ohlcv := NewOHLCVWithSamePx(px, qty, newt)
			s.insertInterval(ohlcv)
			s.updateIndicators(ohlcv)
		}
	case EmptyInstIgnore:
		v := NewOHLCVWithSamePx(v.Px, v.Qty, start)
		s.insertInterval(v)
		s.updateIndicators(v)
	default:
		return fmt.Errorf("unsupported interval: %+v", s.opts.EmptyInst)
	}
	return nil
}

// series_add_indicator.go

func (s *series) AddIndicator(name string, i Indicator) error {
	// enforce series constraint
	if err := i.ApplyOpts(s.opts); err != nil {
		return fmt.Errorf("error applying opts")
	}
	// update with current values downstream
	for _, v := range s.values {
		if err := i.Update(v); err != nil {
			return fmt.Errorf("error updating indicator")
		}
	}
	s.items[name] = i
	return nil
}

// series_add_ohlcv.go

func (s *series) AddOHLCV(v OHLCV) error {
	start := s.getLastIntervalFromTime(v.S)
	v.S = start
	if s.lastOHLC == nil {
		// create first one
		s.insertInterval(v)
	} else if s.lastOHLC.S.Equal(start) {
		// update this interval
		itvl := s.lastOHLC
		itvl.O = v.O
		itvl.H = v.H
		itvl.L = v.L
		itvl.C = v.C
		itvl.V = v.V
	} else if start.Sub(s.lastOHLC.S).Seconds() > 0 {
		// calculate how many intervals are missing
		switch s.opts.EmptyInst {
		case EmptyInstUseLastClose:
			// figure out how many
			m := s.getMultiplierDiff(v.S, s.lastOHLC.S)
			usenew := int(math.Max(float64(m-1), float64(0)))
			for i := 0; i < m; i++ {
				var px, qty float64
				var ohlcv OHLCV
				if i == usenew {
					ohlcv = v
				} else {
					px = s.lastOHLC.C
					qty = 0
					newt := s.lastOHLC.S.Add(time.Duration(s.opts.Interval) * time.Second)
					ohlcv = NewOHLCVWithSamePx(px, qty, newt)
				}
				s.insertInterval(ohlcv)
			}
		case EmptyInstUseZeros:
			// figure out how many
			m := s.getMultiplierDiff(v.S, s.lastOHLC.S)
			usenew := int(math.Max(float64(m-1), float64(0)))
			for i := 0; i < m; i++ {
				var px, qty float64
				var ohlcv OHLCV
				if i == usenew {
					ohlcv = v
				} else {
					px = 0
					qty = 0
					newt := s.lastOHLC.S.Add(time.Duration(s.opts.Interval) * time.Second)
					ohlcv = NewOHLCVWithSamePx(px, qty, newt)
				}
				s.insertInterval(ohlcv)
			}
		case EmptyInstIgnore:
			s.insertInterval(v)
		default:
			return fmt.Errorf("unsupported interval: %+v", s.opts.EmptyInst)
		}
	}
	s.lastOHLC = &v
	return nil
}

// series_get_value_for_interval.go

func (s *series) GetValueForInterval(t time.Time) *Interval {
	if s.lastOHLC == nil {
		return nil
	}
	if !s.lastOHLC.S.Equal(t) {
		// if time is within interval, adjust it
		t = s.getLastIntervalFromTime(t)
	}
	t = t.UTC()
	inds := make(map[string]*float64)
	for k, v := range s.items {
		val := v.GetValueForInterval(t)
		if val != nil {
			inds[k] = &val.Value
		}
	}
	v, ok := s.timemap[t]
	if !ok {
		return nil
	}
	return &Interval{
		StartTime:  v.S,
		OHLCV:      v,
		Indicators: inds,
	}
}
