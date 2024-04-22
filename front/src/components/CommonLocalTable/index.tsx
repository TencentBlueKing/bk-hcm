import { PropType, computed, defineComponent, reactive, ref, watch } from 'vue';
// import components
import { Loading, SearchSelect, Table } from 'bkui-vue';
import Empty from '@/components/empty';
// import types
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import { Column } from 'bkui-vue/lib/table/props';
import './index.scss';

/**
 * 本地搜索. 本地分页 Table 组件
 */
export default defineComponent({
  name: 'CommonLocalTable',
  props: {
    loading: Boolean,
    // 是否显示操作栏
    hasOperation: {
      type: Boolean,
      default: true,
    },
    // 是否显示搜索栏
    hasSearch: {
      type: Boolean,
      default: true,
    },
    // 搜索栏配置项
    searchOptions: {
      type: Object as PropType<{
        searchData: Array<ISearchItem>;
        extra?: Object;
      }>,
    },
    // 表格配置项
    tableOptions: {
      type: Object as PropType<{
        rowKey: string | Function;
        columns: Array<Column>;
        extra?: Object;
      }>,
    },
    // 表格数据
    tableData: Array<any>,
  },
  setup(props, { slots }) {
    // 搜索相关
    const searchVal = ref();
    // 表格相关
    const tableData = ref(props.tableData);
    const pagination = reactive({ count: 0, limit: 10 });
    const hasTopBar = computed(() => props.hasOperation && props.hasSearch);

    // todo: 本地搜索, 表格内容过滤

    watch(
      () => props.tableData,
      (val) => {
        // 解决异步函数 tableData 数据返回不及时的问题
        tableData.value = val;
      },
      { deep: true },
    );

    return () => (
      <div class={['local-table-container', hasTopBar.value && 'has-top-bar']}>
        {/* top-bar */}
        <section class='top-bar'>
          {/* 操作栏 */}
          {props.hasOperation && <div class='operation-area'>{slots.operation?.()}</div>}
          {/* 搜索栏 */}
          {props.hasSearch && (
            <SearchSelect
              class='table-search-selector'
              v-model={searchVal.value}
              data={props.searchOptions.searchData}
              valueBehavior='need-key'
              {...(props.searchOptions.extra || {})}
            />
          )}
        </section>
        {/* 表格 */}
        <Loading class='loading-container' loading={props.loading}>
          <Table
            class='table-container'
            row-key={props.tableOptions.rowKey}
            data={tableData.value}
            columns={props.tableOptions.columns}
            pagination={pagination}
            show-overflow-tooltip
            {...(props.tableOptions.extra || {})}>
            {{
              empty: () => {
                if (props.loading) return null;
                return <Empty />;
              },
            }}
          </Table>
        </Loading>
      </div>
    );
  },
});
