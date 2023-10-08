import { SearchSelect, Loading, Table } from 'bkui-vue';
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import { PropType, defineComponent, reactive, ref } from 'vue';
import './index.scss';
import { Column } from 'bkui-vue/lib/table/props';

export default defineComponent({
  props: {
    searchData: {
      required: true,
      type: Array as PropType<Array<ISearchItem>>,
    },
    columns: {
      required: true,
      type: Array as PropType<Array<Column>>,
    },
    data: {
      required: true,
      type: Array as PropType<Array<Record<string, any>>>,
    },
  },
  setup(props, { slots }) {
    const pagination = reactive({
      start: 0,
      limit: 10,
      count: 100,
    });
    const searchVal = ref([]);
    const isLoading = ref(false);
    return () => (
      <>
        <div class={'felx-row'}>
          { slots.default?.() }
          <SearchSelect
            class='w500 common-search-selector'
            v-model={searchVal.value}
            data={props.searchData}
          />
        </div>
        <Loading loading={isLoading.value}>
          <Table
            data={props.data}
            columns={props.columns}
            pagination={pagination}
          />
        </Loading>
      </>
    );
  },
});
