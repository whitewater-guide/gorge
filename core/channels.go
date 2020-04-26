package core

import "context"

// GenFromSlice creates channel of Measurements from slice
func GenFromSlice(ctx context.Context, measurements []Measurement) <-chan *Measurement {
	out := make(chan *Measurement)
	go func() {
		defer close(out)
		for i := range measurements {
			select {
			case <-ctx.Done():
				return
			case out <- &measurements[i]:
			}
		}
	}()
	return out
}

// SinkToSlice takes channel of measurements and produces channel that will receive
// slice of measurements once input channel is closed or context is canceled
func SinkToSlice(ctx context.Context, in <-chan *Measurement) <-chan []*Measurement {
	out := make(chan []*Measurement)
	go func() {
		defer close(out)
		var res []*Measurement
		for {
			select {
			case m, ok := <-in:
				if !ok {
					out <- res
					return
				}
				res = append(res, m)
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// Cancelable is helper function that cancels channel once context is canceled
func Cancelable(ctx context.Context, c <-chan *Measurement) <-chan *Measurement {
	out := make(chan *Measurement)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-c:
				if !ok {
					return
				}
				select {
				case out <- v:
				case <-ctx.Done():
				}
			}
		}
	}()
	return out
}

// Split splits one channel of measurements into two
func Split(ctx context.Context, in <-chan *Measurement) (<-chan *Measurement, <-chan *Measurement) {
	left := make(chan *Measurement)
	right := make(chan *Measurement)
	go func() {
		defer close(left)
		defer close(right)
		for val := range Cancelable(ctx, in) {
			var left, right = left, right
			for i := 0; i < 2; i++ {
				select {
				case <-ctx.Done():
				case left <- val:
					left = nil
				case right <- val:
					right = nil
				}
			}
		}
	}()
	return left, right
}

// GaugeSinkToSlice converts gauges channels to struct and error
func GaugeSinkToSlice(gauges chan *Gauge, errs chan error) (Gauges, error) {
	var result Gauges
outer:
	for {
		select {
		case err := <-errs:
			if err != nil {
				return nil, err
			}
		case g, ok := <-gauges:
			if g != nil {
				result = append(result, *g)
			}
			if !ok {
				break outer
			}
		}
	}
	return result, nil
}