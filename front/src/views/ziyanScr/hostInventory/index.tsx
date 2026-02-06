import { defineComponent, ref, onMounted, computed } from 'vue';
import { useTable } from '@/hooks/useTable/useTable';
import { Search } from 'bkui-vue/lib/icon';
import apiService from '@/api/scrApi';
import { Form } from 'bkui-vue';
import AreaSelector from '../hostApplication/components/AreaSelector';
import ZoneSelector from '../hostApplication/components/ZoneSelector';
import DevicetypeSelector from '@/views/ziyanScr/components/devicetype-selector/index.vue';
import { VendorEnum } from '@/common/constant';
import './index.scss';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';

const { FormItem } = Form;
export default defineComponent({
  name: 'AllhostInventoryManager',
  setup() {
    const { columns } = useColumns('hostInventor');
    const deviceGroups = ['标准型', '高IO型', '大数据型', '计算型'];
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

    const loadRestrict = async () => {
      const { cpu, mem } = await apiService.getRestrict();
      options.value.cpu = cpu || [];
      options.value.mem = mem || [];
    };
    onMounted(() => {
      loadRestrict();
    });

    const { CommonTable, getListData } = useTable({
      tableOptions: {
        columns: [...columns],
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
    return () => (
      <div class={'apply-list-container cvm-web-wrapper'}>
        <div class={'filter-container'}>
          <Form model={filter.value} formType='vertical' class={'scr-form-wrapper'}>
            <FormItem label='地域'>
              <AreaSelector
                ref='areaSelector'
                v-model={filter.value.region}
                multiple
                clearable
                filterable
                params={{ resourceType: 'QCLOUDCVM' }}></AreaSelector>
            </FormItem>
            <FormItem label='园区'>
              <ZoneSelector
                ref='zoneSelector'
                v-model={filter.value.zone}
                separateCampus={false}
                multiple
                params={{
                  resourceType: 'QCLOUDCVM',
                  region: filter.value.region,
                }}></ZoneSelector>
            </FormItem>
            <FormItem label='实例族'>
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
            </FormItem>
            <FormItem label='机型'>
              <DevicetypeSelector
                v-model={filter.value.device_type}
                resourceType='cvm'
                params={cvmDevicetypeParams.value}
                multiple
                disabled={deviceTypeDisabled.value}
                onChange={handleDeviceTypeChange}
              />
            </FormItem>
            <FormItem label='CPU(核)'>
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
            </FormItem>
            <FormItem label='内存(G)'>
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
            </FormItem>
          </Form>
          <div class='btn-container'>
            <bk-button icon='bk-icon-search' theme='primary' onClick={filterDevices}>
              <Search></Search>
              查询
            </bk-button>
            <bk-button icon='bk-icon-refresh' onClick={clearFilter}>
              重置
            </bk-button>
          </div>
        </div>
        <CommonTable class={'filter-common-table'}></CommonTable>
      </div>
    );
  },
});
