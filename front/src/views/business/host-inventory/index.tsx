import { defineComponent, ref, onMounted, computed } from 'vue';
import { useTable } from '@/hooks/useTable/useTable';
import apiService from '@/api/scrApi';
import { Button } from 'bkui-vue';
import AreaSelector from '@/views/ziyanScr/hostApplication/components/AreaSelector';
import ZoneSelector from '@/views/ziyanScr/hostApplication/components/ZoneSelector';
import { useI18n } from 'vue-i18n';
import cssModule from './index.module.scss';
import GridFilterComp from '@/components/grid-filter-comp';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import DevicetypeSelector from '@/views/ziyanScr/components/devicetype-selector/index.vue';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import routerAction from '@/router/utils/action';
import { GLOBAL_BIZS_KEY, VendorEnum } from '@/common/constant';

export default defineComponent({
  name: 'BusinessHostInventory',
  setup() {
    const { columns } = useColumns('hostInventor');
    const deviceGroups = ['标准型', '高IO型', '大数据型', '计算型'];
    const { t } = useI18n();
    const filter = ref({
      region: [],
      zone: [],
      device_type: [],
      device_group: deviceGroups && [deviceGroups[0]],
      cpu: '',
      mem: '',
      disk: '',
    });
    const options = ref({
      device_groups: deviceGroups,
      device_types: [],
      regions: [],
      zones: [],
      cpu: [],
      mem: [],
    });
    const deviceConfigDisabled = ref(false);
    const deviceTypeDisabled = ref(false);
    const page = ref({
      limit: 50,
      start: 0,
    });
    const queryRules = ref(
      [
        filter.value.region.length && { field: 'dc.region', op: 'in', value: filter.value.region },
        filter.value.zone.length && { field: 'dc.zone', op: 'in', value: filter.value.zone },
        filter.value.device_group.length && {
          field: 'device_family',
          op: 'in',
          value: filter.value.device_group,
        },
        filter.value.device_type.length && { field: 'dc.device_type', op: 'in', value: filter.value.device_type },
        filter.value.cpu && { field: 'cpu_core', op: 'eq', value: filter.value.cpu },
        filter.value.mem && { field: 'memory', op: 'eq', value: filter.value.mem },
      ].filter(Boolean),
    );
    const loadResources = () => {
      getListData();
    };

    const whereAmI = useWhereAmI();

    const emptyform = () => {
      filter.value = {
        region: [],
        zone: [],
        device_type: [],
        device_group: deviceGroups && [deviceGroups[0]],
        cpu: '',
        mem: '',
        disk: '',
      };
    };
    const handleDeviceConfigChange = () => {
      filter.value.device_type = [];
      const { cpu, mem } = filter.value;
      deviceTypeDisabled.value = Boolean(cpu || mem);
    };
    const clearFilter = () => {
      emptyform();
      deviceConfigDisabled.value = false;
      deviceTypeDisabled.value = false;
      filterDevices();
    };
    const handleDeviceGroupChange = () => {
      filter.value.cpu = '';
      filter.value.mem = '';
      filter.value.device_type = [];
    };
    const filterDevices = () => {
      queryRules.value = [
        filter.value.region.length && { field: 'dc.region', op: 'in', value: filter.value.region },
        filter.value.zone.length && { field: 'dc.zone', op: 'in', value: filter.value.zone },
        filter.value.device_group.length && {
          field: 'device_family',
          op: 'in',
          value: filter.value.device_group,
        },
        filter.value.device_type.length && { field: 'dc.device_type', op: 'in', value: filter.value.device_type },
        filter.value.cpu && { field: 'cpu_core', op: 'eq', value: filter.value.cpu },
        filter.value.mem && { field: 'memory', op: 'eq', value: filter.value.mem },
      ].filter(Boolean);

      page.value.start = 0;

      loadResources();
    };
    const handleDeviceTypeChange = () => {
      filter.value.cpu = '';
      filter.value.mem = '';
      deviceConfigDisabled.value = filter.value.device_type.length > 0;
    };
    const loadRestrict = async () => {
      const { cpu, mem } = await apiService.getRestrict();
      options.value.cpu = cpu || [];
      options.value.mem = mem || [];
    };
    onMounted(() => {
      loadRestrict();
    });

    const { CommonTable, getListData, isLoading } = useTable({
      tableOptions: {
        columns: [
          ...columns,
          {
            label: '操作',
            width: 120,
            showOverflowTooltip: false,
            render: ({ row }: { row: any }) => {
              return (
                <Button
                  text
                  theme='primary'
                  disabled={row.listenerNum > 0 || row.delete_protect || row.require_type === 6}
                  v-bk-tooltips={{ content: '滚服由BG统一提交预测', disabled: row.require_type !== 6 }}
                  onClick={() => {
                    routerAction.open({
                      path: '/business/resource-plan/add',
                      query: {
                        [GLOBAL_BIZS_KEY]: whereAmI.getBizsId(),
                        action: 'add',
                        payload: encodeURIComponent(
                          JSON.stringify({
                            region_id: row.region,
                            zone_id: row.zone,
                            cvm: {
                              device_type: row.device_type,
                            },
                          }),
                        ),
                      },
                    });
                  }}>
                  增加预测
                </Button>
              );
            },
          },
        ],
      },
      requestOption: {
        sortOption: {
          sort: 'capacity',
          order: 'DESC',
          legacy: false,
        },
      },
      scrConfig: () => {
        return {
          url: '/api/v1/woa/config/capacity/list_with_device_info',
          payload: {
            filter: {
              op: 'and',
              rules: [...queryRules.value],
            },
            page: page.value,
          },
          pageEnableCountKey: 'count',
          clearRules: true,
        };
      },
    });

    const cvmDevicetypeParams = computed(() => {
      const { region, zone, device_group, cpu, mem, disk } = filter.value;
      return {
        vendor: VendorEnum.ZIYAN,
        region,
        zone,
        device_family: device_group,
        cpu,
        mem,
        disk,
        disable: false,
      };
    });

    return () => (
      <div class={cssModule.page}>
        <GridFilterComp
          rules={[
            {
              title: t('地域'),
              content: (
                <AreaSelector
                  ref='areaSelector'
                  v-model={filter.value.region}
                  multiple
                  clearable
                  filterable
                  params={{ resourceType: 'QCLOUDCVM' }}></AreaSelector>
              ),
            },
            {
              title: t('园区'),
              content: (
                <ZoneSelector
                  ref='zoneSelector'
                  v-model={filter.value.zone}
                  separateCampus={false}
                  multiple
                  params={{
                    resourceType: 'QCLOUDCVM',
                    region: filter.value.region,
                  }}></ZoneSelector>
              ),
            },
            {
              title: t('实例族'),
              content: (
                <bk-select
                  v-model={filter.value.device_group}
                  multiple
                  clearable
                  collapse-tags
                  onChange={handleDeviceGroupChange}>
                  {options.value.device_groups.map((item) => (
                    <bk-option key={item} value={item} label={item}></bk-option>
                  ))}
                </bk-select>
              ),
            },
            {
              title: t('机型'),
              content: (
                <DevicetypeSelector
                  v-model={filter.value.device_type}
                  resourceType='cvm'
                  params={cvmDevicetypeParams.value}
                  multiple
                  disabled={deviceTypeDisabled.value}
                  onChange={handleDeviceTypeChange}
                />
              ),
            },
            {
              title: t('CPU(核)'),
              content: (
                <bk-select
                  v-model={filter.value.cpu}
                  clearable
                  disabled={deviceConfigDisabled.value}
                  filterable
                  onChange={handleDeviceConfigChange}>
                  {options.value.cpu.map((item) => (
                    <bk-option key={item} value={item} label={item}></bk-option>
                  ))}
                </bk-select>
              ),
            },
            {
              title: t('内存(G)'),
              content: (
                <bk-select
                  v-model={filter.value.mem}
                  clearable
                  disabled={deviceConfigDisabled.value}
                  filterable
                  onChange={handleDeviceConfigChange}>
                  {options.value.mem.map((item) => (
                    <bk-option key={item} value={item} label={item}></bk-option>
                  ))}
                </bk-select>
              ),
            },
          ]}
          onSearch={filterDevices}
          onReset={clearFilter}
          loading={isLoading.value}
          col={5}
          class={cssModule.filter}
        />
        <section class={cssModule.table}>
          <CommonTable style={{ height: 'calc(100% - 48px)' }} />
        </section>
      </div>
    );
  },
});
