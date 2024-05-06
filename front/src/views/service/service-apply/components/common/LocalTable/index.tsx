import { PropType, computed, defineComponent, ref, useAttrs } from 'vue';
// import components
import { SearchSelect, Table, TableIColumn } from 'bkui-vue';
import Empty from '@/components/empty';
// import constants
import { CLB_SPECS_REVERSE_MAP } from '@/constants';
// import types
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import './index.scss';

interface ITableSearchOption {
  filterable: boolean;
  data: Array<ISearchItem>;
}

export default defineComponent({
  name: 'LocalTable',
  props: {
    tableData: Array<any>,
    tableColumns: Array<TableIColumn>,
    // 搜索相关的配置项
    searchOption: {
      type: Object as PropType<ITableSearchOption>,
      default: (): ITableSearchOption => ({ filterable: true, data: [] }),
    },
  },
  setup(props) {
    // use hooks
    const attrs = useAttrs();
    const searchValue = ref(); // 搜索值

    /**
     * @returns 过滤条件
     */
    const getFilterConditions = () => {
      const filterConditions = {};
      searchValue.value?.forEach((rule: any) => {
        let ruleValue;
        switch (rule.id) {
          // 负载均衡规格类型需要映射
          case 'SpecType':
            ruleValue = CLB_SPECS_REVERSE_MAP[rule.values[0].id];
            break;
          default:
            ruleValue = rule.values[0].id;
            break;
        }
        if (filterConditions[rule.id]) {
          // 如果 filterConditions[rule.id] 已经存在，则合并为一个数组
          filterConditions[rule.id] = [...filterConditions[rule.id], ruleValue];
        } else {
          filterConditions[rule.id] = [ruleValue];
        }
      });
      return filterConditions;
    };

    // 监听 searchValue 的变化，根据过滤条件过滤得到 实际用于渲染的数据
    const renderTableData = computed(() => {
      const filterConditions = getFilterConditions();
      return props.tableData.filter((item) =>
        Object.keys(filterConditions).every((key) => filterConditions[key].includes(item[key])),
      );
    });

    return () => (
      <div class='local-table-container'>
        <section class='top-bar'>
          {props.searchOption.filterable && (
            <SearchSelect class='table-search-select' v-model={searchValue.value} data={props.searchOption.data} />
          )}
        </section>
        <Table
          class='local-table'
          data={renderTableData.value}
          columns={props.tableColumns}
          pagination
          showOverflowTooltip
          {...attrs}>
          {{ empty: () => <Empty /> }}
        </Table>
      </div>
    );
  },
});
