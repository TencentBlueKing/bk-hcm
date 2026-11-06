import { computed, defineComponent, PropType, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { Dialog, SearchSelect, Table } from 'bkui-vue';
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import { VendorEnum } from '@/common/constant';
import { getSimpleConditionBySearchSelect } from '@/utils/search';
import { ILoadBalancerSlaCapacity, useLoadBalancerClbStore } from '@/store/load-balancer/clb';
import type { SpecAvailability } from '@/api/load_balancers/apply-clb/types';
import { ModelPropertyColumn } from '@/model/typings';

import Empty from '@/components/empty';
import ModalFooter from '@/components/modal/modal-footer.vue';
import './index.scss';

/**
 * 负载均衡实例规格选择对话框
 *
 * 规格中文名与性能指标统一由后端 `load_balancers/sla/capacity/describe`
 * 返回，前端不再静态维护；表格按接口字段展示。
 */
export default defineComponent({
  name: 'LbSpecTypeSelectDialog',
  props: {
    modelValue: Boolean,
    // 已选规格编码
    slaType: String,
    // 云厂商标识，作为接口路径中的 vendor
    vendor: {
      type: String as PropType<VendorEnum>,
      default: VendorEnum.TCLOUD,
    },
    accountId: {
      type: String,
      default: '',
    },
    region: {
      type: String,
      default: '',
    },
    // 当前可用区下各规格的可用性（来自 load_balancers/resources/describe）。
    // 可选：规格参数接口不返回可用性，需要该约束的接入方才传；
    // 不传（默认空数组）时不按可用性过滤，不会收窄可选项
    specAvailabilitySet: {
      type: Array as PropType<Array<SpecAvailability>>,
      default: () => [],
    },
  },
  emits: ['update:modelValue', 'confirm', 'hidden'],
  setup(props, { emit }) {
    const { t } = useI18n();
    const loadBalancerClbStore = useLoadBalancerClbStore();

    const isShow = computed({
      get() {
        return props.modelValue;
      },
      set(val: boolean) {
        emit('update:modelValue', val);
      },
    });

    // 当前选中的规格编码
    const selectedSlaType = ref(props.slaType || '');
    // 接口返回的原始数据
    const slaList = ref<Array<ILoadBalancerSlaCapacity>>([]);
    // 弹窗内独立 loading，覆盖"参数不足不请求"与真实请求两种过程
    const loading = ref(false);
    // 首次加载的全量规格编码，作为搜索候选项，不随搜索结果收窄
    const allSlaTypes = ref<Array<string>>([]);

    // 搜索候选取首次全量结果，避免搜索后候选被自身结果收窄
    const searchData = computed<Array<ISearchItem>>(() => [
      {
        id: 'sla_types',
        name: t('性能容量型规格'),
        children: allSlaTypes.value.map((SlaType) => ({ id: SlaType, name: SlaType })),
        multiple: true,
      },
    ]);
    const searchValue = ref<Array<any>>([]);

    // 表格列按接口返回字段定义
    const columns: Array<ModelPropertyColumn> = [
      {
        id: 'SlaType',
        name: t('性能容量型规格'),
        type: 'string',
      },
      {
        id: 'SlaName',
        name: t('规格名称'),
        type: 'string',
        width: 100,
      },
      {
        id: 'MaxConn',
        name: t('并发连接数上限(个/秒)'),
        type: 'number',
        meta: { display: { format: (value: number) => value?.toLocaleString('en-US') } },
        width: 180,
      },
      {
        id: 'MaxCps',
        name: t('新建连接数上限(个/秒)'),
        type: 'number',
        meta: { display: { format: (value: number) => value?.toLocaleString('en-US') } },
        width: 180,
      },
      {
        id: 'MaxOutBits',
        name: t('最大出口流量(Mbps)'),
        type: 'number',
        meta: { display: { format: (value: number) => value?.toLocaleString('en-US') } },
      },
      {
        id: 'MaxInBits',
        name: t('最大入口流量(Mbps)'),
        type: 'number',
        meta: { display: { format: (value: number) => value?.toLocaleString('en-US') } },
      },
      {
        id: 'MaxQps',
        name: t('最大qps(个/秒)'),
        type: 'number',
        meta: { display: { format: (value: number) => value?.toLocaleString('en-US') } },
      },
    ];

    // 简约型与静态表无指标的 clb.c5.* 不提供购买
    const isVisibleSlaType = (slaType?: string) => slaType !== 'clb.c1.small' && !slaType?.startsWith('clb.c5.');

    // 当前可用区可用的规格编码集合。
    // 不传 specAvailabilitySet（为空）时视为不按可用性过滤，用于无需该约束的接入方
    const availableSlaTypes = computed(
      () =>
        new Set(
          props.specAvailabilitySet
            .filter(({ Availability }) => Availability !== 'Unavailable')
            .map(({ SpecType }) => SpecType),
        ),
    );

    // 规格参数接口不返回可用性，需要时用 resources/describe 的结果做交集，避免选到当前不可用的规格。
    // 未传 specAvailabilitySet 的接入方（空集）不做该过滤，行为与不接入前一致
    const filterUsable = (list: Array<ILoadBalancerSlaCapacity>) =>
      list.filter(
        ({ SlaType }) =>
          isVisibleSlaType(SlaType) && (!availableSlaTypes.value.size || availableSlaTypes.value.has(SlaType)),
      );

    // 搜索时把规格编码作为 sla_types 传给后端过滤，接口侧完成筛选
    const getSlaCapacities = async (slaTypes?: string[]) => {
      const { vendor, accountId, region } = props;
      slaList.value = [];
      if (!vendor || !accountId || !region) return;
      loading.value = true;
      try {
        const list = await loadBalancerClbStore.getLoadBalancerSlaCapacities(vendor, {
          account_id: accountId,
          region,
          sla_types: slaTypes,
        });
        slaList.value = filterUsable(list);
        // 无搜索条件的那次拿到的是全量，用它作为搜索候选
        if (!slaTypes?.length) allSlaTypes.value = slaList.value.map(({ SlaType }) => SlaType);
      } catch (error) {
        // 不落静态兜底：接口失败即空列表，避免展示可能过期的参数
        slaList.value = [];
      } finally {
        loading.value = false;
      }
    };

    watch(
      () => [props.modelValue, props.vendor, props.accountId, props.region],
      () => {
        if (!props.modelValue) return;
        getSlaCapacities();
      },
      { immediate: true },
    );

    // 搜索条件变化 -> 重新请求；清空搜索时取全量
    watch(searchValue, (val) => {
      if (!props.modelValue) return;
      const { sla_types } = getSimpleConditionBySearchSelect(val) ?? {};
      getSlaCapacities(sla_types);
    });

    // 关闭后清空搜索条件，避免下次打开沿用上一次的过滤
    watch(
      () => props.modelValue,
      (visible) => {
        if (!visible) searchValue.value = [];
      },
    );

    watch(
      () => props.slaType,
      (val) => {
        selectedSlaType.value = val || '';
      },
    );

    const handleRowClick = (row: ILoadBalancerSlaCapacity) => {
      selectedSlaType.value = row.SlaType;
    };

    const handleConfirm = () => {
      const result = slaList.value.find(({ SlaType }) => SlaType === selectedSlaType.value);
      if (!result) return;
      emit('confirm', {
        slaType: result.SlaType,
        slaName: result.SlaName,
        // 出口流量上限即公网带宽上限
        bandwidthLimit: result.MaxOutBits,
      });
      isShow.value = false;
    };

    const handleClosed = () => {
      isShow.value = false;
    };

    return () => (
      <Dialog v-model:is-show={isShow.value} title={t('选择实例规格')} width='60vw' onHidden={() => emit('hidden')}>
        {{
          default: () => (
            <>
              <SearchSelect
                class='lb-spec-type-search'
                v-model={searchValue.value}
                data={searchData.value}
                valueBehavior='need-key'
              />
              {/* height 直接交给 Table：loading 遮罩以 Table 为定位宿主，随其高度撑满 */}
              <Table
                v-bkloading={{ loading: loading.value }}
                data={slaList.value}
                rowKey='SlaType'
                height={360}
                show-overflow-tooltip
                onRowClick={(_: any, row: ILoadBalancerSlaCapacity) => handleRowClick(row)}>
                {{
                  default: () => (
                    <>
                      <bk-table-column width='40' min-width='40' show-overflow-tooltip={false}>
                        {{
                          default: ({ row }: { row: ILoadBalancerSlaCapacity }) => (
                            <bk-radio v-model={selectedSlaType.value} label={row.SlaType}>
                              　
                            </bk-radio>
                          ),
                        }}
                      </bk-table-column>
                      {columns.map((column) => (
                        <bk-table-column key={column.id} prop={column.id} label={column.name} width={column.width}>
                          {{
                            default: ({ row }: { row: Record<string, any> }) => (
                              <display-value property={column} value={row[column.id]} display={column.meta?.display} />
                            ),
                          }}
                        </bk-table-column>
                      ))}
                    </>
                  ),
                  empty: () => {
                    if (loading.value || slaList.value.length) return null;
                    return <Empty />;
                  },
                }}
              </Table>
            </>
          ),
          footer: () => (
            <ModalFooter disabled={!selectedSlaType.value} onConfirm={handleConfirm} onClosed={handleClosed} />
          ),
        }}
      </Dialog>
    );
  },
});
