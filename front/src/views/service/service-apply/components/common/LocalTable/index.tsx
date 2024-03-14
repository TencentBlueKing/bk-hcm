import { PropType, defineComponent, ref, useAttrs } from 'vue';
// import components
import { SearchSelect, Table, TableIColumn } from 'bkui-vue';
import Empty from '@/components/empty';
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
    return () => (
      <div class='local-table-container'>
        <section class='top-bar'>
          {props.searchOption.filterable && (
            <SearchSelect class='table-search-select' v-model={searchValue.value} data={props.searchOption.data} />
          )}
        </section>
        <Table
          class='local-table'
          data={props.tableData}
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
