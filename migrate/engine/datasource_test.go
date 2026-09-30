/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package engine

import (
	"errors"
	"path/filepath"
	"testing"

	"hcm/migrate/register"
	"hcm/pkg/cc"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/kit"
	"hcm/pkg/migrate"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelectRegistries(t *testing.T) {
	// register.Registry.database is unexported, so this package cannot build a
	// third database name. Selection is tested with the two real registries.
	// A reversed known list, passed to selectRegistries, checks that output
	// order follows the known list rather than the flag order.

	t.Run("nil values", func(t *testing.T) {
		assertRegistryNames(t, nil, []string{"main", "obs"})
	})
	t.Run("empty slice", func(t *testing.T) {
		assertRegistryNames(t, []string{}, []string{"main", "obs"})
	})

	testCases := []struct {
		name    string
		values  []string
		want    []string
		wantErr bool
		errHas  []string
	}{
		{name: "main", values: []string{"main"}, want: []string{"main"}},
		{name: "obs", values: []string{"obs"}, want: []string{"obs"}},
		{name: "comma main obs", values: []string{"main,obs"}, want: []string{"main", "obs"}},
		{name: "comma obs main still main first", values: []string{"obs,main"}, want: []string{"main", "obs"}},
		{name: "repeated flags", values: []string{"obs", "main"}, want: []string{"main", "obs"}},
		{name: "duplicate flags", values: []string{"main", "main"}, want: []string{"main"}},
		{name: "duplicate in one value", values: []string{"main,main"}, want: []string{"main"}},
		{name: "whitespace around names", values: []string{" main , obs "}, want: []string{"main", "obs"}},
		{name: "padded single flag", values: []string{" main "}, want: []string{"main"}},
		{name: "tab and newline", values: []string{"\tmain\n,\tobs\n"}, want: []string{"main", "obs"}},
		{
			name: "trailing comma", values: []string{"main,"}, wantErr: true,
			errHas: []string{"empty database name"},
		},
		{
			name: "leading comma", values: []string{",main"}, wantErr: true,
			errHas: []string{"empty database name"},
		},
		{
			name: "double comma", values: []string{"main,,obs"}, wantErr: true,
			errHas: []string{"empty database name"},
		},
		{
			name: "empty string element", values: []string{""}, wantErr: true,
			errHas: []string{"empty database name"},
		},
		{
			name: "whitespace element", values: []string{" "}, wantErr: true,
			errHas: []string{"empty database name"},
		},
		{
			name: "unknown foo", values: []string{"foo"}, wantErr: true,
			errHas: []string{"foo", "main, obs"},
		},
		{
			name: "uppercase MAIN", values: []string{"MAIN"}, wantErr: true,
			errHas: []string{"MAIN", "main, obs"},
		},
		{
			name: "uppercase OBS", values: []string{"OBS"}, wantErr: true,
			errHas: []string{"OBS", "supported"},
		},
		{
			name: "main then unknown", values: []string{"main,foo"}, wantErr: true,
			errHas: []string{"foo", "main, obs"},
		},
		{
			name: "all is not a keyword", values: []string{"all"}, wantErr: true,
			errHas: []string{"all", "main, obs"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SelectRegistries(tc.values)
			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				assert.ErrorIs(t, err, migrate.ErrUsage)
				for _, part := range tc.errHas {
					assert.Contains(t, err.Error(), part)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, registryNames(got))
		})
	}

	t.Run("result is a copy of registries", func(t *testing.T) {
		got, err := SelectRegistries(nil)
		require.NoError(t, err)
		require.Equal(t, []*register.Registry{register.Main, register.Obs}, got)
		got[0] = register.Obs
		assert.Same(t, register.Main, registries[0])
		assert.Same(t, register.Obs, registries[1])

		filtered, err := SelectRegistries([]string{"obs", "main"})
		require.NoError(t, err)
		require.Equal(t, []string{"main", "obs"}, registryNames(filtered))
		filtered[0] = nil
		assert.Same(t, register.Main, registries[0])
	})

	t.Run("order follows the known list", func(t *testing.T) {
		known := []*register.Registry{register.Obs, register.Main}
		got, err := selectRegistries(known, []string{"main", "obs"})
		require.NoError(t, err)
		assert.Equal(t, []string{"obs", "main"}, registryNames(got))
		got[0] = nil
		assert.Same(t, register.Obs, known[0])
		assert.Same(t, register.Main, known[1])
	})
}

func assertRegistryNames(t *testing.T, values []string, want []string) {
	t.Helper()
	got, err := SelectRegistries(values)
	require.NoError(t, err)
	assert.Equal(t, want, registryNames(got))
}

func registryNames(regs []*register.Registry) []string {
	names := make([]string, 0, len(regs))
	for _, reg := range regs {
		names = append(names, reg.Database())
	}
	return names
}

func TestDataSource_Open(t *testing.T) {
	kt := kit.New()
	mainOpt := &cc.DataBase{Resource: cc.ResourceDB{
		Endpoints: []string{"10.1.2.3:3306", "10.1.2.4:3306"},
		Database:  "hcm_app",
		User:      "migrate",
		Password:  "s3cr3t-pw",
	}}

	t.Run("nil config skips without opening", func(t *testing.T) {
		opened := 0
		ds := &DataSource{
			configs: map[string]*cc.DataBase{"obs": nil},
			open: func(cc.DataBase) (orm.Interface, error) {
				opened++
				return newFakeOrm(newFakeDo()), nil
			},
		}
		o, ok, err := ds.Open(kt, register.Obs)
		require.NoError(t, err)
		assert.False(t, ok)
		assert.Nil(t, o)
		assert.Equal(t, 0, opened)
	})

	t.Run("unknown mapping", func(t *testing.T) {
		opened := 0
		ds := &DataSource{
			configs: map[string]*cc.DataBase{"main": mainOpt},
			open: func(cc.DataBase) (orm.Interface, error) {
				opened++
				return nil, nil
			},
		}
		// An empty registry has database name "", which is not a key.
		o, ok, err := ds.Open(kt, &register.Registry{})
		require.Error(t, err)
		assert.False(t, ok)
		assert.Nil(t, o)
		assert.Equal(t, 0, opened)
		assert.ErrorIs(t, err, migrate.ErrUsage)
		assert.Contains(t, err.Error(), "has no config mapping")
	})

	t.Run("open error names the database and hides the password", func(t *testing.T) {
		opened := 0
		ds := &DataSource{
			configs: map[string]*cc.DataBase{"main": mainOpt},
			open: func(opt cc.DataBase) (orm.Interface, error) {
				opened++
				assert.Equal(t, "s3cr3t-pw", opt.Resource.Password)
				return nil, errors.New("dial timeout")
			},
		}
		o, ok, err := ds.Open(kt, register.Main)
		require.Error(t, err)
		assert.False(t, ok)
		assert.Nil(t, o)
		assert.Equal(t, 1, opened)
		assert.Contains(t, err.Error(), "main")
		assert.Contains(t, err.Error(), "hcm_app")
		assert.Contains(t, err.Error(), "10.1.2.3:3306")
		assert.Contains(t, err.Error(), "10.1.2.4:3306")
		assert.Contains(t, err.Error(), "dial timeout")
		assert.NotContains(t, err.Error(), "s3cr3t-pw")
		assertNoSentinel(t, err)
	})

	t.Run("success returns the opened orm once", func(t *testing.T) {
		opened := 0
		want := newFakeOrm(newFakeDo())
		ds := &DataSource{
			configs: map[string]*cc.DataBase{"main": mainOpt},
			open: func(cc.DataBase) (orm.Interface, error) {
				opened++
				return want, nil
			},
		}
		o, ok, err := ds.Open(kt, register.Main)
		require.NoError(t, err)
		assert.True(t, ok)
		assert.Same(t, want, o)
		assert.Equal(t, 1, opened)
	})
}

func TestNewDataSource(t *testing.T) {
	mainDB := cc.DataBase{Resource: cc.ResourceDB{
		Endpoints: []string{"10.0.0.1:3306"},
		Database:  "hcm_app",
		Password:  "main-pw",
	}}

	t.Run("obs not configured", func(t *testing.T) {
		opened := 0
		marker := errors.New("open marker")
		ds := newDataSource(cc.DataServiceSetting{Database: mainDB}, func(cc.DataBase) (orm.Interface, error) {
			opened++
			return nil, marker
		})
		assertConfigKeys(t, ds)
		require.NotNil(t, ds.configs["main"])
		assert.Equal(t, mainDB.Resource.Endpoints, ds.configs["main"].Resource.Endpoints)
		assert.Equal(t, mainDB.Resource.Database, ds.configs["main"].Resource.Database)
		assert.Equal(t, mainDB.Resource.Password, ds.configs["main"].Resource.Password)
		assert.Nil(t, ds.configs["obs"])

		_, err := ds.open(cc.DataBase{})
		require.ErrorIs(t, err, marker)
		assert.Equal(t, 1, opened)
	})

	t.Run("obs configured", func(t *testing.T) {
		obsDB := &cc.DataBase{Resource: cc.ResourceDB{
			Endpoints: []string{"10.9.9.9:3306"},
			Database:  "hcm_obs_app",
			Password:  "obs-pw",
		}}
		var got cc.DataBase
		opened := 0
		ds := newDataSource(cc.DataServiceSetting{Database: mainDB, OBSDatabase: obsDB}, func(opt cc.DataBase) (orm.Interface, error) {
			opened++
			got = opt
			return newFakeOrm(newFakeDo()), nil
		})
		assertConfigKeys(t, ds)
		require.NotNil(t, ds.configs["main"])
		assert.Equal(t, mainDB.Resource.Endpoints, ds.configs["main"].Resource.Endpoints)
		assert.Equal(t, mainDB.Resource.Database, ds.configs["main"].Resource.Database)
		assert.Equal(t, mainDB.Resource.Password, ds.configs["main"].Resource.Password)
		require.NotNil(t, ds.configs["obs"])
		assert.Equal(t, obsDB.Resource.Endpoints, ds.configs["obs"].Resource.Endpoints)
		assert.Equal(t, obsDB.Resource.Database, ds.configs["obs"].Resource.Database)
		assert.Equal(t, obsDB.Resource.Password, ds.configs["obs"].Resource.Password)

		o, ok, err := ds.Open(kit.New(), register.Obs)
		require.NoError(t, err)
		assert.True(t, ok)
		assert.NotNil(t, o)
		assert.Equal(t, 1, opened)
		assert.Equal(t, obsDB.Resource.Endpoints, got.Resource.Endpoints)
		assert.Equal(t, obsDB.Resource.Database, got.Resource.Database)
		assert.Equal(t, obsDB.Resource.Password, got.Resource.Password)
		assert.NotEqual(t, mainDB.Resource.Database, got.Resource.Database)
		assert.NotEqual(t, mainDB.Resource.Password, got.Resource.Password)
	})
}

func assertConfigKeys(t *testing.T, ds *DataSource) {
	t.Helper()
	require.Len(t, ds.configs, 2)
	_, mainOK := ds.configs["main"]
	_, obsOK := ds.configs["obs"]
	assert.True(t, mainOK)
	assert.True(t, obsOK)
}

func TestLoadDataSource(t *testing.T) {
	// Failure paths only. LoadDataSource calls cc.InitService, which uses
	// sync.Once and sets time.Local to UTC. These cases never point at a
	// real config file, so LoadSettings returns before InitRuntime.
	testCases := []struct {
		name string
		path string
	}{
		{name: "empty path", path: ""},
		{name: "missing file", path: filepath.Join(t.TempDir(), "no-such-dataservice.yaml")},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ds, err := LoadDataSource(tc.path)
			require.Error(t, err)
			assert.Nil(t, ds)
			assert.ErrorIs(t, err, migrate.ErrUsage)
			assert.Contains(t, err.Error(), "load config file")
			if tc.path != "" {
				assert.Contains(t, err.Error(), tc.path)
			} else {
				assert.Contains(t, err.Error(), `load config file ""`)
			}
		})
	}
}
