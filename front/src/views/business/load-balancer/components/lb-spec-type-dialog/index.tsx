import { computed, defineComponent, PropType, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { Dialog, SearchSelect, Table } from 'bkui-vue';
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import { CLB_SPECS } from '@/common/constant';
import { CLB_SPEC_TYPE_COLUMNS_MAP } from '@/constants/clb';
import { getSimpleConditionBySearchSelect } from '@/utils/search';
import type { SpecAvailability } from '@/api/load_balancers/apply-clb/types';
import { ModelPropertyColumn } from '@/model/typings';

import Empty from '@/components/empty';
import ModalFooter from '@/components/modal/modal-footer.vue';
import './index.scss';

interface IClbSlaSpecRow {
  SlaType: string;
  SlaName: string;
  MaxConn?: number;
  MaxCps?: number;
  MaxOutBits?: number;
  MaxInBits?: number;
  MaxQps?: number;
}

const formatCount = (value?: number) => value?.toLocaleString('en-US');

/**
 * 负载均衡实例规格选择对话框
 *
 * 规格中文名与性能指标由前端静态表维护。
 * 静态表只有一档带宽上限。DescribeSlaCapacity 的示例里 MaxInBits 与 MaxOutBits 相同，
 * 产品规格也只公布这一档上限，因此入口流量沿用同一数值。
 */
export default defineComponent({
  name: 'LbSpecTypeSelectDialog',
  props: {
    modelValue: Boolean,
    // 已选规格编码
    slaType: String,
    // 当前可用区下各规格的可用性（来自 load_balancers/resources/describe）。
    // 可选：静态表不含可用性，需要该约束的接入方才传；
    // 不传（默认空数组）时不按可用性过滤，不会收窄可选项
    specAvailabilitySet: {
      type: Array as PropType<Array<SpecAvailability>>,
      default: () => [],
    },
  },
  emits: ['update:modelValue', 'confirm', 'hidden'],
  setup(props, { emit }) {
    const { t } = useI18n();

    const isShow = computed({
      get() {
        return props.modelValue;
      },
      set(val: boolean) {
        emit('update:modelValue', val);
      },
    });

    const selectedSlaType = ref(props.slaType || '');
    const searchValue = ref<Array<any>>([]);

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
        meta: { display: { format: formatCount } },
        width: 180,
      },
      {
        id: 'MaxCps',
        name: t('新建连接数上限(个/秒)'),
        type: 'number',
        meta: { display: { format: formatCount } },
        width: 180,
      },
      {
        id: 'MaxOutBits',
        name: t('最大出口流量(Mbps)'),
        type: 'number',
        meta: { display: { format: formatCount } },
      },
      {
        id: 'MaxInBits',
        name: t('最大入口流量(Mbps)'),
        type: 'number',
        meta: { display: { format: formatCount } },
      },
      {
        id: 'MaxQps',
        name: t('最大qps(个/秒)'),
        type: 'number',
        meta: { display: { format: formatCount } },
      },
    ];

    // 简约型不提供购买；静态表未收录的 clb.c5.* 同样不展示
    const isVisibleSlaType = (slaType?: string) => slaType !== 'clb.c1.small' && !slaType?.startsWith('clb.c5.');

    const staticSlaList = computed<Array<IClbSlaSpecRow>>(() =>
      Object.entries(CLB_SPEC_TYPE_COLUMNS_MAP)
        .filter(([slaType]) => slaType !== 'shared')
        .map(([SlaType, spec]) => ({
          SlaType,
          SlaName: CLB_SPECS[SlaType] ?? SlaType,
          MaxConn: spec.connectionsPerMinute,
          MaxCps: spec.newConnectionsPerSecond,
          MaxQps: spec.queriesPerSecond,
          MaxOutBits: spec.bandwidthLimit,
          MaxInBits: spec.bandwidthLimit,
        })),
    );

    // 不传 specAvailabilitySet（为空）时视为不按可用性过滤
    const availableSlaTypes = computed(
      () =>
        new Set(
          props.specAvailabilitySet
            .filter(({ Availability }) => Availability !== 'Unavailable')
            .map(({ SpecType }) => SpecType),
        ),
    );

    const usableSlaList = computed(() =>
      staticSlaList.value.filter(
        ({ SlaType }) =>
          isVisibleSlaType(SlaType) && (!availableSlaTypes.value.size || availableSlaTypes.value.has(SlaType)),
      ),
    );

    const searchData = computed<Array<ISearchItem>>(() => [
      {
        id: 'sla_types',
        name: t('性能容量型规格'),
        children: usableSlaList.value.map(({ SlaType }) => ({ id: SlaType, name: SlaType })),
        multiple: true,
      },
    ]);

    const slaList = computed(() => {
      const { sla_types: slaTypes } = getSimpleConditionBySearchSelect(searchValue.value) ?? {};
      if (!slaTypes?.length) return usableSlaList.value;
      const selected = new Set(slaTypes);
      return usableSlaList.value.filter(({ SlaType }) => selected.has(SlaType));
    });

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

    const handleRowClick = (row: IClbSlaSpecRow) => {
      selectedSlaType.value = row.SlaType;
    };

    const handleConfirm = () => {
      const result = usableSlaList.value.find(({ SlaType }) => SlaType === selectedSlaType.value);
      if (!result) return;
      emit('confirm', {
        slaType: result.SlaType,
        slaName: result.SlaName,
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
              <Table
                data={slaList.value}
                rowKey='SlaType'
                height={360}
                show-overflow-tooltip
                onRowClick={(_: any, row: IClbSlaSpecRow) => handleRowClick(row)}>
                {{
                  default: () => (
                    <>
                      <bk-table-column width='40' min-width='40' show-overflow-tooltip={false}>
                        {{
                          default: ({ row }: { row: IClbSlaSpecRow }) => (
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
                    if (slaList.value.length) return null;
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
