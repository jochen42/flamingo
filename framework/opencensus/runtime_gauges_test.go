package opencensus

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"contrib.go.opencensus.io/exporter/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opencensus.io/metric"
	"go.opencensus.io/metric/metricdata"
	"go.opencensus.io/metric/metricexport"
	"go.opencensus.io/metric/metricproducer"
)

type testExporter struct {
	data []*metricdata.Metric
}

func (t *testExporter) ExportMetrics(ctx context.Context, data []*metricdata.Metric) error {
	t.data = append(t.data, data...)
	return nil
}

func TestRuntimeGauges(t *testing.T) {
	tests := []struct {
		name          string
		options       GaugeOptions
		expectedNames []string
	}{
		{
			"custom prefix",
			GaugeOptions{Prefix: "test_"},
			[]string{"test_heap_alloc", "test_heap_objects", "test_heap_release", "test_stack_sys", "test_ptr_lookups", "test_num_goroutines"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reg := metric.NewRegistry()
			metricproducer.GlobalManager().AddProducer(reg)
			defer metricproducer.GlobalManager().DeleteProducer(reg)

			gauges, err := NewRuntimeGauges(reg, test.options)
			require.NoError(t, err)

			ctx, cancel := context.WithCancel(context.Background())
			time.AfterFunc(2*time.Second, func() {
				cancel()
			})
			go gauges.StartRecording(ctx, 1*time.Second)

			exporter := &testExporter{}
			reader := metricexport.NewReader()
			reader.ReadAndExport(exporter)

			assertNames(t, exporter, test.expectedNames)
		})
	}
}

func assertNames(t *testing.T, exporter *testExporter, expectedNames []string) {
	metricNames := make([]string, 0)
	for _, v := range exporter.data {
		metricNames = append(metricNames, v.Descriptor.Name)
	}
	assert.ElementsMatchf(t, expectedNames, metricNames, "actual: %v", metricNames)
}

func TestRuntimeGauges_WithPrometheus(t *testing.T) {
	reg := metric.NewRegistry()
	metricproducer.GlobalManager().AddProducer(reg)
	defer metricproducer.GlobalManager().DeleteProducer(reg)

	gauges, err := NewRuntimeGauges(reg, GaugeOptions{Prefix: "test_"})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(2*time.Second, func() {
		cancel()
	})
	go gauges.StartRecording(ctx, 1*time.Second)

	exporter, err := prometheus.NewExporter(prometheus.Options{})
	require.NoError(t, err)

	server := httptest.NewServer(exporter)
	defer server.Close()

	// wait for at least one metric to be written
	<-time.After(1 * time.Second)

	resp, err := http.Get(server.URL)
	require.NoError(t, err)

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	strBody := string(bytes)
	assert.Regexp(t, "test_heap_alloc \\d+", strBody)
	assert.Regexp(t, "test_heap_objects \\d+", strBody)
	assert.Regexp(t, "test_heap_release \\d+", strBody)
	assert.Regexp(t, "test_stack_sys \\d+", strBody)
	assert.Regexp(t, "test_ptr_lookups \\d+", strBody)
	assert.Regexp(t, "test_num_goroutines \\d+", strBody)
}
