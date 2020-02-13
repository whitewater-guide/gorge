package core

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenFromSlice(t *testing.T) {
	ms := []Measurement{
		GenerateRandMeasurement("all_at_once", "g000", 100, 0, 0),
		GenerateRandMeasurement("all_at_once", "g001", 200, 0, 0),
		GenerateRandMeasurement("all_at_once", "g002", 300, 0, 0),
	}
	out := GenFromSlice(context.Background(), ms)
	m, ok := <-out
	assert.Equal(t, 100.0, m.Flow.Float64Value())
	assert.Equal(t, true, ok)
	m = <-out
	assert.Equal(t, 200.0, m.Flow.Float64Value())
	m = <-out
	assert.Equal(t, 300.0, m.Flow.Float64Value())
	_, ok = <-out
	assert.Equal(t, false, ok)
}

func TestGenFromSliceCanceled(t *testing.T) {
	ms := []Measurement{
		GenerateRandMeasurement("all_at_once", "g000", 100, 0, 0),
		GenerateRandMeasurement("all_at_once", "g001", 200, 0, 0),
		GenerateRandMeasurement("all_at_once", "g002", 300, 0, 0),
	}
	ctx, cancel := context.WithCancel(context.Background())
	out := GenFromSlice(ctx, ms)
	m := <-out
	cancel()
	time.Sleep(time.Millisecond) // otherwise select is flaky
	assert.Equal(t, 100.0, m.Flow.Float64Value())
	_, ok := <-out
	assert.Equal(t, false, ok)
}

func TestSinkToSlice(t *testing.T) {
	m1 := GenerateRandMeasurement("all_at_once", "g000", 100, 0, 0)
	m2 := GenerateRandMeasurement("all_at_once", "g000", 100, 0, 0)
	m3 := GenerateRandMeasurement("all_at_once", "g000", 100, 0, 0)
	expected := []*Measurement{&m1, &m2, &m3}
	in := make(chan *Measurement)
	out := SinkToSlice(context.Background(), in)
	in <- &m1
	in <- &m2
	in <- &m3
	close(in)
	res := <-out
	assert.Equal(t, expected, res)
	_, ok := <-out
	assert.Equal(t, false, ok)
}
func TestSinkToSliceCancel(t *testing.T) {
	m1 := GenerateRandMeasurement("all_at_once", "g000", 100, 0, 0)
	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan *Measurement)
	out := SinkToSlice(ctx, in)
	in <- &m1
	cancel()
	res, ok := <-out
	assert.Equal(t, false, ok)
	assert.Nil(t, res)
}

func TestSplit(t *testing.T) {
	m := GenerateRandMeasurement("all_at_once", "g000", 100, 0, 0)
	in := make(chan *Measurement)
	left, right := Split(context.Background(), in)
	in <- &m
	close(in)
	ml := <-left
	mr := <-right
	assert.Equal(t, &m, ml)
	assert.Equal(t, &m, mr)
	_, lok := <-left
	_, rok := <-right
	assert.False(t, lok)
	assert.False(t, rok)
}
