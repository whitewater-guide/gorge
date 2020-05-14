package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"sort"

	"github.com/sirupsen/logrus"
)

// ErrScriptNotFound means that script is not registered in registry
var ErrScriptNotFound = errors.New("script not found")

// Script represents bunch of methods to harvest measurements and gauges from certain upstream source
type Script interface {
	ListGauges() (Gauges, error)
	// Harvests measurements from upstream and writes them to recv channel, then closes both channels.
	// If unrecoverable error happens during this process, writes it into errs channel and closes both channels.
	// codes are set of gauge codes to harvest from upstream. It's meant to be passed to upstream. The script itself should not make use of it.
	// since is the last timestamp (usually for one-by-one scripts) that is meannt to be passed to upstream. The script itself should not make use of it.
	Harvest(ctx context.Context, recv chan<- *Measurement, errs chan<- error, codes StringSet, since int64)
	SetLogger(logger *logrus.Entry)
	GetLogger() *logrus.Entry
}

// ScriptFactory creates an instance of script and provides is with options
// It must faile if generic options cannot be cast to script's internal options
type ScriptFactory func(name string, options interface{}) (Script, error)

// ScriptDescriptor represents a script registered in gorge
type ScriptDescriptor struct {
	Name string `json:"name"`
	// Description is human-readable name of data source, something that you can google
	Description    string             `json:"description"`
	Mode           HarvestMode        `json:"mode"`
	DefaultOptions func() interface{} `json:"-"`
	Factory        ScriptFactory      `json:"-"`
}

// ScriptRegistry is where all the script we can use must be registered
type ScriptRegistry struct {
	descriptors map[string]ScriptDescriptor
}

// NewRegistry creates new ScriptRegistry
func NewRegistry() *ScriptRegistry {
	return &ScriptRegistry{descriptors: map[string]ScriptDescriptor{}}
}

// Register new script in registry
func (r *ScriptRegistry) Register(d *ScriptDescriptor) {
	r.descriptors[d.Name] = *d
}

// Create new instance of a script with given options.
// options must be instance of a script's internal options
func (r *ScriptRegistry) Create(name string, options interface{}) (Script, HarvestMode, error) {
	d, exists := r.descriptors[name]
	if !exists {
		return nil, 0, ErrScriptNotFound
	}
	script, err := d.Factory(name, options)
	if err != nil {
		return nil, 0, WrapErr(err, "failed to create script instance").With("script", name)
	}
	return script, d.Mode, nil
}

// CreateFromReader create new instance of a script. Options are provided in a form of reader.
// This reader must provide JSON options that can be unmarshalled into instance of script's internal options
func (r *ScriptRegistry) CreateFromReader(name string, optsReader io.Reader) (Script, HarvestMode, error) {
	d, exists := r.descriptors[name]
	if !exists {
		return nil, 0, ErrScriptNotFound
	}
	options := d.DefaultOptions()
	decoder := json.NewDecoder(optsReader)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(options)
	if err != nil && err != io.EOF {
		return nil, 0, WrapErr(err, "failed create options from reader").With("script", name)
	}
	return r.Create(name, options)
}

// ParseJSONOptions parses raw json messages with script's options and merges them into on instance of script's internal options
func (r *ScriptRegistry) ParseJSONOptions(name string, inputs ...json.RawMessage) (interface{}, error) {
	d, exists := r.descriptors[name]
	if !exists {
		return nil, ErrScriptNotFound
	}
	options := d.DefaultOptions()
	for _, raw := range inputs {
		if len(raw) > 0 {
			decoder := json.NewDecoder(bytes.NewReader(raw))
			decoder.DisallowUnknownFields()
			err := decoder.Decode(options)
			if err != nil {
				return nil, WrapErr(err, "failed to unmarshal options")
			}
		}
	}
	return options, nil
}

// GetMode returns harvest mode of a registered script
func (r *ScriptRegistry) GetMode(name string) (HarvestMode, error) {
	d, exists := r.descriptors[name]
	if !exists {
		return 0, ErrScriptNotFound
	}
	return d.Mode, nil
}

// List lists all registered scripts
func (r *ScriptRegistry) List() []ScriptDescriptor {
	result, i := make([]ScriptDescriptor, len(r.descriptors)), 0
	for _, d := range r.descriptors {
		result[i] = d
		i++
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// LoggingScript is used to inject loggers into scripts
// It must be embedded into concrete scripts
type LoggingScript struct {
	logger *logrus.Entry
}

// SetLogger injects logger into script
func (s *LoggingScript) SetLogger(logger *logrus.Entry) {
	s.logger = logger
}

// GetLogger returns injected logger, or discarding logger if none has been injected
func (s *LoggingScript) GetLogger() *logrus.Entry {
	if s.logger == nil {
		logger := logrus.New()
		logger.SetOutput(ioutil.Discard)
		s.logger = logrus.NewEntry(logger)
	}
	return s.logger
}
