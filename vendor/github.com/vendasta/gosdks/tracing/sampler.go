package tracing

import (
	"math"
	"sync"
	"time"

	"go.opencensus.io/trace"
)

// Tracing implementations were taken: from https://github.com/jaegertracing/jaeger-client-go/blob/master/sampler.go
// This is due to open census having tracing implementations that are not suitable for our traffic.

// TODO: Investigate validity of above statement, what about our traffic makes them unsuitable?

// NewRateLimitingSampler creates a sampler that samples at most maxTracesPerMinute. The distribution of sampled
// traces follows burstiness of the service, i.e. a service with uniformly distributed requests will have those
// requests sampled uniformly as well, but if requests are bursty, especially sub-second, then a number of
// sequential requests can be sampled each second.
func NewRateLimitingSampler(maxTracesPerMinute float64) *RateLimitingSampler {
	return &RateLimitingSampler{
		rateLimiter: NewRateLimiter(maxTracesPerMinute, math.Max(maxTracesPerMinute, 1)),
	}
}

type RateLimitingSampler struct {
	rateLimiter RateLimiter
}

func (r *RateLimitingSampler) Sampler(trace.SamplingParameters) trace.SamplingDecision {
	return trace.SamplingDecision{Sample: r.rateLimiter.CheckCredit(1.0)}
}

// GuaranteedThroughputProbabilisticSampler is a sampler that leverages both probabilisticSampler and
// rateLimitingSampler. The lowerBoundSampler is used as a guaranteed lower bound sampler such that
// every operation is sampled at least once in a time interval defined by the lowerBound. ie a lowerBound
// of 1.0 / (60 * 10) will sample an operation at least once every 10 minutes.
//
// It also ensures that we only trace at most the given higher bound threshold. If we've sampled at the
// higher bound, we will by pass the lower bound and probabilistic sampler.
//
// The probabilisticSampler is given higher priority when tags are emitted, ie. if IsSampled() for both
// samplers return true, the tags for probabilisticSampler will be used.
type GuaranteedThroughputProbabilisticSampler struct {
	probabilisticSampler trace.Sampler
	lowerBoundSampler    trace.Sampler
	higherBoundSampler   trace.Sampler
	samplingRate         float64
}

// NewGuaranteedThroughputProbabilisticSampler returns a delegating sampler that applies both
// probabilisticSampler and rateLimitingSampler.
func NewGuaranteedThroughputProbabilisticSampler(lowerBound, higherBound, samplingRate float64) *GuaranteedThroughputProbabilisticSampler {
	return newGuaranteedThroughputProbabilisticSampler(lowerBound, higherBound, samplingRate)
}

func newGuaranteedThroughputProbabilisticSampler(lowerBound, higherBound, samplingRate float64) *GuaranteedThroughputProbabilisticSampler {
	s := &GuaranteedThroughputProbabilisticSampler{
		lowerBoundSampler:    NewRateLimitingSampler(lowerBound).Sampler,
		higherBoundSampler:   NewRateLimitingSampler(higherBound).Sampler,
		probabilisticSampler: trace.ProbabilitySampler(samplingRate),
		samplingRate:         samplingRate,
	}
	return s
}

func (s *GuaranteedThroughputProbabilisticSampler) Sampler(tr trace.SamplingParameters) (result trace.SamplingDecision) {
	defer func() {
		// The lower bound / probabilistic sampler says yes, we need to confirm we're not above the higher bound
		if result.Sample {
			// If we've consumed the higher bound, then bypass the probabilistic and min
			sampleAtHigherBound := s.higherBoundSampler(tr)
			if !sampleAtHigherBound.Sample {
				result = sampleAtHigherBound
			}
		}
	}()
	if sampled := s.probabilisticSampler(tr); sampled.Sample {
		s.lowerBoundSampler(tr)
		return trace.SamplingDecision{Sample: true}
	}
	return s.lowerBoundSampler(tr)
}

// RateLimiter is a filter used to check if a message that is worth itemCost units is within the rate limits.
type RateLimiter interface {
	CheckCredit(itemCost float64) bool
}

type rateLimiter struct {
	sync.Mutex

	creditsPerMinute float64
	balance          float64
	maxBalance       float64
	lastTick         time.Time

	timeNow func() time.Time
}

// NewRateLimiter creates a new rate limiter based on leaky bucket algorithm, formulated in terms of a
// credits balance that is replenished every time CheckCredit() method is called (tick) by the amount proportional
// to the time elapsed since the last tick, up to max of creditsPerMinute. A call to CheckCredit() takes a cost
// of an item we want to pay with the balance. If the balance exceeds the cost of the item, the item is "purchased"
// and the balance reduced, indicated by returned value of true. Otherwise the balance is unchanged and return false.
//
// This can be used to limit a rate of messages emitted by a service by instantiating the Rate Limiter with the
// max number of messages a service is allowed to emit per second, and calling CheckCredit(1.0) for each message
// to determine if the message is within the rate limit.
//
// It can also be used to limit the rate of traffic in bytes, by setting creditsPerMinute to desired throughput
// as bytes/minute, and calling CheckCredit() with the actual message size.
func NewRateLimiter(creditsPerMinute, maxBalance float64) RateLimiter {
	return &rateLimiter{
		creditsPerMinute: creditsPerMinute,
		balance:          maxBalance,
		maxBalance:       maxBalance,
		lastTick:         time.Now(),
		timeNow:          time.Now}
}

func (b *rateLimiter) CheckCredit(itemCost float64) bool {
	b.Lock()
	defer b.Unlock()
	// calculate how much time passed since the last tick, and update current tick
	currentTime := b.timeNow()
	elapsedTime := currentTime.Sub(b.lastTick)
	b.lastTick = currentTime
	// calculate how much credit have we accumulated since the last tick
	b.balance += elapsedTime.Minutes() * b.creditsPerMinute
	if b.balance > b.maxBalance {
		b.balance = b.maxBalance
	}
	// if we have enough credits to pay for current item, then reduce balance and allow
	if b.balance >= itemCost {
		b.balance -= itemCost
		return true
	}
	return false
}
