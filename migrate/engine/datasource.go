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
	"fmt"
	"slices"
	"strings"

	"hcm/migrate/register"
	"hcm/pkg/cc"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// DataSource holds the database configs loaded from the data-service config
// file and opens the database of a registry from them.
type DataSource struct {
	// configs maps a registry's database name to its config. A nil value
	// means the database is optional and not configured.
	configs map[string]*cc.DataBase
	open    func(opt cc.DataBase) (orm.Interface, error)
}

// LoadDataSource loads the data-service config file once and keeps the
// database configs. The service name must be set first, otherwise the file is
// not decoded as a DataServiceSetting.
func LoadDataSource(configFile string) (*DataSource, error) {
	cc.InitService(cc.DataServiceName)
	if err := cc.LoadSettings(&cc.SysOption{ConfigFile: configFile}); err != nil {
		return nil, fmt.Errorf("%w: load config file %q failed, err: %v", ErrUsage, configFile, err)
	}
	return newDataSource(cc.DataService(), openOrm), nil
}

// newDataSource maps every registry's database name to its config in setting.
func newDataSource(setting cc.DataServiceSetting, open func(opt cc.DataBase) (orm.Interface, error)) *DataSource {
	return &DataSource{
		configs: map[string]*cc.DataBase{
			register.Main.Database(): &setting.Database,
			register.Obs.Database():  setting.OBSDatabase,
		},
		open: open,
	}
}

// Open connects to the database of reg and returns the orm migrations run
// with. ok is false when the database is not configured, e.g. obs on an
// edition without obsDatabase; the caller skips that database and it is not
// an error. The connection is probed once, with no retry: waiting for the
// database is the job of the deployment's initContainer.
func (d *DataSource) Open(kt *kit.Kit, reg *register.Registry) (o orm.Interface, ok bool, err error) {
	opt, known := d.configs[reg.Database()]
	if !known {
		return nil, false, fmt.Errorf("%w: database %s has no config mapping", ErrUsage, reg.Database())
	}
	if opt == nil {
		logs.Warnf("database %s is not configured in config file, skip it and all its migrations, rid: %s",
			reg.Database(), kt.Rid)
		return nil, false, nil
	}

	o, err = d.open(*opt)
	if err != nil {
		logs.Errorf("connect database failed, err: %v, database: %s, endpoints: %v, db: %s, rid: %s",
			err, reg.Database(), opt.Resource.Endpoints, opt.Resource.Database, kt.Rid)
		return nil, false, fmt.Errorf("connect %s database %s at %v failed, err: %v",
			reg.Database(), opt.Resource.Database, opt.Resource.Endpoints, err)
	}
	return o, true, nil
}

// openOrm connects with a single ping inside dao.NewDaoSet and returns the
// set's orm. That orm has no ModifySQLOpts attached, so its Do() sends every
// statement as built and tenant SQL rewriting never renames a table.
func openOrm(opt cc.DataBase) (orm.Interface, error) {
	set, err := dao.NewDaoSet(opt)
	if err != nil {
		return nil, err
	}
	return set.GetOrm(), nil
}

// registries lists every migrated database in execution order: main first,
// then obs, so logs and troubleshooting are predictable.
var registries = []*register.Registry{register.Main, register.Obs}

// SelectRegistries returns the registries named by the --database values, in
// execution order. Each value may hold several comma separated names and may
// be repeated. No value selects every database. An empty or unknown name is a
// usage error.
func SelectRegistries(values []string) ([]*register.Registry, error) {
	return selectRegistries(registries, values)
}

func selectRegistries(known []*register.Registry, values []string) ([]*register.Registry, error) {
	if len(values) == 0 {
		return slices.Clone(known), nil
	}

	names := make([]string, 0, len(known))
	for _, r := range known {
		names = append(names, r.Database())
	}

	wanted := make(map[string]struct{})
	for _, value := range values {
		for _, name := range strings.Split(value, ",") {
			name = strings.TrimSpace(name)
			if name == "" {
				return nil, fmt.Errorf("%w: --database %q contains an empty database name", ErrUsage, value)
			}
			if !slices.Contains(names, name) {
				return nil, fmt.Errorf("%w: unknown database %q, supported: %s",
					ErrUsage, name, strings.Join(names, ", "))
			}
			wanted[name] = struct{}{}
		}
	}

	out := make([]*register.Registry, 0, len(wanted))
	for _, r := range known {
		if _, ok := wanted[r.Database()]; ok {
			out = append(out, r)
		}
	}
	return out, nil
}
