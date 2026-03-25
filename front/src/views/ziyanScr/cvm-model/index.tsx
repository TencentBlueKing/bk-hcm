import { computed, defineComponent, ref, onMounted, reactive } from 'vue';
import { useTable } from '@/hooks/useTable/useTable';
import { Search } from 'bkui-vue/lib/icon';
import { Dialog, Form } from 'bkui-vue';
import apiService from '@/api/scrApi';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import AreaSelector from '../hostApplication/components/AreaSelector';
import ZoneSelector from '../hostApplication/components/ZoneSelector';
import CreateDevice from './CreateDevice/index';
import DevicetypeSelector from '@/views/ziyanScr/components/devicetype-selector/index.vue';
import { VendorEnum } from '@/common/constant';
import './index.scss';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
const { FormItem } = Form;
export default defineComponent({
  name: 'AllhostInventoryManager',
  setup() {
    const { columns } = useColumns('cvmModel');
    const { selections, handleSelectionChange } = useSelection();
    const deviceGroups = ['标准型', '高IO型', '大数据型', '计算型'];
    const filter = ref({
      region: [],
      zone: [],
      device_type: [],
      device_family: deviceGroups && [deviceGroups[0]],
      cpu_core: '',
      memory: '',
      disable: undefined as string | boolean | undefined,
    });
    const options = ref({
      device_families: deviceGroups,
      device_types: [],
      regions: [],
      zones: [],
      cpu: [],
      mem: [],
      enabled: [
        { label: '是', value: false }, // 可申请 = 是，对应 disable = false
        { label: '否', value: true }, // 可申请 = 否，对应 disable = true
      ],
    });
    const deviceConfigDisabled = ref(false);
    const deviceTypeDisabled = ref(false);
    const batchEditDialogVisible = ref(false);
    const createDeviceDialogState = reactive({ isShow: false, isHidden: false });
    const loadingState = reactive({
      update: false,
      create: false,
    });
    const batchEditForm = ref({
      disable: 0,
    });
    const queryRules = ref(
      [
        { field: 'vendor', op: 'eq', value: VendorEnum.ZIYAN },
        filter.value.region.length && { field: 'region', op: 'in', value: filter.value.region },
        filter.value.zone.length && { field: 'zone', op: 'in', value: filter.value.zone },
        filter.value.device_family.length && { field: 'device_family', op: 'in', value: filter.value.device_family },
        filter.value.device_type.length && { field: 'device_type', op: 'in', value: filter.value.device_type },
        filter.value.cpu_core && { field: 'cpu_core', op: 'eq', value: filter.value.cpu_core },
        filter.value.memory && { field: 'memory', op: 'eq', value: filter.value.memory },
        filter.value.disable !== undefined &&
          filter.value.disable !== '' && { field: 'disable', op: 'eq', value: filter.value.disable },
      ].filter(Boolean),
    );

    const { CommonTable, getListData, pagination } = useTable({
      tableOptions: {
        columns,
        extra: {
          onSelect: (selections: any) => {
            handleSelectionChange(selections, () => true, false);
          },
          onSelectAll: (selections: any) => {
            handleSelectionChange(selections, () => true, true);
          },
        },
      },
      requestOption: {
        sortOption: {
          legacy: false,
        },
      },
      scrConfig: () => {
        return {
          url: '/api/v1/woa/config/findmany/config/cvm/device',
          pageEnableCountKey: 'count',
          payload: {
            filter: {
              op: 'and',
              rules: [...queryRules.value],
            },
          },
          filter: { simpleConditions: true, requestId: 'devices' },
        };
      },
    });

    const loadResources = () => {
      pagination.start = 0;
      pagination.current = 1;
      getListData();
    };
    const handleDeviceConfigChange = () => {
      filter.value.device_type = [];
      const { cpu_core, memory } = filter.value;
      deviceTypeDisabled.value = Boolean(cpu_core || memory);
    };
    const clearFilter = () => {
      filter.value = {
        region: [],
        zone: [],
        device_type: [],
        device_family: deviceGroups && [deviceGroups[0]],
        cpu_core: '',
        memory: '',
        disable: undefined,
      };
      deviceConfigDisabled.value = false;
      deviceTypeDisabled.value = false;
      filterDevices();
    };
    const handleDeviceFamilyChange = () => {
      filter.value.cpu_core = '';
      filter.value.memory = '';
      filter.value.device_type = [];
    };
    const batchUpdates = () => {
      batchEditDialogVisible.value = true;
    };
    const createNewModel = () => {
      createDeviceDialogState.isHidden = false;
      createDeviceDialogState.isShow = true;
    };
    const triggerShow = (val: boolean) => {
      batchEditDialogVisible.value = val;
      batchEditForm.value = {
        disable: 0,
      };
    };
    const handleConfirm = async () => {
      try {
        loadingState.update = true;
        const { disable } = serializeBatchEditForm();
        const deviceTypes = selections.value.map((row) => ({ id: row.id, disable: disable ?? row.disable }));
        await apiService.updateCvmDeviceTypeConfigs({ device_types: deviceTypes });
        batchEditDialogVisible.value = false;
        selections.value = [];
        batchEditForm.value = {
          disable: 0,
        };
        loadResources();
      } finally {
        loadingState.update = false;
      }
    };
    const serializeBatchEditForm = () => {
      const { disable } = batchEditForm.value;
      return {
        disable: disable !== 0 ? Boolean(disable) : undefined,
      };
    };
    const filterDevices = () => {
      queryRules.value = [
        { field: 'vendor', op: 'eq', value: VendorEnum.ZIYAN },
        filter.value.region.length && { field: 'region', op: 'in', value: filter.value.region },
        filter.value.zone.length && { field: 'zone', op: 'in', value: filter.value.zone },
        filter.value.device_family.length && { field: 'device_family', op: 'in', value: filter.value.device_family },
        filter.value.device_type.length && { field: 'device_type', op: 'in', value: filter.value.device_type },
        filter.value.cpu_core && { field: 'cpu_core', op: 'eq', value: filter.value.cpu_core },
        filter.value.memory && { field: 'memory', op: 'eq', value: filter.value.memory },
        filter.value.disable !== undefined &&
          filter.value.disable !== '' && { field: 'disable', op: 'eq', value: filter.value.disable },
      ].filter(Boolean);

      loadResources();
    };
    const handleDeviceTypeChange = () => {
      filter.value.cpu_core = '';
      filter.value.memory = '';
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

    const cvmDevicetypeParams = computed(() => {
      const { region, zone, device_family, cpu_core, memory, disable } = filter.value;
      return {
        vendor: VendorEnum.ZIYAN,
        region,
        zone,
        device_family,
        cpu: cpu_core,
        mem: memory,
        disable: disable !== undefined && disable !== '' ? Boolean(disable) : undefined,
      };
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
                v-model={filter.value.device_family}
                multiple
                clearable
                collapse-tags
                onChange={handleDeviceFamilyChange}>
                {options.value.device_families.map((item) => (
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
                v-model={filter.value.cpu_core}
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
                v-model={filter.value.memory}
                clearable
                disabled={deviceConfigDisabled.value}
                filterable
                onChange={handleDeviceConfigChange}>
                {options.value.mem.map((item) => (
                  <bk-option key={item} value={item} label={item}></bk-option>
                ))}
              </bk-select>
            </FormItem>
            <FormItem label='可申请'>
              <bk-select
                v-model={filter.value.disable}
                clearable
                disabled={deviceConfigDisabled.value}
                filterable
                allowEmptyValues={[false]}
                onChange={handleDeviceConfigChange}>
                {options.value.enabled.map((item) => (
                  <bk-option key={item.value} value={item.value} label={item.label}></bk-option>
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
        <div class='btn-container oper-btn-pad'>
          <bk-button icon='bk-icon-refresh' disabled={!selections.value.length} onClick={batchUpdates}>
            批量更新
          </bk-button>
          <bk-button icon='bk-icon-refresh' onClick={createNewModel}>
            创建新机型
          </bk-button>
        </div>
        <CommonTable class={'filter-common-table'} />
        <Dialog
          class='common-dialog'
          close-icon={false}
          isShow={batchEditDialogVisible.value}
          isLoading={loadingState.update}
          title='批量更新'
          width={600}
          onConfirm={handleConfirm}
          onClosed={() => triggerShow(false)}>
          <bk-form>
            <bk-form-item label='可申请'>
              <bk-select
                v-model={batchEditForm.value.disable}
                style='width: 250px'
                clearable={false}
                allowEmptyValues={[false, 0]}>
                {[
                  {
                    value: 0,
                    label: '保持不变',
                  },
                  {
                    value: false, // 可申请 = 是，对应 disable = false
                    label: '是',
                  },
                  {
                    value: true, // 可申请 = 否，对应 disable = true
                    label: '否',
                  },
                ].map(({ label, value }) => {
                  return <bk-option key={value} label={label} value={value}></bk-option>;
                })}
              </bk-select>
            </bk-form-item>
          </bk-form>
        </Dialog>
        {!createDeviceDialogState.isHidden && (
          <CreateDevice
            v-model:isShow={createDeviceDialogState.isShow}
            onSubmit-success={loadResources}
            onHidden={() => (createDeviceDialogState.isHidden = true)}
          />
        )}
      </div>
    );
  },
});
