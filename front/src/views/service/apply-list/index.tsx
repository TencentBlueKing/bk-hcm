import { defineComponent, ref } from 'vue';
import './index.scss';
import { useTable } from '@/hooks/useTable/useTable';
import { APPLY_TYPES, searchData } from './constants';
import { Button, Tab } from 'bkui-vue';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useRoute, useRouter } from 'vue-router';

const { TabPanel } = Tab;
export default defineComponent({
  setup() {
    const applyType = ref('all');
    const { columns } = useColumns('myApply');
    const router = useRouter();
    const route = useRoute();
    const { CommonTable } = useTable({
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
                  // 这里应该用 vendorHandler
                  router.push({
                    path: '/service/my-apply/detail',
                    query: {
                      ...route.query,
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
      },
    });
    return () => (
      <div class={'apply-list-wrapper'}>
        <Tab type='unborder-card' v-model:active={applyType.value} class={'header-tab'}>
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
