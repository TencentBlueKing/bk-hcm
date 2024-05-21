import { PropType, defineComponent, onMounted, onUnmounted, ref, watch } from 'vue';
import { Radio } from 'bkui-vue';
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import CommonDialog from '@/components/common-dialog';
import CommonLocalTable from '@/components/CommonLocalTable';
import { useI18n } from 'vue-i18n';
import bus from '@/common/bus';
import { CLB_SPECS } from '@/common/constant';
import { SpecAvailability } from '@/api/load_balancers/apply-clb/types';
import { CLBSpecType } from '@/typings';
import { CLB_SPEC_TYPE_COLUMNS_MAP } from '@/constants';
import './index.scss';

/**
 * 负载均衡规格类型选择对话框
 */
export default defineComponent({
  name: 'LbSpecTypeSelectDialog',
  props: { modelValue: String as PropType<CLBSpecType> },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    const { t } = useI18n();

    const isShow = ref(false);

    // 搜索条件
    const searchData: Array<ISearchItem> = [
      { id: 'SpecType', name: '规格类型' },
      { id: 'connectionsPerMinute', name: '每分钟并发连接数' },
      { id: 'newConnectionsPerSecond', name: '每秒新建连接数' },
      { id: 'queriesPerSecond', name: '每秒查询数' },
      { id: 'bandwidthLimit', name: '带宽上限' },
    ];

    // 当前选中的规格类型
    const selectedLbSpecType = ref('');

    // 表格字段
    const columns = [
      {
        label: t('型号'),
        field: 'SpecType',
        render: ({ data }: any) => {
          return (
            <Radio v-model={selectedLbSpecType.value} label={data.SpecType}>
              <span class='font-small'>{CLB_SPECS[data.SpecType]}</span>
            </Radio>
          );
        },
      },
      {
        label: t('每分钟并发连接数（个）'),
        field: 'connectionsPerMinute',
      },
      {
        label: t('每秒新建连接数（个）'),
        field: 'newConnectionsPerSecond',
      },
      {
        label: t('每秒查询数（个）'),
        field: 'queriesPerSecond',
      },
      {
        label: t('带宽上限（Mbps）'),
        field: 'bandwidthLimit',
      },
    ];
    // 表格数据
    const tableData = ref<Array<SpecAvailability>>([]);

    // click-handler - 点击表格row触发的钩子
    const handleRowClick = (row: SpecAvailability) => {
      selectedLbSpecType.value = row.SpecType;
    };

    // submit-handler - 选择机型
    const handleSelectClbSpecType = () => {
      emit('update:modelValue', selectedLbSpecType.value ? selectedLbSpecType.value : 'shared');
    };

    watch(
      () => props.modelValue,
      (val) => {
        if (val === 'shared') {
          selectedLbSpecType.value = '';
        }
      },
    );

    onMounted(() => {
      // 显示弹框
      bus.$on('showLbSpecTypeSelectDialog', () => (isShow.value = true));
      bus.$on(
        'updateSpecAvailabilitySet',
        (data: any[]) =>
          (tableData.value = data
            .map((item) => {
              const { connectionsPerMinute, newConnectionsPerSecond, queriesPerSecond, bandwidthLimit } =
                CLB_SPEC_TYPE_COLUMNS_MAP[item.SpecType];
              return { ...item, connectionsPerMinute, newConnectionsPerSecond, queriesPerSecond, bandwidthLimit };
            })
            .sort((a, b) => parseInt(a.bandwidthLimit, 10) - parseInt(b.bandwidthLimit, 10))),
      );
    });

    onUnmounted(() => {
      bus.$off('showLbSpecTypeSelectDialog');
      bus.$off('updateSpecAvailabilitySet');
    });

    return () => (
      <CommonDialog
        v-model:isShow={isShow.value}
        class='lb-spec-type-select-dialog'
        title='选择实例规格'
        width='60vw'
        onHandleConfirm={handleSelectClbSpecType}>
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
      </CommonDialog>
    );
  },
});
