import { computed, defineComponent, ref } from 'vue';
import './index.scss';
import { Button, Tab } from 'bkui-vue';
import { BkTabPanel } from 'bkui-vue/lib/tab';
import { AccountLevelEnum, reviewData, searchData, tabs, secondaryReviewData, secondarySearchData } from './constants';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';

export default defineComponent({
  setup(props, ctx) {
    const accountLevel = ref(AccountLevelEnum.FirstLevel);
    const { columns: firstAccountColumns } = useColumns(AccountLevelEnum.FirstLevel);
    const { columns: secondaryAccountColumns } = useColumns(AccountLevelEnum.SecondLevel);

    const { CommonTable: FirstLevelTable } = useTable({
      tableOptions: {
        columns: firstAccountColumns,
        reviewData: reviewData,
      },
      searchOptions: {
        searchData,
      },
      requestOption: {},
    });

    const { CommonTable: SecondaryLevelTable } = useTable({
      tableOptions: {
        columns: secondaryAccountColumns,
        reviewData: secondaryReviewData,
      },
      searchOptions: {
        searchData: secondarySearchData,
      },
      requestOption: {},
    });

    const tabs = [
      {
        key: AccountLevelEnum.FirstLevel,
        label: '一级账号',
        _component: FirstLevelTable,
      },
      {
        key: AccountLevelEnum.SecondLevel,
        label: '二级账号',
        _component: SecondaryLevelTable,
      },
    ];

    return () => (
      <div>
        <Tab v-model:active={accountLevel.value} type='card-grid' class={'account-manage-wrapper'}>
          {tabs.map(({ key, label, _component }) => (
            <BkTabPanel key={key} label={label} name={key} renderDirective='if'>
              <_component>
                {{
                  operation: () => (
                    <Button theme='primary'>
                      {accountLevel.value === AccountLevelEnum.FirstLevel ? '录入一级账号' : '创建二级账号'}
                    </Button>
                  ),
                }}
              </_component>
            </BkTabPanel>
          ))}
        </Tab>
      </div>
    );
  },
});
