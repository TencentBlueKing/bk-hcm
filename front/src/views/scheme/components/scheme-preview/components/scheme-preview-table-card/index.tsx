import { PropType, defineComponent, ref, watch } from 'vue';
import './index.scss';
import { Table, Tag, Loading, Button, Dialog, Form, Input } from 'bkui-vue';
import { AngleDown, AngleRight } from 'bkui-vue/lib/icon';
import VendorTcloud from '@/assets/image/vendor-tcloud.png';
// @ts-ignore
import AppSelect from '@blueking/app-select';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { useSchemeStore } from '@/store';
import { QueryRuleOPEnum } from '@/typings';
import { IServiceArea } from '@/typings/scheme';

const { FormItem } = Form;

export default defineComponent({
  props: {
    compositeScore: {
      type: Number,
      required: true,
      default: 0,
    },
    costScore: {
      type: Number,
      required: true,
      default: 0,
    },
    netScore: {
      type: Number,
      required: true,
      default: 0,
    },
    resultIdcIds: {
      type: Array as PropType<Array<string>>,
      required: true,
    },
    idx: {
      type: Number,
      required: true,
    },
    onViewDetail: {
      required: true,
      type: Function,
    },
  },
  setup(props) {
    const businessMapStore = useBusinessMapStore();
    const schemeStore = useSchemeStore();
    const columns = [
      {
        field: 'name',
        label: '部署点名称',
      },
      {
        field: 'vendor',
        label: '云厂商',
      },
      {
        field: 'region',
        label: '所在地',
      },
      {
        field: 'service_areas',
        label: '服务区域',
      },
      {
        field: 'ping',
        label: '平均延迟',
      },
    ];
    const tableData = ref([]);
    const isLoading = ref(false);
    const isExpanded = ref(false);
    const isDialogShow = ref(false);
    const idcServiceAreasMap = ref<Map<string, Array<IServiceArea>>>(new Map());

    watch(
      () => isExpanded.value,
      async () => {
        if (isExpanded.value && !tableData.value.length) {
          isLoading.value = true;
          const listIdcPromise = schemeStore.listIdc(
            {
              op: QueryRuleOPEnum.AND,
              rules: [
                {
                  field: 'id',
                  op: QueryRuleOPEnum.IN,
                  value: props.resultIdcIds,
                },
              ],
            },
            {
              start: 0,
              limit: props.resultIdcIds.length,
            },
          );
          const queryIdcServiceAreaPromise = schemeStore.queryIdcServiceArea(
            props.resultIdcIds,
            schemeStore.userDistribution,
          );
          const [listIdcRes, queryIdcServiceAreaRes] = await Promise.all([
            listIdcPromise,
            queryIdcServiceAreaPromise,
          ]);
          queryIdcServiceAreaRes.data.forEach((v) => {
            idcServiceAreasMap.value.set(v.idc_id, v.service_areas);
          });
          tableData.value = listIdcRes.data.map(v => ({
            name: v.name,
            vendor: v.vendor,
            region: v.region,
            service_areas: idcServiceAreasMap.value.get(v.id).reduce((acc, cur) => {
              acc += `${cur.country_name}, ${cur.province_name};`;
              return acc;
            }, ''),
            ping: 0,
          }));
          isLoading.value = false;
        }
      },
      {
        immediate: true,
      },
    );

    return () => (
      <div class={'scheme-preview-table-card-container'}>
        <div
          class={`scheme-preview-table-card-header ${
            isExpanded.value ? '' : 'scheme-preview-table-card-header-closed'
          }`}>
          {isExpanded.value ? (
            <AngleDown
              width={'40px'}
              height={'30px'}
              fill='#63656E'
              onClick={() => (isExpanded.value = !isExpanded.value)}
              class={'scheme-preview-table-card-header-expand-icon'}
            />
          ) : (
            <AngleRight
              width={'40px'}
              height={'30px'}
              fill='#63656E'
              onClick={() => (isExpanded.value = !isExpanded.value)}
              class={'scheme-preview-table-card-header-expand-icon'}
            />
          )}

          <p class={'scheme-preview-table-card-header-title'}>方案一</p>
          <Tag
            theme='success'
            radius='11px'
            class={'scheme-preview-table-card-header-tag'}>
            分布式部署
          </Tag>
          <img
            src={VendorTcloud}
            class={'scheme-preview-table-card-header-icon'}
          />
          <div class={'scheme-preview-table-card-header-score'}>
            <div class={'scheme-preview-table-card-header-score-item'}>
              综合评分：{' '}
              <span class={'score-value'}>{props.compositeScore}</span>
            </div>
            <div class={'scheme-preview-table-card-header-score-item'}>
              网络评分： <span class={'score-value'}>{props.netScore}</span>
            </div>
            <div class={'scheme-preview-table-card-header-score-item'}>
              方案成本： <span class={'score-value'}>$ {props.costScore}</span>
            </div>
          </div>
          <div class={'scheme-preview-table-card-header-operation'}>
            <Button class={'mr8'} onClick={props.onViewDetail}>查看详情</Button>
            <Button theme='primary' onClick={() => (isDialogShow.value = true)}>
              保存
            </Button>
          </div>
        </div>
        <div
          class={`scheme-preview-table-card-panel ${
            isExpanded.value ? '' : 'scheme-preview-table-card-panel-invisable'
          }`}>
          <Loading loading={isLoading.value}>
            <Table data={tableData.value} columns={columns} />
          </Loading>
        </div>

        <Dialog
          title='保存该方案'
          isShow={isDialogShow.value}
          onClosed={() => (isDialogShow.value = false)}
          onConfirm={() => (isDialogShow.value = false)}>
          <Form formType='vertical'>
            <FormItem label='方案名称' required>
              <Input />
            </FormItem>
            <FormItem label='标签'>
              <AppSelect data={businessMapStore.businessList} />
            </FormItem>
          </Form>
        </Dialog>
      </div>
    );
  },
});
