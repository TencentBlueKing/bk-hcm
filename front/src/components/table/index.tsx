import { Loading, Table } from 'bkui-vue';
import type { Column } from 'bkui-vue/lib/table/props';
import { defineComponent, ref, computed } from 'vue';
import './index.scss';
import Empty from '@/components/empty';
export interface IProp {
  // 列表数据
  dataList: Array<any>;
  // table 配置项
  tableOptions: {
    // 表格字段
    columns: Array<Column> | (() => Array<Column>);
    // 其他 table 属性/自定义事件, 比如 settings, onSelectionChange...
    extra?: Object;
  };
}
export default defineComponent({
  props: {
    dataList: {
      type: Array,
      default: (): any => [],
    },
    tableOptions: {
      type: Object,
      default: (): any => {},
    },
  },
  emits: ['querylist', 'emptyform', 'handlePageLimitChange', 'handlePageValueChange', 'handleSort'],
  setup(props: IProp, { slots, emit }) {
    const isLoading = ref(false);
    const pagination = computed(() => {
      return { count: props.dataList.length, limit: 10 };
    });
    const data = computed(() => {
      isLoading.value = false;
      if (props.dataList) {
        return props.dataList;
      }
      return [];
    });
    // 点击查询按钮触发调父组件的方法请求接口
    const clickquerylist = () => {
      isLoading.value = true;
      emit('querylist');
    };
    // 点击清空按钮触发调父组件的方法请求接口，清空表单数据
    const clickemptyform = () => {
      emit('emptyform');
    };
    // 点击分页组件
    const handlePageValueChange = (value: any) => {
      pagination.value.current = value;
      isLoading.value = true;
      emit('handlePageLimitChange', pagination.value);
    };
    // 点击分页组件
    const handlePageLimitChange = (limit: any) => {
      pagination.value.limit = limit;
      isLoading.value = true;
      emit('handlePageValueChange', pagination.value);
    };
    // 点击排序
    const handleSort = () => {
      isLoading.value = true;
      emit('handleSort');
    };
    return () => (
      <div>
        <div>{slots.select?.()}</div>
        <bk-button class='ml10' theme='primary' onClick={clickquerylist} circle>
          <search class='f22' />
          <span>查询</span>
        </bk-button>
        <bk-button class='ml10' onClick={clickemptyform} circle>
          <search class='f22' />
          <span>清空</span>
        </bk-button>
        <Loading loading={isLoading.value} class='loading-table-container'>
          <Table
            class='table-container'
            data={data.value}
            rowKey='id'
            columns={props.tableOptions.columns}
            pagination={pagination.value}
            remotePagination
            showOverflowTooltip
            {...(props.tableOptions.extra || {})}
            onPageLimitChange={handlePageLimitChange}
            onPageValueChange={handlePageValueChange}
            onColumnSort={handleSort}
            onColumnFilter={() => {}}>
            {{
              empty: () => {
                if (isLoading.value) return null;
                return <Empty />;
              },
            }}
          </Table>
        </Loading>
      </div>
    );
  },
});
