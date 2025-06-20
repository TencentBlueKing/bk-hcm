import { computed, defineComponent, PropType, ref, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import { SpecAvailability } from '@/api/load_balancers/apply-clb/types';
import { CLB_SPECS } from '@/common/constant';
import { CLB_SPEC_TYPE_COLUMNS_MAP } from '@/constants';

import CommonLocalTable from '@/components/CommonLocalTable';

/**
 * 负载均衡规格类型选择对话框
 */
export default defineComponent({
  name: 'LbSpecTypeSelectDialog',
  props: {
    modelValue: Boolean,
    slaType: String,
    specAvailabilitySet: Array as PropType<Array<SpecAvailability>>,
  },
  emits: ['update:modelValue', 'confirm', 'hidden'],
  setup(props, { emit }) {
    const { t } = useI18n();

    const isShow = computed({
      get() {
        return props.modelValue;
      },
      set(val) {
        emit('update:modelValue', val);
      },
    });

    // 搜索条件
    const searchData: Array<ISearchItem> = [
      { id: 'SpecType', name: '规格类型' },
      { id: 'connectionsPerMinute', name: '每分钟并发连接数' },
      { id: 'newConnectionsPerSecond', name: '每秒新建连接数' },
      { id: 'queriesPerSecond', name: '每秒查询数' },
      { id: 'bandwidthLimit', name: '带宽上限' },
    ];

    // 当前选中的规格类型
    const selectedLbSpecType = ref(props.slaType || '');

    // 表格字段
    const columns = [
      {
        label: '',
        width: 40,
        minWidth: 40,
        showOverflowTooltip: false,
        render: ({ row }: any) => (
          <bk-radio
            v-model={selectedLbSpecType.value}
            label={row.SpecType}
            disabled={row.Availability === 'Unavailable'}>
            　
          </bk-radio>
        ),
      },
      {
        label: t('型号'),
        field: 'SpecType',
        render: ({ data }: any) => CLB_SPECS[data.SpecType],
      },
      {
        label: t('每分钟并发连接数（个）'),
        field: 'connectionsPerMinute',
        render: ({ cell }: { cell: number }) => cell.toLocaleString('en-US'),
      },
      {
        label: t('每秒新建连接数（个）'),
        field: 'newConnectionsPerSecond',
        render: ({ cell }: { cell: number }) => cell.toLocaleString('en-US'),
      },
      {
        label: t('每秒查询数（个）'),
        field: 'queriesPerSecond',
        render: ({ cell }: { cell: number }) => cell.toLocaleString('en-US'),
      },
      {
        label: t('带宽上限（Mbps）'),
        field: 'bandwidthLimit',
        render: ({ cell }: { cell: number }) => cell.toLocaleString('en-US'),
      },
      {
        label: t('可用性'),
        field: 'Availability',
        render: ({ cell }: { cell: 'Available' | 'Unavailable' }) => {
          return <bk-tag theme={cell === 'Available' ? 'success' : 'danger'}>{cell}</bk-tag>;
        },
      },
    ];
    // 表格数据
    const tableData = ref<Array<SpecAvailability>>([]);

    // click-handler - 点击表格row触发的钩子
    const handleRowClick = (row: SpecAvailability) => {
      if (row.Availability === 'Unavailable') return;
      selectedLbSpecType.value = row.SpecType;
    };

    // submit-handler - 选择机型
    const handleSelectClbSpecType = () => {
      emit('confirm', {
        slaType: selectedLbSpecType.value ? selectedLbSpecType.value : 'shared',
        bandwidthLimit: CLB_SPEC_TYPE_COLUMNS_MAP[selectedLbSpecType.value].bandwidthLimit,
      });
    };

    watchEffect(() => {
      tableData.value = props.specAvailabilitySet
        // shared不展示
        .filter((item) => item.SpecType !== 'shared')
        .map((item) => {
          const { connectionsPerMinute, newConnectionsPerSecond, queriesPerSecond, bandwidthLimit } =
            CLB_SPEC_TYPE_COLUMNS_MAP[item.SpecType];
          return { ...item, connectionsPerMinute, newConnectionsPerSecond, queriesPerSecond, bandwidthLimit };
        })
        .sort((a, b) => a.bandwidthLimit - b.bandwidthLimit);
    });

    return () => (
      <bk-dialog
        v-model:isShow={isShow.value}
        title='选择实例规格'
        width='60vw'
        onConfirm={handleSelectClbSpecType}
        onHidden={() => emit('hidden')}>
        <CommonLocalTable
          searchOptions={{ searchData }}
          tableOptions={{
            rowKey: 'SpecType',
            columns,
            extra: {
              onRowClick: (_: any, row: any) => handleRowClick(row),
            },
          }}
          tableData={tableData.value}
        />
      </bk-dialog>
    );
  },
});
