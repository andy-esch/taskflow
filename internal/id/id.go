// Package id mints and validates the stable identifier shared by tasks, epics,
// and audits (ADR-0003): a 12-character, lexically sortable string.
//
// An id is 60 bits, big-endian, so a plain byte comparison orders two ids by
// value — and therefore by creation time:
//
//	high 43 bits  Unix milliseconds (room through ~year 2248)
//	low  17 bits  randomness, or a monotonic counter within a millisecond
//
// rendered as 12 lowercase Crockford base32 characters (5 bits each). Crockford
// drops the look-alikes i, l, o, u, and lowercase keeps an id
// case-insensitive-filesystem-safe (it ends up in the filename).
//
// # Uniqueness scope (read before trusting this for concurrency)
//
// New is strictly monotonic *within one process*: a single process never repeats
// or reorders an id. It is NOT coordinated across processes — two concurrent
// writers (e.g. parallel cron agents) minting in the same millisecond collide
// with probability 1/2^17. Cross-process uniqueness is therefore the *writer's*
// job, not this package's: the store must reject-and-regenerate a duplicate id on
// create, or a single serve writer must serialize mints. A collision that slips
// through is a duplicate *id*, which id resolution surfaces as a recoverable
// ErrAmbiguous (the same class as a duplicate slug today), not corruption.
package id

import (
	"crypto/rand"
	"encoding/binary"
	"strings"
	"sync"
	"time"
)

// alphabet is Crockford base32 (lowercase), indexed by 5-bit value.
const alphabet = "0123456789abcdefghjkmnpqrstvwxyz"

// Length is the fixed character width of every id.
const Length = 12

const (
	timeBits = 43
	randBits = 17
	timeMask = uint64(1)<<timeBits - 1
	randMask = uint64(1)<<randBits - 1
)

// timeBits + randBits must exactly fill the id's bit budget (Length*5 == 60).
// These zero-sized declarations fail to compile if the split is ever edited to
// overflow or underflow the budget, rather than silently skewing the layout.
const (
	_ = uint(Length*5 - timeBits - randBits) // fails if the fields overflow 60 bits
	_ = uint(timeBits + randBits - Length*5) // fails if they underflow 60 bits
)

// mono keeps New strictly monotonic within a process: two calls in the same
// millisecond can't collide (the low bits become a counter) and ids stay
// lexically sorted. Process-local generation state, not an injected dependency,
// so a package-level guard is its natural home — like a CSPRNG source.
var mono struct {
	sync.Mutex
	last uint64
}

// New returns a fresh id stamped with the current time, strictly greater than any
// id this process has already returned (see the package "Uniqueness scope" note
// for the cross-process caveat). It panics only if the system CSPRNG is
// unavailable — unrecoverable for an identity mint, the stance crypto/rand's own
// helpers take.
func New() string {
	nowShifted := (uint64(time.Now().UnixMilli()) & timeMask) << randBits
	rnd := randomTail() // outside the lock: crypto/rand is the slow part and needs no serialization
	mono.Lock()
	v := nextValue(mono.last, nowShifted, rnd)
	mono.last = v
	mono.Unlock()
	return encode(v)
}

// NewAt returns an id stamped with a specific time (Unix milliseconds) and a
// random tail. Unlike New it is STATELESS — it never touches the process monotonic
// counter — so it is for minting ids for *known* entities out of natural time
// order: the migration backfilling existing files from their created: date, and
// fixtures. Callers minting several ids at the same millisecond (e.g. same-day
// historical files) must dedupe: a shared-ms collision has probability 1/2^randBits
// per pair, so regenerate on a clash.
func NewAt(unixMilli int64) string {
	v := (uint64(unixMilli) & timeMask) << randBits
	return encode(v | randomTail())
}

// nextValue is the pure, total monotonic step: given the last value this process
// emitted, the current (already time-masked and shifted) millisecond, and a random
// tail, it returns the next value — always strictly greater than last. It absorbs
// every awkward case in one exhaustively-testable place: a fresh (or higher-random)
// millisecond takes now|rnd directly; the same millisecond with a lower random, a
// backwards clock, or a >2^randBits same-ms burst that has already borrowed into the
// time bits all fall through to last+1, which stays strictly increasing (a full-tick
// burst simply rolls the value into the next ms slot).
func nextValue(last, nowShifted, rnd uint64) uint64 {
	if v := nowShifted | rnd; v > last {
		return v
	}
	return last + 1
}

// randomTail returns randBits of cryptographic randomness. It panics only if the
// system CSPRNG is unavailable — unrecoverable for an identity mint.
func randomTail() uint64 {
	var r [4]byte
	if _, err := rand.Read(r[:]); err != nil {
		panic("id: crypto/rand unavailable: " + err.Error())
	}
	return uint64(binary.BigEndian.Uint32(r[:])) & randMask
}

// encode renders the low 60 bits of v as Length Crockford chars, most-significant
// group first, so lexical order matches numeric (hence chronological) order.
func encode(v uint64) string {
	var b [Length]byte
	for i := range b {
		b[i] = alphabet[(v>>(5*(Length-1-i)))&0x1f]
	}
	return string(b[:])
}

// decode is the inverse of encode: it reads a valid id's Length Crockford chars
// back into its 60-bit value. Callers must pass a Valid id.
func decode(s string) uint64 {
	var v uint64
	for i := 0; i < len(s); i++ {
		v = v<<5 | uint64(strings.IndexByte(alphabet, s[i]))
	}
	return v
}

// Valid reports whether s is exactly Length characters, all in the alphabet. It
// is strict (not Crockford's lenient decode): ids this package mints are always
// lowercase and unambiguous, so anything else is treated as malformed.
func Valid(s string) bool {
	if len(s) != Length {
		return false
	}
	for i := 0; i < len(s); i++ {
		if strings.IndexByte(alphabet, s[i]) < 0 {
			return false
		}
	}
	return true
}

// Time recovers the creation timestamp (millisecond precision) encoded in a valid
// id, or the zero Time if s is malformed. A useful fallback sort key and forensic
// hint — though under an extreme same-ms burst the counter can borrow into the
// time bits, so treat the timestamp as a lower bound, not gospel.
func Time(s string) time.Time {
	if !Valid(s) {
		return time.Time{}
	}
	return time.UnixMilli(int64(decode(s) >> randBits))
}
