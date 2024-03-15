import { Loading, SearchSelect, Table } from 'bkui-vue';
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import { Column } from 'bkui-vue/lib/table/props';
import { defineComponent, reactive, ref } from 'vue';
import './index.scss';

export interface IProp {
  data: Array<any>;
  columns: Array<Column>;
  searchData: Array<ISearchItem>;
}

export const useLocalTable = (props: IProp) => {
  const CommonLocalTable = defineComponent({
    setup(_props, { slots }) {
      const pagination = reactive({
        start: 0,
        limit: 10,
        count: 100,
      });
      const searchVal = ref([]);
      const isLoading = ref(false);
      return () => (
        <div class='local-table-container'>
          <section class='operation-wrap'>
            {slots.tab?.()}
            <SearchSelect class='common-search-selector w400' v-model={searchVal.value} data={props.searchData} />
          </section>
          <Loading loading={isLoading.value}>
            <Table data={props.data} columns={props.columns} pagination={pagination} />
          </Loading>
        </div>
      );
    },
  });
  return {
    CommonLocalTable,
  };
};
