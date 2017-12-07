// Copyright 2017, OpenCensus Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stackdriver

import (
	"reflect"
	"testing"
	"time"

	distributionpb "google.golang.org/genproto/googleapis/api/distribution"
	"google.golang.org/genproto/googleapis/api/label"

	"go.opencensus.io/stats"
	"go.opencensus.io/tag"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

func TestExporter_makeReq(t *testing.T) {
	m, err := stats.NewMeasureFloat64("test-measure", "measure desc", "unit")
	if err != nil {
		t.Fatal(err)
	}
	defer stats.DeleteMeasure(m)

	key, err := tag.NewKey("test_key")
	if err != nil {
		t.Fatal(err)
	}

	cumView, err := stats.NewView("cumview", "desc", []tag.Key{key}, m, stats.CountAggregation{}, stats.Cumulative{})
	if err != nil {
		t.Fatal(err)
	}
	if err := stats.RegisterView(cumView); err != nil {
		t.Fatal(err)
	}
	defer cumView.Unregister()

	distView, err := stats.NewView("distview", "desc", nil, m, stats.DistributionAggregation([]float64{2, 4, 7}), stats.Interval{})
	if err != nil {
		t.Fatal(err)
	}
	if err := stats.RegisterView(distView); err != nil {
		t.Fatal(err)
	}
	defer distView.Unregister()

	start := time.Now()
	end := start.Add(time.Minute)
	count1 := stats.CountData(10)
	count2 := stats.CountData(16)
	sum1 := stats.SumData(5.5)
	sum2 := stats.SumData(-11.1)
	mean1 := stats.MeanData{
		Mean:  3.3,
		Count: 7,
	}
	mean2 := stats.MeanData{
		Mean:  -7.7,
		Count: 5,
	}

	tests := []struct {
		name   string
		projID string
		vd     *stats.ViewData
		want   []*monitoringpb.CreateTimeSeriesRequest
	}{
		{
			name:   "count agg + cum timeline",
			projID: "proj-id",
			vd:     newTestCumViewData(cumView, start, end, &count1, &count2),
			want: []*monitoringpb.CreateTimeSeriesRequest{{
				Name: monitoring.MetricProjectPath("proj-id"),
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type:   "custom.googleapis.com/opencensus/cumview",
							Labels: map[string]string{"test_key": "test-value-1"},
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "global",
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamp.Timestamp{
										Seconds: start.Unix(),
										Nanos:   int32(start.Nanosecond()),
									},
									EndTime: &timestamp.Timestamp{
										Seconds: end.Unix(),
										Nanos:   int32(end.Nanosecond()),
									},
								},
								Value: &monitoringpb.TypedValue{Value: &monitoringpb.TypedValue_Int64Value{
									Int64Value: 10,
								}},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type:   "custom.googleapis.com/opencensus/cumview",
							Labels: map[string]string{"test_key": "test-value-2"},
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "global",
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamp.Timestamp{
										Seconds: start.Unix(),
										Nanos:   int32(start.Nanosecond()),
									},
									EndTime: &timestamp.Timestamp{
										Seconds: end.Unix(),
										Nanos:   int32(end.Nanosecond()),
									},
								},
								Value: &monitoringpb.TypedValue{Value: &monitoringpb.TypedValue_Int64Value{
									Int64Value: 16,
								}},
							},
						},
					},
				},
			}},
		},
		{
			name:   "sum agg + cum timeline",
			projID: "proj-id",
			vd:     newTestCumViewData(cumView, start, end, &sum1, &sum2),
			want: []*monitoringpb.CreateTimeSeriesRequest{{
				Name: monitoring.MetricProjectPath("proj-id"),
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type:   "custom.googleapis.com/opencensus/cumview",
							Labels: map[string]string{"test_key": "test-value-1"},
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "global",
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamp.Timestamp{
										Seconds: start.Unix(),
										Nanos:   int32(start.Nanosecond()),
									},
									EndTime: &timestamp.Timestamp{
										Seconds: end.Unix(),
										Nanos:   int32(end.Nanosecond()),
									},
								},
								Value: &monitoringpb.TypedValue{Value: &monitoringpb.TypedValue_DoubleValue{
									DoubleValue: 5.5,
								}},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type:   "custom.googleapis.com/opencensus/cumview",
							Labels: map[string]string{"test_key": "test-value-2"},
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "global",
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamp.Timestamp{
										Seconds: start.Unix(),
										Nanos:   int32(start.Nanosecond()),
									},
									EndTime: &timestamp.Timestamp{
										Seconds: end.Unix(),
										Nanos:   int32(end.Nanosecond()),
									},
								},
								Value: &monitoringpb.TypedValue{Value: &monitoringpb.TypedValue_DoubleValue{
									DoubleValue: -11.1,
								}},
							},
						},
					},
				},
			}},
		},
		{
			name:   "mean agg + cum timeline",
			projID: "proj-id",
			vd:     newTestCumViewData(cumView, start, end, &mean1, &mean2),
			want: []*monitoringpb.CreateTimeSeriesRequest{{
				Name: monitoring.MetricProjectPath("proj-id"),
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type:   "custom.googleapis.com/opencensus/cumview",
							Labels: map[string]string{"test_key": "test-value-1"},
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "global",
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamp.Timestamp{
										Seconds: start.Unix(),
										Nanos:   int32(start.Nanosecond()),
									},
									EndTime: &timestamp.Timestamp{
										Seconds: end.Unix(),
										Nanos:   int32(end.Nanosecond()),
									},
								},
								Value: &monitoringpb.TypedValue{Value: &monitoringpb.TypedValue_DistributionValue{
									DistributionValue: &distributionpb.Distribution{
										Count: 7,
										Mean:  3.3,
										SumOfSquaredDeviation: 0,
										BucketOptions: &distributionpb.Distribution_BucketOptions{
											Options: &distributionpb.Distribution_BucketOptions_ExplicitBuckets{
												ExplicitBuckets: &distributionpb.Distribution_BucketOptions_Explicit{
													Bounds: []float64{0},
												},
											},
										},
										BucketCounts: []int64{0, 7},
									},
								}},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type:   "custom.googleapis.com/opencensus/cumview",
							Labels: map[string]string{"test_key": "test-value-2"},
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "global",
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamp.Timestamp{
										Seconds: start.Unix(),
										Nanos:   int32(start.Nanosecond()),
									},
									EndTime: &timestamp.Timestamp{
										Seconds: end.Unix(),
										Nanos:   int32(end.Nanosecond()),
									},
								},
								Value: &monitoringpb.TypedValue{Value: &monitoringpb.TypedValue_DistributionValue{
									DistributionValue: &distributionpb.Distribution{
										Count: 5,
										Mean:  -7.7,
										SumOfSquaredDeviation: 0,
										BucketOptions: &distributionpb.Distribution_BucketOptions{
											Options: &distributionpb.Distribution_BucketOptions_ExplicitBuckets{
												ExplicitBuckets: &distributionpb.Distribution_BucketOptions_Explicit{
													Bounds: []float64{0},
												},
											},
										},
										BucketCounts: []int64{0, 5},
									},
								}},
							},
						},
					},
				},
			}},
		},
		{
			name:   "dist agg + time window",
			projID: "proj-id",
			vd:     newTestDistViewData(distView, start, end),
			want:   []*monitoringpb.CreateTimeSeriesRequest{},
		},
	}
	for _, tt := range tests {
		e := &Exporter{o: Options{ProjectID: tt.projID}}
		resps := e.makeReq([]*stats.ViewData{tt.vd}, maxTimeSeriesPerUpload)
		if got, want := len(resps), len(tt.want); got != want {
			t.Fatalf("%v: Exporter.makeReq() returned %d responses; want %d", tt.name, got, want)
		}
		if len(tt.want) == 0 {
			return
		}
		if !reflect.DeepEqual(resps, tt.want) {
			t.Errorf("%v: Exporter.makeReq() = %v, want %v", tt.name, resps, tt.want)
		}
	}
}

func TestExporter_makeReq_batching(t *testing.T) {
	m, err := stats.NewMeasureFloat64("test-measure", "measure desc", "unit")
	if err != nil {
		t.Fatal(err)
	}
	defer stats.DeleteMeasure(m)

	key, err := tag.NewKey("test_key")
	if err != nil {
		t.Fatal(err)
	}

	view, err := stats.NewView("view", "desc", []tag.Key{key}, m, stats.CountAggregation{}, stats.Cumulative{})
	if err != nil {
		t.Fatal(err)
	}
	if err := stats.RegisterView(view); err != nil {
		t.Fatal(err)
	}
	defer view.Unregister()

	tests := []struct {
		name      string
		iter      int
		limit     int
		wantReqs  int
		wantTotal int
	}{
		{
			name:      "4 vds; 3 limit",
			iter:      2,
			limit:     3,
			wantReqs:  2,
			wantTotal: 4,
		},
		{
			name:      "4 vds; 4 limit",
			iter:      2,
			limit:     4,
			wantReqs:  1,
			wantTotal: 4,
		},
		{
			name:      "4 vds; 5 limit",
			iter:      2,
			limit:     5,
			wantReqs:  1,
			wantTotal: 4,
		},
	}

	count1 := stats.CountData(10)
	count2 := stats.CountData(16)

	for _, tt := range tests {
		var vds []*stats.ViewData
		for i := 0; i < tt.iter; i++ {
			vds = append(vds, newTestCumViewData(view, time.Now(), time.Now(), &count1, &count2))
		}

		e := &Exporter{}
		resps := e.makeReq(vds, tt.limit)
		if len(resps) != tt.wantReqs {
			t.Errorf("%v: got %v; want %d requests", tt.name, resps, tt.wantReqs)
		}

		var total int
		for _, resp := range resps {
			total += len(resp.TimeSeries)
		}
		if got, want := total, tt.wantTotal; got != want {
			t.Errorf("%v: len(resps[...].TimeSeries) = %d; want %d", tt.name, got, want)
		}
	}
}

func TestEqualAggWindowTagKeys(t *testing.T) {
	key1, _ := tag.NewKey("test-key-one")
	key2, _ := tag.NewKey("test-key-two")
	tests := []struct {
		name    string
		md      *metricpb.MetricDescriptor
		agg     stats.Aggregation
		keys    []tag.Key
		window  stats.Window
		wantErr bool
	}{
		{
			name: "count agg + cum",
			md: &metricpb.MetricDescriptor{
				MetricKind: metricpb.MetricDescriptor_CUMULATIVE,
				ValueType:  metricpb.MetricDescriptor_INT64,
			},
			agg:     stats.CountAggregation{},
			window:  stats.Cumulative{},
			wantErr: false,
		},
		{
			name: "sum agg + cum",
			md: &metricpb.MetricDescriptor{
				MetricKind: metricpb.MetricDescriptor_CUMULATIVE,
				ValueType:  metricpb.MetricDescriptor_DOUBLE,
			},
			agg:     stats.SumAggregation{},
			window:  stats.Cumulative{},
			wantErr: false,
		},
		{
			name: "mean agg + cum",
			md: &metricpb.MetricDescriptor{
				MetricKind: metricpb.MetricDescriptor_CUMULATIVE,
				ValueType:  metricpb.MetricDescriptor_DISTRIBUTION,
			},
			agg:     stats.MeanAggregation{},
			window:  stats.Cumulative{},
			wantErr: false,
		},
		{
			name: "distribution agg + cum - mismatch",
			md: &metricpb.MetricDescriptor{
				MetricKind: metricpb.MetricDescriptor_CUMULATIVE,
				ValueType:  metricpb.MetricDescriptor_DISTRIBUTION,
			},
			agg:     stats.CountAggregation{},
			window:  stats.Cumulative{},
			wantErr: true,
		},
		{
			name: "mean agg + cum - mismatch",
			md: &metricpb.MetricDescriptor{
				MetricKind: metricpb.MetricDescriptor_CUMULATIVE,
				ValueType:  metricpb.MetricDescriptor_DOUBLE,
			},
			agg:     stats.MeanAggregation{},
			window:  stats.Cumulative{},
			wantErr: true,
		},
		{
			name: "distribution agg + delta",
			md: &metricpb.MetricDescriptor{
				MetricKind: metricpb.MetricDescriptor_DELTA,
				ValueType:  metricpb.MetricDescriptor_DISTRIBUTION,
			},
			agg:     stats.DistributionAggregation{},
			window:  stats.Interval{},
			wantErr: false,
		},
		{
			name: "distribution agg + cum",
			md: &metricpb.MetricDescriptor{
				MetricKind: metricpb.MetricDescriptor_CUMULATIVE,
				ValueType:  metricpb.MetricDescriptor_DISTRIBUTION,
			},
			agg:     stats.DistributionAggregation{},
			window:  stats.Interval{},
			wantErr: true,
		},
		{
			name: "distribution agg + cum with keys",
			md: &metricpb.MetricDescriptor{
				MetricKind: metricpb.MetricDescriptor_CUMULATIVE,
				ValueType:  metricpb.MetricDescriptor_DISTRIBUTION,
				Labels: []*label.LabelDescriptor{
					{Key: "test_key_one"},
					{Key: "test_key_two"},
				},
			},
			agg:     stats.DistributionAggregation{},
			window:  stats.Cumulative{},
			keys:    []tag.Key{key1, key2},
			wantErr: false,
		},
		{
			name: "distribution agg + cum with keys -- mismatch",
			md: &metricpb.MetricDescriptor{
				MetricKind: metricpb.MetricDescriptor_CUMULATIVE,
				ValueType:  metricpb.MetricDescriptor_DISTRIBUTION,
			},
			agg:     stats.DistributionAggregation{},
			window:  stats.Cumulative{},
			keys:    []tag.Key{key1, key2},
			wantErr: true,
		},
		{
			name: "count agg + cum with pointers",
			md: &metricpb.MetricDescriptor{
				MetricKind: metricpb.MetricDescriptor_CUMULATIVE,
				ValueType:  metricpb.MetricDescriptor_INT64,
			},
			agg:     &stats.CountAggregation{},
			window:  &stats.Cumulative{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		err := equalAggWindowTagKeys(tt.md, tt.agg, tt.window, tt.keys)
		if err != nil && !tt.wantErr {
			t.Errorf("%s:: equalAggWindowTagKeys() = %q; want no error", tt.name, err)
		}
		if err == nil && tt.wantErr {
			t.Errorf("%s:: equalAggWindowTagKeys() = %q; want error", tt.name, err)
		}
	}
}

func newTestCumViewData(v *stats.View, start, end time.Time, data1, data2 stats.AggregationData) *stats.ViewData {
	key, _ := tag.NewKey("test-key")
	tag1 := tag.Tag{Key: key, Value: "test-value-1"}
	tag2 := tag.Tag{Key: key, Value: "test-value-2"}
	return &stats.ViewData{
		View: v,
		Rows: []*stats.Row{
			{
				Tags: []tag.Tag{tag1},
				Data: data1,
			},
			{
				Tags: []tag.Tag{tag2},
				Data: data2,
			},
		},
		Start: start,
		End:   end,
	}
}

func newTestDistViewData(v *stats.View, start, end time.Time) *stats.ViewData {
	return &stats.ViewData{
		View: v,
		Rows: []*stats.Row{
			{Data: &stats.DistributionData{
				Count:           5,
				Min:             1,
				Max:             7,
				Mean:            3,
				SumOfSquaredDev: 1.5,
				CountPerBucket:  []int64{2, 2, 1},
			}},
		},
		Start: start,
		End:   end,
	}
}
