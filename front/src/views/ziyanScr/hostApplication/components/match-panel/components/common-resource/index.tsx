import { defineComponent, ref, computed } from 'vue';
import classes from '../../index.module.scss';
import useFormModel from '@/hooks/useFormModel';
import { Button, Form, Input, Message, Table } from 'bkui-vue';
import { Column } from 'bkui-vue/lib/table/props';
import http from '@/http';
import { timeFormatter } from '@/common/util';
import { INSTANCE_CHARGE_MAP, VendorEnum } from '@/common/constant';
import { removeEmptyFields } from '@/utils/scr/remove-query-fields';
import { useLegacyTableSettings } from '@/hooks/use-table-settings';
import useTableSelection from '@/hooks/use-table-selection';
import AreaSelector from '@/views/ziyanScr/hostApplication/components/AreaSelector';
import ZoneSelector from '@/views/ziyanScr/hostApplication/components/ZoneSelector';
import DiskTypeSelect from '../../../DiskTypeSelect';
import { useUserStore } from '@/store';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import DevicetypeSelector from '@/views/ziyanScr/components/devicetype-selector/index.vue';
import CvmImageSelector from '@/views/ziyanScr/components/ostype-selector/cvm-image-selector.vue';
import IdcpmOstypeSelector from '@/views/ziyanScr/components/ostype-selector/idcpm-ostype-selector.vue';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;
const { FormItem } = Form;

export default defineComponent({
  props: {
    formModelData: {
      type: Object,
    },
    handleClose: Function,
  },
  setup(props) {
    const { getBusinessApiPath } = useWhereAmI();

    const isRowSelectEnable = () => true;
    const DATA_ROW_KEY = 'asset_id';

    const tableRef = ref(null);

    const { selections, resetSelections, handleSelectAll, handleSelectChange } = useTableSelection({
      isRowSelectable: isRowSelectEnable,
      rowKey: DATA_ROW_KEY,
    });

    const userStore = useUserStore();
    const { formModel, forceClear } = useFormModel({
      resource_type: props.formModelData.resource_type,
      ips: '',
      spec: {
        region: [props.formModelData.spec.region],
        zone: [props.formModelData.spec.zone],
        device_type: [props.formModelData.spec.device_type],
        disk_type: [props.formModelData.spec.disk_type],
        os_type: '',
        instance_charge_type: props.formModelData.spec.charge_type,
      },
    });
    const options = ref([
      {
        value: 'IDCPM',
        label: 'IDC_物理机',
      },
      {
        value: 'QCLOUDCVM',
        label: '腾讯云_CVM',
      },
    ]);
    const deviceList = ref([]);
    const isLoading = ref(false);
    const getListData = async () => {
      isLoading.value = true;
      try {
        const { data } = await getDeviceList();
        deviceList.value = data.info || [];
      } finally {
        isLoading.value = false;
        resetSelections();
        tableRef.value?.clearSelection();
      }
    };
    const getDeviceList = () => {
      const { resource_type, spec } = formModel;

      return http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/${getBusinessApiPath()}task/findmany/apply/match/device`,
        removeEmptyFields({ resource_type, ips: ipArray.value, spec }),
      );
    };
    const tableColumns = ref<(Column & { defaultHidden?: boolean })[]>([
      { type: 'selection', fixed: true, width: 30, minWidth: 30 },
      {
        field: 'match_tag',
        label: '星标',
        width: 60,
        defaultHidden: true,
        render: ({ data }: any) => {
          if (data.matchTag) {
            return <i class='hcm-icon bkhcm-icon-collect' color='gold'></i>;
          }
          return <span>-</span>;
        },
      },
      {
        field: 'asset_id',
        label: '固资号',
        fixed: true,
        width: 130,
      },
      {
        field: 'ip',
        label: '内网 IP',
        width: 130,
      },
      {
        field: 'device_type',
        label: '机型',
        width: 130,
      },
      {
        field: 'instance_charge_type',
        label: '计费模式',
        width: 90,
        render: ({ cell }: any) => INSTANCE_CHARGE_MAP[cell] || '--',
      },
      {
        field: 'billing_start_time',
        label: '计费开始时间',
        width: 150,
        render: ({ data }: any) => (data.instance_charge_type ? timeFormatter(data.billing_start_time) : '--'),
      },
      {
        field: 'billing_expire_time',
        label: '计费结束时间',
        width: 150,
        render: ({ data }: any) => (data.instance_charge_type ? timeFormatter(data.billing_expire_time) : '--'),
      },
      {
        field: 'os_type',
        label: '操作系统',
        width: 230,
      },
      {
        field: 'equipment',
        label: '机架号',
        width: 80,
      },
      {
        field: 'zone',
        label: '园区',
        width: 100,
      },
      {
        field: 'module',
        label: '模块',
        width: 120,
      },
      {
        field: 'idc_unit',
        label: 'IDC 单元',
        width: 90,
      },
      {
        field: 'idc_logic_area',
        label: '逻辑区域',
        width: 110,
      },
      {
        field: 'input_time',
        label: '入库时间',
        defaultHidden: true,
        width: 120,
        render: ({ cell }: any) => timeFormatter(cell),
      },
      {
        field: 'match_score',
        label: '匹配得分',
        defaultHidden: true,
        width: 120,
      },
    ]);

    const { settings: tableSettings } = useLegacyTableSettings(tableColumns.value.slice(1));

    const chareTypeSet = computed(() => new Set(selections.value.map((item) => item.instance_charge_type)));
    const initialChareTypeSet = computed(() => new Set([props.formModelData.spec.charge_type]));
    const isInitialChareTypeEmpty = computed(() => !props.formModelData.spec.charge_type?.length);
    const isChareTypeHasEmpty = computed(() => [...chareTypeSet.value].some((item) => !item.length));
    const isChareTypeDifferent = computed(
      () =>
        chareTypeSet.value.size > 0 &&
        !isInitialChareTypeEmpty.value &&
        !isChareTypeHasEmpty.value &&
        (initialChareTypeSet.value.size !== chareTypeSet.value.size ||
          [...chareTypeSet.value].some((item) => !initialChareTypeSet.value.has(item))),
    );

    // 待交付数
    const pendingNum = computed(() => props.formModelData.pending_num);

    const matchButtonDisabled = computed(
      () => selections.value.length === 0 || selections.value.length > pendingNum.value,
    );

    const ipArray = computed(() => {
      const ipv4 = /^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$/;
      const ips = [];
      formModel.ips
        .split(/\r?\n/)
        .map((ip) => ip.trim())
        .filter((ip) => ip.length > 0)
        .forEach((item) => {
          if (ipv4.test(item)) {
            ips.push(item);
          }
        });
      return ips;
    });
    const onRegionChange = () => {
      formModel.spec.zone = [];
    };
    const onResourceTypeChange = () => {
      formModel.spec.region = [];
      formModel.spec.zone = [];
      formModel.spec.device_type = [];
    };

    const cvmDevicetypeParams = computed(() => {
      const { region, zone } = formModel.spec;
      return { vendor: VendorEnum.ZIYAN, region, zone, disable: false };
    });

    const loadResource = async () => {
      getListData();
    };
    const submitSelectedDevices = async () => {
      const { suborder_id } = props.formModelData;
      const device = selections.value.map((device) => {
        const { bk_host_id, asset_id, ip } = device;
        return {
          bk_host_id,
          asset_id,
          ip,
        };
      });
      await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/${getBusinessApiPath()}task/commit/apply/match`, {
        suborder_id,
        operator: userStore.username,
        device,
      });
      Message({
        message: '匹配成功',
        theme: 'success',
      });
      props.handleClose();
    };
    return () => (
      <div class={classes['apply-list-container']}>
        <div class={classes['filter-container']}>
          <Form model={formModel} class={classes['scr-form-wrapper']}>
            <FormItem label='资源类型'>
              <bk-select v-model={formModel.resource_type} onChange={onResourceTypeChange}>
                {options.value.map((opt) => (
                  <bk-option key={opt.value} value={opt.value} label={opt.label} />
                ))}
              </bk-select>
            </FormItem>
            <FormItem label='地域'>
              <AreaSelector
                ref='areaSelector'
                v-model={formModel.spec.region}
                params={{ resourceType: formModel.resource_type }}
                onChange={onRegionChange}
                {...{ multiple: true }}></AreaSelector>
            </FormItem>
            <FormItem label='园区'>
              <ZoneSelector
                ref='zoneSelector'
                v-model={formModel.spec.zone}
                params={{
                  resourceType: formModel.resource_type,
                  region: formModel.spec.region,
                }}
                {...{ multiple: true }}
              />
            </FormItem>
            <FormItem label='机型'>
              <DevicetypeSelector
                v-model={formModel.spec.device_type}
                resourceType={formModel.resource_type === 'QCLOUDCVM' ? 'cvm' : 'idcpm'}
                params={cvmDevicetypeParams.value}
                multiple
              />
            </FormItem>
            <FormItem label='操作系统'>
              {formModel.resource_type === 'QCLOUDCVM' ? (
                <CvmImageSelector v-model={formModel.spec.os_type} region={formModel.spec.region} idKey='image_name' />
              ) : (
                <IdcpmOstypeSelector v-model={formModel.spec.os_type} />
              )}
            </FormItem>
            <FormItem label='数据盘类型'>
              <DiskTypeSelect v-model={formModel.spec.disk_type} multiple />
            </FormItem>
            <FormItem label='内网IP'>
              <Input type='textarea' v-model={formModel.ips} />
            </FormItem>
            <FormItem label='计费模式'>
              <bk-select v-model={formModel.spec.instance_charge_type}>
                {Object.entries(INSTANCE_CHARGE_MAP).map(([value, label]) => (
                  <bk-option key={value} value={value} label={label} />
                ))}
              </bk-select>
            </FormItem>
          </Form>
        </div>
        <Button theme={'primary'} onClick={loadResource} class={'ml24 mr8'} loading={isLoading.value}>
          查询
        </Button>
        <Button
          onClick={() => {
            forceClear();
            getListData();
          }}>
          重置
        </Button>
        <Button
          theme='success'
          disabled={matchButtonDisabled.value}
          onClick={submitSelectedDevices}
          class={'ml24'}
          v-bk-tooltips={{
            content: '已选择数超过待交付数',
            disabled: !selections.value.length || !matchButtonDisabled.value,
          }}>
          手工匹配资源
        </Button>
        <div class={classes['data-list-container']} v-bkloading={{ loading: isLoading.value }}>
          {isInitialChareTypeEmpty.value && <bk-alert theme='error' title='原单据的计费模式为空' />}
          {!isInitialChareTypeEmpty.value && isChareTypeHasEmpty.value && (
            <bk-alert theme='error' title='选择匹配的资源，存在计费模式为空的情况' />
          )}
          {isChareTypeDifferent.value && (
            <bk-alert
              theme='warning'
              title={`所匹配的主机，与原单据的计费模式不同，请确认。原单据(${[...initialChareTypeSet.value]
                .map((type) => INSTANCE_CHARGE_MAP[type as string] || '--')
                .join('、')})，选择匹配(${[...chareTypeSet.value]
                .map((type) => INSTANCE_CHARGE_MAP[type as string] || '--')
                .join('、')})`}
            />
          )}

          <Table
            ref={tableRef}
            data={deviceList.value}
            columns={tableColumns.value}
            settings={tableSettings.value}
            rowKey={DATA_ROW_KEY}
            showOverflowTooltip={true}
            isRowSelectEnable={isRowSelectEnable}
            onSelectAll={handleSelectAll}
            maxHeight={'calc(100vh - 420px)'}
            onSelectionChange={handleSelectChange}>
            {{
              prepend: () =>
                deviceList.value.length ? (
                  <div class={classes['table-prepend']}>
                    待交付数：<em class={classes['pending-num']}>{props.formModelData.pending_num}</em>，已
                    {selections.value.length === deviceList.value.length ? '全选' : '选择'}：
                    <em class={classes['selected-num']}>{selections.value.length}</em>
                  </div>
                ) : null,
            }}
          </Table>
        </div>
      </div>
    );
  },
});
