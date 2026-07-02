package id

import (
	"sync"
	"testing"
)

// atMillis builds an id at a fixed timestamp with a zero random tail —
// deterministic, for tests that assert exact ordering / round-trips.
func atMillis(unixMilli int64) string {
	return encode((uint64(unixMilli) & timeMask) << randBits)
}

func TestNew_MonotonicUniqueSorted(t *testing.T) {
	// Far more than fit in one millisecond of randomness, so the same-ms counter
	// path is exercised; strictly-increasing implies both unique and sorted.
	const n = 5000
	prev := ""
	for i := 0; i < n; i++ {
		s := New()
		if len(s) != Length || !Valid(s) {
			t.Fatalf("New() = %q is malformed", s)
		}
		if prev != "" && s <= prev {
			t.Fatalf("New() not strictly increasing: %q <= %q", s, prev)
		}
		prev = s
	}
}

// nextValue is where every awkward monotonic case lives; test them deterministically.
func TestNextValue(t *testing.T) {
	ms := func(n uint64) uint64 { return n << randBits } // a millisecond, already shifted
	const lowMax = randMask                              // full random tail
	cases := []struct {
		name           string
		last, nowShift uint64
		rnd            uint64
		want           uint64
	}{
		{"fresh: takes now|rnd", 0, ms(1000), 7, ms(1000) | 7},
		{"advancing ms: takes now|rnd", ms(1000) | 500, ms(1001), 3, ms(1001) | 3},
		{"same ms, higher rnd: jumps to now|rnd", ms(1000) | 500, ms(1000), 900, ms(1000) | 900},
		{"same ms, lower rnd: increments off last", ms(1000) | 500, ms(1000), 100, ms(1000) | 501},
		{"same ms, equal rnd: increments (no repeat)", ms(1000) | 500, ms(1000), 500, ms(1000) | 501},
		{"clock regression: increments off the higher last", ms(2000) | 10, ms(1000), 42, ms(2000) | 11},
		{"full-tick burst: borrows into the next ms slot", ms(1000) | lowMax, ms(1000), 0, ms(1001)},
	}
	for _, c := range cases {
		got := nextValue(c.last, c.nowShift, c.rnd)
		if got != c.want {
			t.Errorf("%s: nextValue(%d,%d,%d) = %d, want %d", c.name, c.last, c.nowShift, c.rnd, got, c.want)
		}
		if got <= c.last {
			t.Errorf("%s: nextValue not strictly greater than last (%d <= %d)", c.name, got, c.last)
		}
	}
}

func TestNewAt(t *testing.T) {
	// Stateless historical mint: valid, the timestamp round-trips, and it must NOT
	// disturb New's process monotonic counter.
	before := New()
	for _, ms := range []int64{0, 1_000, 1_700_000_000_000} {
		got := NewAt(ms)
		if !Valid(got) {
			t.Fatalf("NewAt(%d) = %q not Valid", ms, got)
		}
		if Time(got).UnixMilli() != ms {
			t.Errorf("Time(NewAt(%d)).UnixMilli() = %d", ms, Time(got).UnixMilli())
		}
	}
	if after := New(); after <= before {
		t.Errorf("NewAt disturbed the monotonic counter: New() went %q -> %q", before, after)
	}
}

func TestValid(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{New(), true},
		{"0123456789ab", true},   // exactly 12, all in alphabet
		{"", false},              // empty
		{"abc", false},           // too short
		{"0123456789abc", false}, // too long (13)
		{"0123456789ai", false},  // 'i' is dropped by Crockford
		{"0123456789ol", false},  // 'o'/'l' dropped too
		{"0123456789AB", false},  // uppercase not accepted
		{"0123456789a!", false},  // punctuation
	}
	for _, c := range cases {
		if got := Valid(c.in); got != c.want {
			t.Errorf("Valid(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestSortableByTime(t *testing.T) {
	// Ascending timestamps yield lexically ascending ids (time is the high bits).
	// Stay within the timeBits range so the mask never truncates.
	var prev string
	for ms := int64(1000); ms < int64(1)<<timeBits; ms *= 7 {
		s := atMillis(ms)
		if prev != "" && s <= prev {
			t.Fatalf("not sorted: atMillis(%d)=%q <= previous %q", ms, s, prev)
		}
		prev = s
	}
}

func TestTimeRoundTrip(t *testing.T) {
	for _, ms := range []int64{0, 1, 1_700_000_000_000, 1_800_000_000_000} {
		if got := Time(atMillis(ms)).UnixMilli(); got != ms {
			t.Errorf("Time(atMillis(%d)).UnixMilli() = %d", ms, got)
		}
	}
	if !Time("not-an-id").IsZero() {
		t.Error("Time of an invalid id should be the zero Time")
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	// Full 60-bit values, including nonzero random tails, must survive
	// encode->decode — the low bits that atMillis/Time never exercise.
	const max60 = uint64(1)<<60 - 1
	for _, v := range []uint64{0, 1, 0xffff, uint64(1) << 16, uint64(1) << 59, 0x123456789abc, max60} {
		if got := decode(encode(v)); got != v {
			t.Errorf("decode(encode(%#x)) = %#x", v, got)
		}
	}
}

func TestEncodeGoldenVectors(t *testing.T) {
	// Pin the exact alphabet + big-endian packing, so a reordering or endianness
	// regression is caught by an absolute value, not just relative ordering.
	cases := []struct {
		v    uint64
		want string
	}{
		{0, "000000000000"},
		{1, "000000000001"},
		{0x1f, "00000000000z"},
		{0xffff, "000000001zzz"},
	}
	for _, c := range cases {
		if got := encode(c.v); got != c.want {
			t.Errorf("encode(%#x) = %q, want %q", c.v, got, c.want)
		}
	}
}

func TestNew_ConcurrentUnique(t *testing.T) {
	// Best available substitute for -race (cgo unavailable here): many goroutines
	// mint at once; the mutex must yield no duplicates.
	const g, per = 100, 100
	var wg sync.WaitGroup
	out := make(chan string, g*per)
	for i := 0; i < g; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < per; j++ {
				out <- New()
			}
		}()
	}
	wg.Wait()
	close(out)
	seen := make(map[string]bool, g*per)
	for s := range out {
		if seen[s] {
			t.Fatalf("concurrent duplicate: %q", s)
		}
		seen[s] = true
	}
	if len(seen) != g*per {
		t.Fatalf("got %d unique ids, want %d", len(seen), g*per)
	}
}
