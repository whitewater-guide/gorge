package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockScript struct {
	*LoggingScript
}

type mockOptions struct {
	Gauges int     `json:"gauges,omitempty"`
	Value  float64 `json:"value,omitempty"`
}

func (m *mockScript) ListGauges() (Gauges, error) {
	panic("implement me")
}

func (m *mockScript) Harvest(ctx context.Context, recv chan<- *Measurement, errs chan<- error, codes StringSet, since int64) {
	panic("implement me")
}

var allAtOnce = ScriptDescriptor{
	Name: "all_at_once",
	Mode: AllAtOnce,
	DefaultOptions: func() interface{} {
		return &mockOptions{
			Gauges: 10,
		}
	},
	Factory: func(name string, options interface{}) (script Script, err error) {
		if _, ok := options.(*mockOptions); ok {
			return &mockScript{}, nil
		}
		return nil, fmt.Errorf("failed to cast %T to %T", options, mockOptions{})
	},
}

var oneByOne = ScriptDescriptor{
	Name: "one_by_one",
	Mode: OneByOne,
	DefaultOptions: func() interface{} {
		return &mockOptions{
			Gauges: 10,
		}
	},
	Factory: func(name string, options interface{}) (script Script, err error) {
		return nil, errors.New("fail")
	},
}

func setup() *ScriptRegistry {
	res := NewRegistry()
	res.Register(&oneByOne)
	res.Register(&allAtOnce)
	return res
}

func TestScriptRegistry_List(t *testing.T) {
	registry := setup()
	res := registry.List()
	assert.Equal(t, "all_at_once", res[0].Name)
	assert.Equal(t, "one_by_one", res[1].Name)
}

func TestScriptRegistry_GetMode(t *testing.T) {
	registry := setup()
	mode, err := registry.GetMode("all_at_once")
	if assert.NoError(t, err) {
		assert.Equal(t, AllAtOnce, mode)
	}
	_, err = registry.GetMode("foo")
	assert.Error(t, err)
}

func TestScriptRegistry_Create(t *testing.T) {
	registry := setup()
	tests := []struct {
		name    string
		script  string
		options interface{}
		err     bool
	}{
		{
			name:    "success",
			script:  "all_at_once",
			options: &mockOptions{Gauges: 7, Value: 7},
			err:     false,
		},
		{
			name:    "bad options",
			script:  "all_at_once",
			options: "foo",
			err:     true,
		},
		{
			name:    "bad script",
			script:  "foo",
			options: "foo",
			err:     true,
		},
		{
			name:    "factory error",
			script:  "one_by_one",
			options: &mockOptions{Gauges: 7, Value: 7},
			err:     true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			res, mode, err := registry.Create(tt.script, tt.options)
			if tt.err {
				assert.Error(t, err)
			} else {
				if assert.NoError(t, err) {
					assert.Equal(t, AllAtOnce, mode)
					assert.IsType(t, &mockScript{}, res)
				}
			}
		})
	}
}

func TestScriptRegistry_ParseJsonOptions(t *testing.T) {
	registry := setup()
	tests := []struct {
		name     string
		script   string
		input    []json.RawMessage
		expected interface{}
		err      bool
	}{
		{
			name:     "success no input",
			script:   "all_at_once",
			input:    []json.RawMessage{},
			expected: &mockOptions{Gauges: 10, Value: 0},
		},
		{
			name:     "success single input",
			script:   "all_at_once",
			input:    []json.RawMessage{json.RawMessage(`{"value": 7}`)},
			expected: &mockOptions{Gauges: 10, Value: 7},
		},
		{
			name:     "success merge inputs",
			script:   "one_by_one",
			input:    []json.RawMessage{json.RawMessage(`{"value": 7}`), json.RawMessage(`{"gauges": 3}`)},
			expected: &mockOptions{Gauges: 3, Value: 7},
		},
		{
			name:   "error bad input",
			script: "one_by_one",
			input:  []json.RawMessage{json.RawMessage(`{"value": 7}`), json.RawMessage(`{"gauges}`)},
			err:    true,
		},
		{
			name:   "error bad script",
			script: "foo",
			input:  []json.RawMessage{json.RawMessage(`{"value": 7}`)},
			err:    true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			res, err := registry.ParseJSONOptions(tt.script, tt.input...)
			if tt.err {
				assert.Error(t, err)
			} else {
				if assert.NoError(t, err) {
					assert.Equal(t, tt.expected, res)
				}
			}
		})
	}
}

func TestScriptRegistry_CreateFromReader(t *testing.T) {
	registry := setup()
	tests := []struct {
		name   string
		script string
		reader io.Reader
		err    bool
	}{
		{
			name:   "success with json",
			script: "all_at_once",
			reader: strings.NewReader(`{"gauges": 7, "value": 7}`),
		},
		{
			name:   "success with empty reader",
			script: "all_at_once",
			reader: http.NoBody,
		},
		{
			name:   "bad options",
			script: "all_at_once",
			reader: strings.NewReader("{{"),
			err:    true,
		},
		{
			name:   "bad script",
			script: "foo",
			reader: strings.NewReader("foo"),
			err:    true,
		},
		{
			name:   "factory error",
			script: "one_by_one",
			reader: strings.NewReader(`{"gauges": 7, "value": 7}`),
			err:    true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			res, mode, err := registry.CreateFromReader(tt.script, tt.reader)
			if tt.err {
				assert.Error(t, err)
			} else {
				if assert.NoError(t, err) {
					assert.Equal(t, AllAtOnce, mode)
					assert.IsType(t, &mockScript{}, res)
				}
			}
		})
	}
}
