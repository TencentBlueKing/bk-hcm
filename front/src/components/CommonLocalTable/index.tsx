import { PropType, computed, defineComponent, reactive, ref, watch } from 'vue';
// import components
import { Loading, SearchSelect, Table } from 'bkui-vue';
import Empty from '@/components/empty';
// import types
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import { CLB_SPECS_REVERSE_MAP } from '@/constants';
import { Column } from 'bkui-vue/lib/table/props';
import { getLocalFilterConditions } from '@/utils';
import './index.scss';
import { VendorReverseMap } from '@/common/constant';

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
  setup(props, { slots, expose }) {
    // 搜索相关
    const searchValue = ref();
    // 表格相关
    const tableRef = ref();
    const tableData = ref(props.tableData);
    const tableSort = ref<{ key: string; type: string }>({
      key: props.tableOptions.columns[0].field as string,
      type: 'asc',
    });
    const pagination = reactive({ limit: 10, count: props.tableData.length });
    const hasTopBar = computed(() => props.hasOperation && props.hasSearch);

    const clearSelection = () => {
      tableRef.value?.clearSelection();
    };

    // 监听 searchValue 的变化，根据过滤条件过滤得到 实际用于渲染的数据
    const renderTableData = computed(() => {
      const filterConditions: { [key]: any } = getLocalFilterConditions(searchValue.value, (rule) => {
        switch (rule.id) {
          // 负载均衡规格类型需要映射
          case 'SpecType':
            return CLB_SPECS_REVERSE_MAP[rule.values[0].id];
          case 'vendor':
            return VendorReverseMap[rule.values[0].id];
          default:
            return rule.values[0].id;
        }
      });

      const getCompareValue = (item: any, field: string) => {
        const config: { [key]: any } = {
          lb_vip: ['private_ipv4_addresses', 'public_ipv4_addresses'],
        };
        if (Object.hasOwn(item, field)) return item[field];
        if (Object.hasOwn(config, field)) {
          return config[field].reduce((acc: any[], cur: string) => {
            const value = Array.isArray(item[cur]) ? item[cur] : [item[cur]];
            acc.push(...value);
            return acc;
          }, []);
        }
        return [];
      };
      const { key, type } = tableSort.value;
      const resultData = (
        Object.keys(filterConditions).length
          ? props.tableData.filter((item) => {
              return Object.keys(filterConditions).every((key) => {
                const values = getCompareValue(item, key);
                if (Array.isArray(values)) return values.some((data) => filterConditions[key].includes(data));
                return filterConditions[key].includes(values);
              });
            })
          : [...props.tableData]
      ).sort((prev, next) => {
        if (type === 'asc') return String(prev[key]).localeCompare(String(next[key]));
        return String(next[key]).localeCompare(String(prev[key]));
      });
      // 更新分页器
      pagination.count = resultData.length;
      return resultData;
    });

    watch(
      () => props.tableData,
      (val) => {
        // 解决异步函数 tableData 数据返回不及时的问题
        tableData.value = val;
      },
      { deep: true },
    );

    const handleSortBy = ({ column, type }: { column: any; type: string }) => {
      tableSort.value = {
        key: column.field,
        type,
      };
    };

    expose({ clearSelection });

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
              v-model={searchValue.value}
              data={props.searchOptions.searchData}
              valueBehavior='need-key'
              {...(props.searchOptions.extra || {})}
            />
          )}
        </section>
        {/* 表格 */}
        <Loading class='loading-container' loading={props.loading}>
          <Table
            ref={tableRef}
            class='table-container'
            row-key={props.tableOptions.rowKey}
            data={renderTableData.value}
            columns={props.tableOptions.columns}
            pagination={pagination}
            show-overflow-tooltip
            onColumnSort={handleSortBy}
            {...(props.tableOptions.extra || {})}>
            {{
              empty: () => {
                if (props.loading || renderTableData.value?.length) return null;
                return <Empty />;
              },
            }}
          </Table>
        </Loading>
      </div>
    );
  },
});
