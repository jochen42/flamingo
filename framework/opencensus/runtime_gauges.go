package opencensus

import (
	"context"
	"runtime"
	"time"

	"github.com/pkg/errors"
	"go.opencensus.io/metric"
	"go.opencensus.io/metric/metricdata"
)

type (
	// RuntimeGauges collects runtime metrics in gauge entries.
	RuntimeGauges struct {
		heapAlloc     *metric.Int64GaugeEntry
		heapObjects   *metric.Int64GaugeEntry
		heapReleased  *metric.Int64GaugeEntry
		stackSys      *metric.Int64GaugeEntry
		ptrLookups    *metric.Int64GaugeEntry
		numGoroutines *metric.Int64GaugeEntry
	}

	// GaugeOptions options for runtime gauge metrics
	GaugeOptions struct {
		Prefix string
	}
)

// NewRuntimeGauges create a new runtimeGauges
func NewRuntimeGauges(reg *metric.Registry, gaugeOptions GaugeOptions) (*RuntimeGauges, error) {
	opt := gaugeOptions

	rg := new(RuntimeGauges)
	var err error

	rg.heapAlloc, err = createInt64GaugeEntry(reg, opt.Prefix+"heap_alloc", "Process heap allocation", metricdata.UnitBytes)
	if err != nil {
		return nil, err
	}

	rg.heapObjects, err = createInt64GaugeEntry(reg, opt.Prefix+"heap_objects", "The number of objects allocated on the heap", metricdata.UnitDimensionless)
	if err != nil {
		return nil, err
	}

	rg.heapReleased, err = createInt64GaugeEntry(reg, opt.Prefix+"heap_release", "The number of objects released from the heap", metricdata.UnitBytes)
	if err != nil {
		return nil, err
	}

	rg.stackSys, err = createInt64GaugeEntry(reg, opt.Prefix+"stack_sys", "The memory used by stack spans and OS thread stacks", metricdata.UnitBytes)
	if err != nil {
		return nil, err
	}

	rg.ptrLookups, err = createInt64GaugeEntry(reg, opt.Prefix+"ptr_lookups", "The number of pointer lookups", metricdata.UnitDimensionless)
	if err != nil {
		return nil, err
	}

	rg.numGoroutines, err = createInt64GaugeEntry(reg, opt.Prefix+"num_goroutines", "Number of current goroutines", metricdata.UnitDimensionless)
	if err != nil {
		return nil, err
	}

	return rg, nil
}

// StartRecording starts recoding of runtime metrics
func (r *RuntimeGauges) StartRecording(ctx context.Context, delay time.Duration) {
	mem := &runtime.MemStats{}

	tick := time.NewTicker(delay)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-tick.C:
			runtime.ReadMemStats(mem)
			r.heapAlloc.Set(int64(mem.HeapAlloc))
			r.heapObjects.Set(int64(mem.HeapObjects))
			r.heapReleased.Set(int64(mem.HeapReleased))
			r.stackSys.Set(int64(mem.StackSys))
			r.ptrLookups.Set(int64(mem.Lookups))

			r.numGoroutines.Set(int64(runtime.NumGoroutine()))
		}
	}
}

func createInt64GaugeEntry(reg *metric.Registry, name string, description string, unit metricdata.Unit) (*metric.Int64GaugeEntry, error) {
	gauge, err := reg.AddInt64Gauge(
		name,
		metric.WithDescription(description),
		metric.WithUnit(unit))
	if err != nil {
		return nil, errors.WithMessage(err, "error creating gauge for "+name)
	}

	entry, err := gauge.GetEntry()
	if err != nil {
		return nil, errors.WithMessage(err, "error getting gauge entry for "+name)
	}

	return entry, nil
}
