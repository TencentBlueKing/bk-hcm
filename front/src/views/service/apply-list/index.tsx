import { computed, defineComponent, ref, watch } from 'vue';
import './index.scss';
import { useTable } from '@/hooks/useTable/useTable';
import { APPLY_TYPES, searchData } from './constants';
import { Button, Tab } from 'bkui-vue';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useRoute, useRouter } from 'vue-router';

const { TabPanel } = Tab;
export default defineComponent({
  setup() {
    const { columns } = useColumns('myApply');
    const router = useRouter();
    const route = useRoute();
    const { CommonTable, getListData } = useTable({
      searchOptions: {
        searchData,
      },
      tableOptions: {
        columns: [
          {
            label: '单号',
            field: 'sn',
            render: ({ data }: any) => (
              <Button
                text
                theme='primary'
                onClick={() => {
                  router.push({
                    path: '/service/my-apply/detail',
                    query: {
                      ...route.query,
                      id: data.id,
                      type: data.type,
                    },
                  });
                }}>
                {data.sn}
              </Button>
            ),
          },
          ...columns,
        ],
      },
      requestOption: {
        type: 'applications',
        immediate: false,
      },
    });

    const applyType = ref(route.query?.type || 'all');

    const computedRules = computed(() => {
      return APPLY_TYPES.find(({ name }) => name === applyType.value).rules;
    });

    const saveActiveType = (val: string) => {
      router.replace({ query: { type: val } });
    };

    watch(
      () => applyType.value,
      () => {
        getListData(applyType.value === 'all' ? [] : computedRules.value, 'applications', true);
      },
      { immediate: true },
    );

    return () => (
      <div class={'apply-list-wrapper'}>
        <Tab
          type='unborder-card'
          v-model:active={applyType.value}
          class={'header-tab'}
          onUpdate:active={saveActiveType}>
          {APPLY_TYPES.map(({ label, name }) => (
            <TabPanel name={name} label={label} />
          ))}
        </Tab>
        <div class={'table-wrapper'}>
          <CommonTable />
        </div>
      </div>
    );
  },
});
