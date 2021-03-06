package statsd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/vendasta/gosdks/verrors"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/vendasta/gosdks/config"
)

var client statsdInterface
var clientNotInitialized = errors.New("StatsD client has not initialized")

const (
	nanosPerMilli = 1000000.0
)

// Initialize must be called before any tracing is done.
// Note that this package uses environment variables to determine the server to send metrics to.
func Initialize(statsNamespace string, tags []string) error {
	env := config.CurEnv().Name()
	tags = append(tags,
		fmt.Sprintf("env:%s", env),
		fmt.Sprintf("service:%s", statsNamespace),
		fmt.Sprintf("namespace:%s", config.GetGkeNamespace()),
	)
	tag := config.GetTag()
	if tag != "" {
		tags = append(tags, fmt.Sprintf("tag:%s", tag))
	}

	//use a fake client on local
	if config.IsLocal() {
		client = &fakeStatsD{}
		return nil
	}
	//use the datadog client on real environments
	c, err := createStatsdClient(statsNamespace, tags)
	if err != nil {
		return err
	}

	client = &dataDogStatsD{
		Client: c,
	}

	containerInitializedEvent(statsNamespace, tags)
	return nil
}

// InitializeCustomClient lets you create a more configurable statsd client
// Note that this package uses environment variables to determine the server to send metrics to.
func InitializeCustomClient(metricPrefix string, tags []string) (statsd.ClientInterface, error) {
	if config.IsLocal() {
		return &statsd.NoOpClient{}, nil
	}

	env := config.CurEnv().Name()
	tags = append(tags,
		fmt.Sprintf("env:%s", env),
		fmt.Sprintf("namespace:%s", config.GetGkeNamespace()),
	)

	c, err := createStatsdClient(metricPrefix, tags)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func createStatsdClient(namespace string, tags []string) (*statsd.Client, error) {
	ddAgentAddr := os.Getenv("DD_AGENT_ADDR")
	if ddAgentAddr == "" {
		nodeName := os.Getenv("GKE_NODENAME")
		if nodeName != "" {
			ddAgentAddr = fmt.Sprintf("%s:8125", nodeName)
		} else {
			ddAgentAddr = "dd-agent.default.svc.cluster.local:8125"
		}
	}

	//use the datadog client on real environments
	c, err := statsd.New(ddAgentAddr, statsd.WithNamespace(namespace), statsd.WithTags(tags))
	if err != nil {
		fmt.Printf("Error initializing statsd client. %s", err.Error())
		return nil, err
	}

	return c, nil
}

// EventPriority is the priority for the event
type EventPriority string

const (
	// Normal is the "normal" Priority for events
	Normal EventPriority = "normal"
	// Low is the "low" Priority for events
	Low EventPriority = "low"
)

// EventAlertType is the alert type of the event
type EventAlertType string

const (
	// Info is the "info" AlertType for events
	Info EventAlertType = "info"
	// Error is the "error" AlertType for events
	Error EventAlertType = "error"
	// Warning is the "warning" AlertType for events
	Warning EventAlertType = "warning"
	// Success is the "success" AlertType for events
	Success EventAlertType = "success"
)

// An Event is an object that can be posted to your event stream.
// Mirrored from https://github.com/DataDog/datadog-go/blob/master/statsd/statsd.go#L384
type Event struct {
	// Title of the event.  Required.
	Title string
	// Text is the description of the event.  Required.
	Text string
	// Timestamp is a timestamp for the event.  If not provided, the dogstatsd
	// server will set this to the current time.
	Timestamp time.Time
	// Hostname for the event.
	Hostname string
	// AggregationKey groups this event with others of the same key.
	AggregationKey string
	// Priority of the event.  Can be statsd.Low or statsd.Normal.
	Priority EventPriority
	// SourceTypeName is a source type for the event.
	SourceTypeName string
	// AlertType can be statsd.Info, statsd.Error, statsd.Warning, or statsd.Success.
	// If absent, the default value applied by the dogstatsd server is Info.
	AlertType EventAlertType
	// Tags for the event.
	Tags []string
}

// containerOnlineEvent sends an event signifying that a container initialized
func containerInitializedEvent(namespace string, tags []string) {
	event := &Event{
		Title:          fmt.Sprintf("%s Container Online", namespace),
		Text:           "A container initialized",
		AggregationKey: "container_initialized",
		Priority:       Low,
		AlertType:      Info,
		Tags:           tags,
	}
	LogEvent(event)
}

// Gauge measures the value of a metric at a particular time.
func Gauge(name string, value float64, tags []string, rate float64) error {
	if client == nil {
		return clientNotInitialized
	}
	return client.Gauge(name, value, tags, rate)
}

// Count tracks how many times something happened per second.
func Count(name string, value int64, tags []string, rate float64) error {
	if client == nil {
		return clientNotInitialized
	}
	return client.Count(name, value, tags, rate)
}

// Histogram tracks the statistical distribution of a set of values on each host.
func Histogram(name string, value float64, tags []string, rate float64) error {
	if client == nil {
		return clientNotInitialized
	}
	return client.Histogram(name, value, tags, rate)
}

// Distribution tracks the statistical distribution of a set of values across your infrastructure.
func Distribution(name string, value float64, tags []string, rate float64) error {
	if client == nil {
		return clientNotInitialized
	}
	return client.Distribution(name, value, tags, rate)
}

// Decr is just Count of 1
func Decr(name string, tags []string, rate float64) error {
	if client == nil {
		return clientNotInitialized
	}
	return client.Decr(name, tags, rate)
}

// Incr is just Count of 1
func Incr(name string, tags []string, rate float64) error {
	if client == nil {
		return clientNotInitialized
	}
	return client.Incr(name, tags, rate)
}

// Set counts the number of unique elements in a group.
func Set(name string, value string, tags []string, rate float64) error {
	if client == nil {
		return clientNotInitialized
	}
	return client.Set(name, value, tags, rate)
}

// Timing sends timing information, it is an alias for TimeInMilliseconds
func Timing(name string, value time.Duration, tags []string, rate float64) error {
	if client == nil {
		return clientNotInitialized
	}
	return client.Timing(name, value, tags, rate)
}

// TimeInMilliseconds sends timing information in milliseconds.
// It is flushed by statsd with percentiles, mean and other info (https://github.com/etsy/statsd/blob/master/docs/metric_types.md#timing)
func TimeInMilliseconds(name string, value float64, tags []string, rate float64) error {
	if client == nil {
		return clientNotInitialized
	}
	return client.TimeInMilliseconds(name, value, tags, rate)
}

// LogEvent logs the event
func LogEvent(event *Event) error {
	if client == nil {
		return clientNotInitialized
	}
	return client.Event(convertEvent(event))
}

// WithMonitoring runs a function wrapped with monitoring
// it ticks a Histogram metric with the time the function run took
// it adds a 'status' tag using `verrors.FromError(err).HTTPCode()` for the status code
//
// WithMonitoring returns the error value from running the function
//
// Example usage:
//
// err := statsd.WithMonitoring("generate-exec-report", []string{"report-period:weekly"}, 1, func() error {
//   err := doMyWork()
//   if err != nil {
//     return err
//   }
//   return nil
// })
func WithMonitoring(name string, tags []string, sampleRate float64, job func() error) error {
	started := time.Now()
	err := job()
	ended := time.Now()
	statusCode := 200
	if err != nil {
		statusCode = verrors.FromError(err).HTTPCode()
	}
	tags = append(tags, fmt.Sprintf("status:%d", statusCode))
	Timing(name, ended.Sub(started), tags, sampleRate)
	return err
}

// converts our event to the underlying datadog statsd event
func convertEvent(e *Event) *statsd.Event {
	eventType := fmt.Sprintf("aggregation_key:%s", e.AggregationKey)
	ev := &statsd.Event{
		Title:          e.Title,
		Text:           e.Text,
		Timestamp:      e.Timestamp,
		Hostname:       e.Hostname,
		AggregationKey: e.AggregationKey,
		SourceTypeName: e.SourceTypeName,
		Tags:           append(e.Tags, eventType),
	}
	setEventPriority(ev, e.Priority)
	setEventAlertType(ev, e.AlertType)
	return ev
}

// sets the statsd event's priority
func setEventPriority(ev *statsd.Event, ep EventPriority) {
	// Set type with enum values directly since dogstats type is private
	switch ep {
	case Low:
		ev.Priority = statsd.Low
	case Normal:
		ev.Priority = statsd.Normal
	}
}

// sets the statsd event's alert type
func setEventAlertType(ev *statsd.Event, eat EventAlertType) {
	// Set type with enum values directly since dogstats type is private
	switch eat {
	case Info:
		ev.AlertType = statsd.Info
	case Error:
		ev.AlertType = statsd.Error
	case Warning:
		ev.AlertType = statsd.Warning
	case Success:
		ev.AlertType = statsd.Success
	}
}

type statsdInterface interface {
	// Gauge measures the value of a metric at a particular time.
	Gauge(name string, value float64, tags []string, rate float64) error

	// Count tracks how many times something happened per second.
	Count(name string, value int64, tags []string, rate float64) error

	// Histogram tracks the statistical distribution of a set of values on each host.
	Histogram(name string, value float64, tags []string, rate float64) error

	// Distribution tracks the statistical distribution of a set of values across your infrastructure.
	Distribution(name string, value float64, tags []string, rate float64) error

	// Decr is just Count of 1
	Decr(name string, tags []string, rate float64) error

	// Incr is just Count of 1
	Incr(name string, tags []string, rate float64) error

	// Set counts the number of unique elements in a group.
	Set(name string, value string, tags []string, rate float64) error

	// Timing sends timing information, it is an alias for TimeInMilliseconds
	Timing(name string, value time.Duration, tags []string, rate float64) error

	// TimeInMilliseconds sends timing information in milliseconds.
	// It is flushed by statsd with percentiles, mean and other info (https://github.com/etsy/statsd/blob/master/docs/metric_types.md#timing)
	TimeInMilliseconds(name string, value float64, tags []string, rate float64) error

	// Event sends an event
	Event(event *statsd.Event) error
}

type dataDogStatsD struct {
	*statsd.Client
}
