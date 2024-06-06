import { defineComponent, ref } from 'vue';
import './index.scss';
import { Button, Tab } from 'bkui-vue';
import { BkTabPanel } from 'bkui-vue/lib/tab';
import { AccountLevelEnum, reviewData, searchData, secondaryReviewData, secondarySearchData } from './constants';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import CommonSideslider from '@/components/common-sideslider';
import FirstLevelAccountDetail from '../account-detail/first-level-account-detail';
import SecondLevelAccountDetail from '../account-detail/second-level-account-detail';
import { useRoute, useRouter } from 'vue-router';

export default defineComponent({
  setup() {
    const router = useRouter();
    const route = useRoute();

    const accountLevel = ref(AccountLevelEnum.FirstLevel);
    const { columns: firstAccountColumns } = useColumns(AccountLevelEnum.FirstLevel);
    const { columns: secondaryAccountColumns } = useColumns(AccountLevelEnum.SecondLevel);

    const isFirstLevelSideSliderShow = ref(false);
    const isSecondLevelSideSliderShow = ref(false);

    const curFirstLevelAccount = ref({});
    const curSecondLeveleAccount = ref({});

    const { CommonTable: FirstLevelTable } = useTable({
      tableOptions: {
        columns: [
          {
            label: '一级帐号名称',
            field: 'primaryAccountName',
            render: ({ data }: any) => (
              <Button
                text
                theme='primary'
                onClick={() => {
                  // SideSlider展示详情(可编辑)
                  curFirstLevelAccount.value = data;
                  isFirstLevelSideSliderShow.value = true;
                }}>
                {data.primaryAccountName}
              </Button>
            ),
          },
          ...firstAccountColumns,
        ],
        reviewData,
      },
      searchOptions: {
        searchData,
      },
      requestOption: {},
    });

    const { CommonTable: SecondaryLevelTable } = useTable({
      tableOptions: {
        columns: [
          {
            label: '二级帐号名称',
            field: 'secondaryAccountName',
            render: ({ data }: any) => (
              <Button
                text
                theme='primary'
                onClick={() => {
                  curSecondLeveleAccount.value = data;
                  isSecondLevelSideSliderShow.value = true;
                }}>
                {data.secondaryAccountName}
              </Button>
            ),
          },
          ...secondaryAccountColumns,
        ],
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
      <div class={'account-manage-wrapper'}>
        <div class={'header'}>
          <p class={'title'}>云账号管理</p>
        </div>
        <Tab v-model:active={accountLevel.value} type='card-grid' class={'account-table'}>
          {tabs.map(({ key, label, _component }) => (
            <BkTabPanel key={key} label={label} name={key} renderDirective='if'>
              <_component>
                {{
                  operation: () => (
                    <Button
                      theme='primary'
                      onClick={() => {
                        router.push({
                          path:
                            accountLevel.value === AccountLevelEnum.FirstLevel
                              ? '/bill/account-manage/first-account'
                              : '/bill/account-manage/second-account',
                          query: {
                            ...route.query,
                          },
                        });
                      }}>
                      {accountLevel.value === AccountLevelEnum.FirstLevel ? '录入一级账号' : '创建二级账号'}
                    </Button>
                  ),
                }}
              </_component>
            </BkTabPanel>
          ))}
        </Tab>

        {/* 一级账号详情及编辑 */}
        <CommonSideslider
          v-model:isShow={isFirstLevelSideSliderShow.value}
          width={640}
          title={'一级账号详情'}
          noFooter={true}>
          <FirstLevelAccountDetail detail={curFirstLevelAccount.value} />
        </CommonSideslider>

        {/* 二级账号详情及编辑 */}
        <CommonSideslider
          v-model:isShow={isSecondLevelSideSliderShow.value}
          width={640}
          title={'二级账号详情'}
          noFooter={true}>
          <SecondLevelAccountDetail detail={curSecondLeveleAccount.value} />
        </CommonSideslider>
      </div>
    );
  },
});
