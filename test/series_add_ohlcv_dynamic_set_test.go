package pine_test

import (
	"fmt"
	"testing"
	"time"

	. "github.com/xpt-nl/pine"
)

func TestSeriesAddOHLCV(t *testing.T) {
	opts := SeriesOpts{
		Interval: 300,
		Max:      100,
	}
	_, err := NewSeries(nil, opts)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	fivemin := now.Add(5 * time.Minute)
	data := []OHLCV{
		{
			O: 14,
			H: 15,
			L: 13,
			C: 14,
			V: 131,
			S: now,
		},
	}
	s, err := NewSeries(data, opts)
	if err != nil {
		t.Fatal(err)
	}

	// adding same interval should replace OHLCV
	rep := OHLCV{
		O: 14,
		H: 19,
		L: 10,
		C: 19,
		V: 2311,
		S: now,
	}
	// adding new interval should create new OHLCV
	newv := OHLCV{
		O: 18,
		H: 20,
		L: 17,
		C: 19,
		V: 723,
		S: fivemin,
	}

	if err := s.AddOHLCV(rep); err != nil {
		t.Fatal(fmt.Errorf("error adding ohlcv: %+v: %w", rep, err))
	}
	v := s.GetValueForInterval(now)
	if v == nil {
		t.Fatal("expected v to be non nil but got nil")
	}
	h := v.OHLCV
	if h.O != 14 {
		t.Fatalf("expected new open to be 14 but got %+v", h.O)
	} else if h.H != 19 {
		t.Fatalf("expected new high to be 19 but got %+v", h.H)
	} else if h.L != 10 {
		t.Fatalf("expected new high to be 10 but got %+v", h.L)
	} else if h.V != 2311 {
		t.Fatalf("expected vol to be 2311 but got %+v", h.V)
	} else if h.C != 19 {
		t.Fatalf("expected close to be 20 but got %+v", h.C)
	}

	if err := s.AddOHLCV(newv); err != nil {
		t.Fatal(fmt.Errorf("error adding ohlcv: %+v: %w", rep, err))
	}
	v = s.GetValueForInterval(fivemin)
	if v == nil {
		t.Fatal("expected v to be non nil but got nil")
	}
	h = v.OHLCV
	if h.O != 18 {
		t.Fatalf("expected new open to be 18 but got %+v", h.O)
	} else if h.H != 20 {
		t.Fatalf("expected new high to be 20 but got %+v", h.H)
	} else if h.L != 17 {
		t.Fatalf("expected new high to be 17 but got %+v", h.L)
	} else if h.V != 723 {
		t.Fatalf("expected vol to be 723 but got %+v", h.V)
	} else if h.C != 19 {
		t.Fatalf("expected close to be 19 but got %+v", h.C)
	}

	// // This should update low, close, and volume
	// tpqlow := TPQ{
	// 	Timestamp: fivemin,
	// 	Px:        3,
	// 	Qty:       4,
	// }
	// if err := s.AddExec(tpqlow); err != nil {
	// 	t.Fatal(errors.Wrapf(err, "error adding exec: %+v", tpqlow))
	// }
	// v = s.GetValueForInterval(fivemin)
	// if v == nil {
	// 	t.Fatal("expected v to be non nil but got nil")
	// }
	// l := v.OHLCV
	// if l.O != 13 {
	// 	t.Fatalf("expected new open to be 13 but got %+v", h.O)
	// } else if l.H != 20 {
	// 	t.Fatalf("expected new high to be 20 but got %+v", h.H)
	// } else if l.V != 1+12+4 {
	// 	t.Fatalf("expected vol to be 13 but got %+v", h.V)
	// } else if l.C != 3 {
	// 	t.Fatalf("expected close to be 3 but got %+v", h.C)
	// } else if l.L != 3 {
	// 	t.Fatalf("expected close to be 3 but got %+v", h.L)
	// }

	// // This should create new interval
	// tenmin := fivemin.Add(5 * time.Minute)
	// tpqnew := TPQ{
	// 	Timestamp: tenmin,
	// 	Px:        10,
	// 	Qty:       9,
	// }
	// if err := s.AddExec(tpqnew); err != nil {
	// 	t.Fatal(errors.Wrapf(err, "error adding exec: %+v", tpqnew))
	// }
	// v = s.GetValueForInterval(tenmin)
	// if v == nil {
	// 	t.Fatal("expected v to be non nil but got nil")
	// }
	// n := v.OHLCV
	// if n.S.Sub(l.S).Seconds() != 300 {
	// 	t.Fatalf("expected starting interval to have 300 second diff but got %+v", n.S.Sub(l.S).Seconds())
	// } else if n.O != 10 {
	// 	t.Fatalf("expected new open to be 10 but got %+v", n.O)
	// } else if n.H != 10 {
	// 	t.Fatalf("expected new high to be 10 but got %+v", n.H)
	// } else if n.V != 9 {
	// 	t.Fatalf("expected vol to be 9 but got %+v", n.V)
	// } else if n.C != 10 {
	// 	t.Fatalf("expected close to be 10 but got %+v", n.C)
	// } else if n.L != 10 {
	// 	t.Fatalf("expected close to be 10 but got %+v", n.L)
	// }

	// // This should create 2 intervals since this spans two intervals
	// // refer to ExecInst
	// twemin := tenmin.Add(10 * time.Minute)
	// tpqtwe := TPQ{
	// 	Timestamp: twemin,
	// 	Px:        14,
	// 	Qty:       3,
	// }
	// if err := s.AddExec(tpqtwe); err != nil {
	// 	t.Fatal(errors.Wrapf(err, "error adding exec: %+v", tpqtwe))
	// }
	// v = s.GetValueForInterval(twemin.Add(-5 * time.Minute))
	// if v == nil {
	// 	t.Fatal("expected v to be non nil but got nil")
	// }
	// n = v.OHLCV
	// if n.S.Sub(l.S).Seconds() != 600 {
	// 	t.Fatalf("expected starting interval to have 600 second diff but got %+v", n.S.Sub(l.S).Seconds())
	// } else if n.O != 10 {
	// 	t.Fatalf("expected new open to be 10 but got %+v", n.O)
	// } else if n.H != 10 {
	// 	t.Fatalf("expected new high to be 10 but got %+v", n.H)
	// } else if n.V != 0 {
	// 	t.Fatalf("expected vol to be 0 but got %+v", n.V)
	// } else if n.C != 10 {
	// 	t.Fatalf("expected close to be 10 but got %+v", n.C)
	// } else if n.L != 10 {
	// 	t.Fatalf("expected close to be 10 but got %+v", n.L)
	// }

	// v = s.GetValueForInterval(twemin)
	// if v == nil {
	// 	t.Fatal("expected v to be non nil but got nil")
	// }
	// n = v.OHLCV
	// if n.S.Sub(l.S).Seconds() != 900 {
	// 	t.Fatalf("expected starting interval to have 900 second diff but got %+v", n.S.Sub(l.S).Seconds())
	// } else if n.O != 14 {
	// 	t.Fatalf("expected new open to be 14 but got %+v", n.O)
	// } else if n.H != 14 {
	// 	t.Fatalf("expected new high to be 14 but got %+v", n.H)
	// } else if n.V != 3 {
	// 	t.Fatalf("expected vol to be 0 but got %+v", n.V)
	// } else if n.C != 14 {
	// 	t.Fatalf("expected close to be 14 but got %+v", n.C)
	// } else if n.L != 14 {
	// 	t.Fatalf("expected close to be 14 but got %+v", n.L)
	// }
}
