// Copyright 2024 TiKV Project Authors.
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

package storage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/pingcap/failpoint"
	"github.com/pingcap/kvproto/pkg/metapb"

	"github.com/tikv/pd/pkg/core"
	"github.com/tikv/pd/pkg/storage/endpoint"
	"github.com/tikv/pd/pkg/storage/kv"
)

func TestRegionReload(t *testing.T) {
	re := require.New(t)
	se := endpoint.NewStorageEndpoint(kv.NewMemoryKV(), nil)
	regionCount := 10
	for id := range regionCount {
		region := newTestRegionMeta(uint64(id))
		re.NoError(se.SaveRegion(region))
	}
	re.Equal(uint64(0), se.NextRegionID())

	// load region will fail after load 5 regions
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	re.NoError(failpoint.Enable("github.com/tikv/pd/pkg/storage/endpoint/slowLoadRegion", "return(true)"))
	defer func() {
		re.NoError(failpoint.Disable("github.com/tikv/pd/pkg/storage/endpoint/slowLoadRegion"))
	}()
	re.Error(se.LoadRegions(ctx, func(_ *core.RegionInfo) []*core.RegionInfo {
		return nil
	}))
	re.NotEqual(uint64(regionCount), se.NextRegionID())

	// load regions will success and the loading cost will less than 10s
	ctx, cancel = context.WithTimeout(context.Background(), time.Second*8)
	defer cancel()
	re.NoError(se.LoadRegions(ctx, func(_ *core.RegionInfo) []*core.RegionInfo {
		return nil
	}))
	re.Equal(uint64(regionCount), se.NextRegionID())
}

func TestRegionStorage(t *testing.T) {
	re := require.New(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var (
		regionStorage endpoint.RegionStorage
		err           error
	)
	regionStorage, err = NewRegionStorageWithLevelDBBackend(ctx, t.TempDir(), nil)
	re.NoError(err)
	re.NotNil(regionStorage)
	// Load regions from the storage.
	regions := make([]*core.RegionInfo, 0)
	appendRegionFunc := func(region *core.RegionInfo) []*core.RegionInfo {
		regions = append(regions, region)
		return nil
	}
	err = regionStorage.LoadRegions(ctx, appendRegionFunc)
	re.NoError(err)
	re.Empty(regions)
	// Save regions to the storage.
	region1 := newTestRegionMeta(1)
	err = regionStorage.SaveRegion(region1)
	re.NoError(err)
	region2 := newTestRegionMeta(2)
	err = regionStorage.SaveRegion(region2)
	re.NoError(err)
	regions = make([]*core.RegionInfo, 0)
	err = regionStorage.LoadRegions(ctx, appendRegionFunc)
	re.NoError(err)
	re.Empty(regions)
	// Flush and load.
	err = regionStorage.Flush()
	re.NoError(err)
	regions = make([]*core.RegionInfo, 0)
	err = regionStorage.LoadRegions(ctx, appendRegionFunc)
	re.NoError(err)
	re.Len(regions, 2)
	re.Equal(region1, regions[0].GetMeta())
	re.Equal(region2, regions[1].GetMeta())
	newRegion := &metapb.Region{}
	ok, err := regionStorage.LoadRegion(3, newRegion)
	re.NoError(err)
	re.False(ok)
	ok, err = regionStorage.LoadRegion(1, newRegion)
	re.NoError(err)
	re.True(ok)
	re.Equal(region1, newRegion)
	ok, err = regionStorage.LoadRegion(2, newRegion)
	re.NoError(err)
	re.True(ok)
	re.Equal(region2, newRegion)
	// Delete and load.
	ok, err = regionStorage.LoadRegion(2, newRegion)
	re.NoError(err)
	re.True(ok)
	re.Equal(region2, newRegion)
	re.Equal(regions[1].GetMeta(), newRegion)
	// Close the storage.
	err = regionStorage.Close()
	re.NoError(err)
}
