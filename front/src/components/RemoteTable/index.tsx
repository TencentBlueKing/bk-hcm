import { PropType, computed, defineComponent, onMounted, ref } from 'vue';
import { Table, Loading } from 'bkui-vue';
import Empty from '../empty';
import usePagination from '@/hooks/usePagination';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import http from '@/http';
import { QueryRuleOPEnum, RulesItem } from '@/typings';
import './index.scss';

type UrlType = string | (() => string);
type PayloadType = object | (() => object);
type RulesType = RulesItem[] | ((dataList?: any[]) => RulesItem[]);
interface RequestApi {
  // 请求url地址
  url: UrlType;
  // 请求载荷
  payload: PayloadType;
  // 请求过滤条件requestBody
  rules: RulesType;
  // 是否取消本次请求
  reject: (dataList: any[]) => boolean;
}

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineComponent({
  name: 'RemoteTable',
  props: {
    columnName: String,
    noSort: { type: Boolean, default: false },
    apis: Array as PropType<RequestApi[]>,
  },
  setup(props, { expose }) {
    const isLoading = ref(false);
    const dataList = ref([]);
    const { pagination, handlePageLimitChange, handlePageValueChange } = usePagination(() => getDataList());

    const { columns, settings } = useColumns(props.columnName);

    const renderColumns = computed(() => {
      if (props.noSort) {
        return columns.map((item: any) => ({ ...item, sort: false }));
      }
      return columns;
    });

    const buildApiMethod = async (fetchUrl: string, rules: RulesItem[], payload: PayloadType) => {
      const [detailsRes, countRes] = await Promise.all(
        [false, true].map((isCount) =>
          http.post(BK_HCM_AJAX_URL_PREFIX + fetchUrl, {
            ...(typeof payload === 'function' ? payload() : payload),
            filter: { op: QueryRuleOPEnum.AND, rules },
            page: {
              limit: isCount ? 0 : pagination.limit,
              start: isCount ? 0 : pagination.start,
              sort: isCount ? undefined : 'created_at',
              order: isCount ? undefined : 'DESC',
              count: isCount,
            },
          }),
        ),
      );
      dataList.value = detailsRes.data.details || [];
      pagination.count = countRes.data.count || 0;
    };

    const getDataList = async () => {
      isLoading.value = true;
      try {
        for (const api of props.apis) {
          const reject = typeof api.reject === 'function' ? api.reject(dataList.value) : false;
          const fetchUrl = typeof api.url === 'string' ? api.url : api.url();
          const rules = typeof api.rules === 'function' ? api.rules(dataList.value) : api.rules;
          const payload = typeof api.payload === 'function' ? api.payload() : api.payload;
          // 请求数据
          !reject && (await buildApiMethod(fetchUrl, rules || [], payload || {}));
        }
      } finally {
        isLoading.value = false;
      }

      return dataList.value;
    };

    expose({ dataList, getDataList });

    onMounted(() => {
      getDataList();
    });

    return () => (
      <Loading loading={isLoading.value} class='remote-table has-selection'>
        <Table
          class='table-container'
          data={dataList.value}
          rowKey='id'
          columns={renderColumns.value}
          settings={settings.value}
          pagination={pagination}
          remotePagination
          showOverflowTooltip
          onPageLimitChange={handlePageLimitChange}
          onPageValueChange={handlePageValueChange}>
          {{
            empty: () => {
              if (isLoading.value) return null;
              return <Empty />;
            },
          }}
        </Table>
      </Loading>
    );
  },
});
