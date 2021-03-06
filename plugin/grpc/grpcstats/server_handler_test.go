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
//

package grpcstats

import (
	"errors"
	"testing"

	"golang.org/x/net/context"

	istats "go.opencensus.io/stats"
	"go.opencensus.io/tag"

	"google.golang.org/grpc/stats"
)

func TestServerDefaultCollections(t *testing.T) {
	k1, _ := tag.NewKey("k1")
	k2, _ := tag.NewKey("k2")

	type tagPair struct {
		k tag.Key
		v string
	}

	type wantData struct {
		v    func() *istats.View
		rows []*istats.Row
	}
	type rpc struct {
		tags        []tagPair
		tagInfo     *stats.RPCTagInfo
		inPayloads  []*stats.InPayload
		outPayloads []*stats.OutPayload
		end         *stats.End
	}

	type testCase struct {
		label string
		rpcs  []*rpc
		wants []*wantData
	}
	tcs := []testCase{
		{
			"1",
			[]*rpc{
				{
					[]tagPair{{k1, "v1"}},
					&stats.RPCTagInfo{FullMethodName: "/package.service/method"},
					[]*stats.InPayload{
						{Length: 10},
					},
					[]*stats.OutPayload{
						{Length: 10},
					},
					&stats.End{Error: nil},
				},
			},
			[]*wantData{
				{
					func() *istats.View { return RPCServerRequestCountView },
					[]*istats.Row{
						{
							[]tag.Tag{
								{Key: keyMethod, Value: "method"},
								{Key: keyService, Value: "package.service"},
							},
							newDistributionData(rpcCountBucketBoundaries, []int64{0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, 1, 1, 1, 1, 0),
						},
					},
				},
				{
					func() *istats.View { return RPCServerResponseCountView },
					[]*istats.Row{
						{
							[]tag.Tag{
								{Key: keyMethod, Value: "method"},
								{Key: keyService, Value: "package.service"},
							},
							newDistributionData(rpcCountBucketBoundaries, []int64{0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, 1, 1, 1, 1, 0),
						},
					},
				},
				{
					func() *istats.View { return RPCServerRequestBytesView },
					[]*istats.Row{
						{
							[]tag.Tag{
								{Key: keyMethod, Value: "method"},
								{Key: keyService, Value: "package.service"},
							},
							newDistributionData(rpcBytesBucketBoundaries, []int64{0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, 1, 10, 10, 10, 0),
						},
					},
				},
				{
					func() *istats.View { return RPCServerResponseBytesView },
					[]*istats.Row{
						{
							[]tag.Tag{
								{Key: keyMethod, Value: "method"},
								{Key: keyService, Value: "package.service"},
							},
							newDistributionData(rpcBytesBucketBoundaries, []int64{0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, 1, 10, 10, 10, 0),
						},
					},
				},
			},
		},
		{
			"2",
			[]*rpc{
				{
					[]tagPair{{k1, "v1"}},
					&stats.RPCTagInfo{FullMethodName: "/package.service/method"},
					[]*stats.InPayload{
						{Length: 10},
					},
					[]*stats.OutPayload{
						{Length: 10},
						{Length: 10},
						{Length: 10},
					},
					&stats.End{Error: nil},
				},
				{
					[]tagPair{{k1, "v11"}},
					&stats.RPCTagInfo{FullMethodName: "/package.service/method"},
					[]*stats.InPayload{
						{Length: 10},
						{Length: 10},
					},
					[]*stats.OutPayload{
						{Length: 10},
						{Length: 10},
					},
					&stats.End{Error: errors.New("someError")},
				},
			},
			[]*wantData{
				{
					func() *istats.View { return RPCServerErrorCountView },
					[]*istats.Row{
						{
							[]tag.Tag{
								{Key: keyMethod, Value: "method"},
								{Key: keyOpStatus, Value: "someError"},
								{Key: keyService, Value: "package.service"},
							},
							newCountData(1),
						},
					},
				},
				{
					func() *istats.View { return RPCServerRequestCountView },
					[]*istats.Row{
						{
							[]tag.Tag{
								{Key: keyMethod, Value: "method"},
								{Key: keyService, Value: "package.service"},
							},
							newDistributionData(rpcCountBucketBoundaries, []int64{0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, 2, 1, 2, 1.5, 0.5),
						},
					},
				},
				{
					func() *istats.View { return RPCServerResponseCountView },
					[]*istats.Row{
						{
							[]tag.Tag{
								{Key: keyMethod, Value: "method"},
								{Key: keyService, Value: "package.service"},
							},
							newDistributionData(rpcCountBucketBoundaries, []int64{0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, 2, 2, 3, 2.5, 0.5),
						},
					},
				},
			},
		},
		{
			"3",
			[]*rpc{
				{
					[]tagPair{{k1, "v1"}},
					&stats.RPCTagInfo{FullMethodName: "/package.service/method"},
					[]*stats.InPayload{
						{Length: 1},
					},
					[]*stats.OutPayload{
						{Length: 1},
						{Length: 1024},
						{Length: 65536},
					},
					&stats.End{Error: nil},
				},
				{
					[]tagPair{{k1, "v1"}, {k2, "v2"}},
					&stats.RPCTagInfo{FullMethodName: "/package.service/method"},
					[]*stats.InPayload{
						{Length: 1024},
					},
					[]*stats.OutPayload{
						{Length: 4096},
						{Length: 16384},
					},
					&stats.End{Error: errors.New("someError1")},
				},
				{
					[]tagPair{{k1, "v11"}, {k2, "v22"}},
					&stats.RPCTagInfo{FullMethodName: "/package.service/method"},
					[]*stats.InPayload{
						{Length: 2048},
						{Length: 16384},
					},
					[]*stats.OutPayload{
						{Length: 2048},
						{Length: 4096},
						{Length: 16384},
					},
					&stats.End{Error: errors.New("someError2")},
				},
			},
			[]*wantData{
				{
					func() *istats.View { return RPCServerErrorCountView },
					[]*istats.Row{
						{
							[]tag.Tag{
								{Key: keyMethod, Value: "method"},
								{Key: keyOpStatus, Value: "someError1"},
								{Key: keyService, Value: "package.service"},
							},
							newCountData(1),
						},
						{
							[]tag.Tag{
								{Key: keyMethod, Value: "method"},
								{Key: keyOpStatus, Value: "someError2"},
								{Key: keyService, Value: "package.service"},
							},
							newCountData(1),
						},
					},
				},
				{
					func() *istats.View { return RPCServerRequestCountView },
					[]*istats.Row{
						{
							[]tag.Tag{
								{Key: keyMethod, Value: "method"},
								{Key: keyService, Value: "package.service"},
							},
							newDistributionData(rpcCountBucketBoundaries, []int64{0, 0, 2, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, 3, 1, 2, 1.333333333, 0.333333333*2),
						},
					},
				},
				{
					func() *istats.View { return RPCServerResponseCountView },
					[]*istats.Row{
						{
							[]tag.Tag{
								{Key: keyMethod, Value: "method"},
								{Key: keyService, Value: "package.service"},
							},
							newDistributionData(rpcCountBucketBoundaries, []int64{0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, 3, 2, 3, 2.666666666, 0.333333333*2),
						},
					},
				},
				{
					func() *istats.View { return RPCServerRequestBytesView },
					[]*istats.Row{
						{
							[]tag.Tag{
								{Key: keyMethod, Value: "method"},
								{Key: keyService, Value: "package.service"},
							},
							newDistributionData(rpcBytesBucketBoundaries, []int64{0, 1, 1, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0}, 4, 1, 16384, 4864.25, 59678208.25*3),
						},
					},
				},
				{
					func() *istats.View { return RPCServerResponseBytesView },
					[]*istats.Row{
						{
							[]tag.Tag{
								{Key: keyMethod, Value: "method"},
								{Key: keyService, Value: "package.service"},
							},
							newDistributionData(rpcBytesBucketBoundaries, []int64{0, 1, 1, 1, 2, 2, 1, 0, 0, 0, 0, 0, 0, 0, 0}, 8, 1, 65536, 13696.125, 481423542.982143*7),
						},
					},
				},
			},
		},
	}

	for _, tc := range tcs {
		for _, v := range serverViews {
			if err := v.Subscribe(); err != nil {
				t.Error(err)
			}
		}

		h := &ServerStatsHandler{}
		for _, rpc := range tc.rpcs {
			mods := []tag.Mutator{}
			for _, t := range rpc.tags {
				mods = append(mods, tag.Upsert(t.k, t.v))
			}
			ts, err := tag.NewMap(context.Background(), mods...)
			if err != nil {
				t.Errorf("%q: NewMap = %v", tc.label, err)
			}
			encoded := tag.Encode(ts)
			ctx := stats.SetTags(context.Background(), encoded)

			ctx = h.TagRPC(ctx, rpc.tagInfo)

			for _, in := range rpc.inPayloads {
				h.HandleRPC(ctx, in)
			}

			for _, out := range rpc.outPayloads {
				h.HandleRPC(ctx, out)
			}

			h.HandleRPC(ctx, rpc.end)
		}

		for _, wantData := range tc.wants {
			gotRows, err := wantData.v().RetrieveData()
			if err != nil {
				t.Errorf("%q: RetrieveData (%q) = %v", tc.label, wantData.v().Name(), err)
				continue
			}

			for _, gotRow := range gotRows {
				if !containsRow(wantData.rows, gotRow) {
					t.Errorf("%q: unwanted row for view %q: %v", tc.label, wantData.v().Name(), gotRow)
					break
				}
			}

			for _, wantRow := range wantData.rows {
				if !containsRow(gotRows, wantRow) {
					t.Errorf("%q: missing row for view %q: %v", tc.label, wantData.v().Name(), wantRow)
					break
				}
			}
		}

		// Unregister views to cleanup.
		for _, v := range serverViews {
			if err := v.Unsubscribe(); err != nil {
				t.Error(err)
			}
		}
	}
}

func newCountData(v int) *istats.CountData {
	cav := istats.CountData(v)
	return &cav
}

func newDistributionData(bounds []float64, countPerBucket []int64, count int64, min, max, mean, sumOfSquaredDev float64) *istats.DistributionData {
	return &istats.DistributionData{
		Count:           count,
		Min:             min,
		Max:             max,
		Mean:            mean,
		SumOfSquaredDev: sumOfSquaredDev,
		CountPerBucket:  countPerBucket,
	}
}
