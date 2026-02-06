import { defineComponent, ref, computed, reactive, watchEffect } from 'vue';
import classes from '../../index.module.scss';
import http from '@/http';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { removeEmptyFields } from '@/utils/scr/remove-query-fields';
import useTableSelection from '@/hooks/use-table-selection';
import { Button, Form, Message, Table } from 'bkui-vue';
import { Column } from 'bkui-vue/lib/table/props';
import QcloudRegionSelector from '@/views/ziyanScr/components/qcloud-resource/region-selector.vue';
import QcloudZoneSelector from '@/views/ziyanScr/components/qcloud-resource/zone-selector.vue';
import DevicetypeSelector from '@/views/ziyanScr/components/devicetype-selector/index.vue';
import QcloudZoneValue from '@/views/ziyanScr/components/qcloud-resource/zone-value.vue';
import QcloudRegionValue from '@/views/ziyanScr/components/qcloud-resource/region-value.vue';
import { VendorEnum } from '@/common/constant';

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
    const DATA_ROW_KEY = 'row_key';

    const tableRef = ref(null);

    const { selections, resetSelections, handleSelectAll, handleSelectChange } = useTableSelection({
      isRowSelectable: isRowSelectEnable,
      rowKey: DATA_ROW_KEY,
    });

    const options = ref([
      { value: 'IDCPM', label: 'IDC_物理机' },
      { value: 'QCLOUDCVM', label: '腾讯云_CVM' },
    ]);

    const formModel = reactive({
      resource_type: '',
      spec: { bk_cloud_regions: [], bk_cloud_zones: [], device_type: [] },
    });
    watchEffect(() => {
      const { resource_type: resourceType, spec } = props.formModelData ?? {};
      const { region, zone, device_type: deviceType } = spec;
      Object.assign(formModel, {
        resource_type: resourceType,
        spec: {
          bk_cloud_regions: [region],
          bk_cloud_zones: [zone],
          device_type: [deviceType],
        },
      });
    });

    const tableColumns = ref<Column[]>([
      { type: 'selection', width: 30, minWidth: 30 },
      { label: '机型', field: 'device_type' },
      {
        label: '地域',
        field: 'bk_cloud_region',
        render: ({ data }: any) => <QcloudRegionValue value={data.bk_cloud_region} />,
      },
      {
        label: '园区',
        field: 'bk_cloud_zone',
        render: ({ data }: any) => <QcloudZoneValue value={data.bk_cloud_zone} />,
      },
      { label: '数量', field: 'amount' },
      {
        label: '匹配数量',
        width: 250,
        render: ({ row }: any) => {
          return <bk-input size='mini' type='number' min={1} max={500} v-model={row.replicas}></bk-input>;
        },
      },
    ]);
    const deviceList = ref([]);
    const isLoading = ref(false);
    const getDeviceList = async () => {
      isLoading.value = true;
      try {
        const res = await http.post(
          `/api/v1/woa/${getBusinessApiPath()}pool/findmany/recall/match/device`,
          removeEmptyFields({ ...formModel }),
        );
        const list = res.data?.info || [];
        deviceList.value = list.map((item: any, index: number) => ({
          ...item,
          [DATA_ROW_KEY]: `${item.device_type ?? ''}_${item.bk_cloud_zone ?? ''}_${index}`,
        }));
      } finally {
        isLoading.value = false;
        resetSelections();
        tableRef.value?.clearSelection();
      }
    };
    const handleReset = () => {
      Object.assign(formModel, {
        resource_type: props.formModelData.resource_type,
        spec: { bk_cloud_regions: [], bk_cloud_zones: [], device_type: [] },
      });

      getDeviceList();
    };

    const onRegionChange = () => {
      formModel.spec.bk_cloud_zones = [];
    };
    const onResourceTypeChange = () => {
      Object.assign(formModel, { spec: { bk_cloud_regions: [], bk_cloud_zones: [], device_type: [] } });
    };
    const onZoneChange = () => {
      formModel.spec.device_type = [];
    };

    const cvmDevicetypeParams = computed(() => {
      const { bk_cloud_regions: region, bk_cloud_zones: zone } = formModel.spec || {};
      return { vendor: VendorEnum.ZIYAN, region, zone, disable: false };
    });

    const submitSelectedDevices = async () => {
      const {
        suborder_id,
        spec: { image_id, os_type },
      } = props.formModelData;
      const spec = selections.value.map((device) => {
        const { device_type, bk_cloud_region, bk_cloud_zone, replicas } = device;
        return {
          device_type,
          bk_cloud_region,
          bk_cloud_zone,
          replicas,
          image_id,
          os_type,
        };
      });
      await http.post(
        `/api/v1/woa/${getBusinessApiPath()}task/commit/apply/pool/match`,
        { suborder_id, spec },
        { removeEmptyFields: true },
      );
      Message({ message: '匹配成功', theme: 'success' });
      props.handleClose();
    };
    return () => (
      <div class={classes['apply-list-container']}>
        <div class={classes['filter-container']}>
          <Form model={formModel} class={classes['scr-form-wrapper']}>
            <FormItem label='资源类型' property='resource_type'>
              <bk-select v-model={formModel.resource_type} onChange={onResourceTypeChange}>
                {options.value.map((opt) => (
                  <bk-option key={opt.value} value={opt.value} label={opt.label} />
                ))}
              </bk-select>
            </FormItem>
            <FormItem label='地域' property='bk_cloud_regions'>
              <QcloudRegionSelector v-model={formModel.spec.bk_cloud_regions} onChange={onRegionChange} />
            </FormItem>
            <FormItem label='园区' property='bk_cloud_zones'>
              <QcloudZoneSelector
                v-model={formModel.spec.bk_cloud_zones}
                region={formModel.spec.bk_cloud_regions}
                onChange={onZoneChange}
              />
            </FormItem>
            <FormItem label='机型' property='device_type'>
              <DevicetypeSelector
                v-model={formModel.spec.device_type}
                resourceType={formModel.resource_type === 'QCLOUDCVM' ? 'cvm' : 'idcpm'}
                params={cvmDevicetypeParams.value}
                multiple
              />
            </FormItem>
          </Form>
        </div>
        <Button theme={'primary'} class={'ml24 mr8'} loading={isLoading.value} onClick={getDeviceList}>
          查询
        </Button>
        <Button disabled={isLoading.value} onClick={handleReset}>
          重置
        </Button>
        <Button theme='success' disabled={selections.value.length === 0} onClick={submitSelectedDevices} class={'ml24'}>
          手工匹配资源
        </Button>

        <div class={classes['data-list-container']} v-bkloading={{ loading: isLoading.value }}>
          <Table
            ref={tableRef}
            data={deviceList.value}
            columns={tableColumns.value}
            rowKey={DATA_ROW_KEY}
            showOverflowTooltip={true}
            isRowSelectEnable={isRowSelectEnable}
            onSelectAll={handleSelectAll}
            maxHeight={'calc(100vh - 350px)'}
            onSelectionChange={handleSelectChange}>
            {{
              prepend: () =>
                deviceList.value.length ? (
                  <div class={classes['table-prepend']}>
                    已{selections.value.length === deviceList.value.length ? '全选' : '选择'}机型总类：
                    <em class={classes['selected-num']}>{selections.value.length}</em>， 待交付数：
                    <em class={classes['pending-num']}>{props.formModelData.pending_num}</em>
                  </div>
                ) : null,
            }}
          </Table>
        </div>
      </div>
    );
  },
});
