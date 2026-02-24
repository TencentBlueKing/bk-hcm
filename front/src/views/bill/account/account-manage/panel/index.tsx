/**
 * @deprecated 此文件已废弃。一级账号和二级账号已拆分为独立页面文件：
 * - 一级账号：./root-account-list.tsx
 * - 二级账号：./main-account-list.tsx
 * 共用 accountLevel prop 区分的方式已弃用，每个账号类型有各自独立的页面组件。
 */
import { PropType, defineComponent, ref } from 'vue';
import { useRouter } from 'vue-router';

import { Button } from 'bkui-vue';
import FirstLevelAccountDetail from '../../account-detail/first-level-account-detail';
import SecondLevelAccountDetail from '../../account-detail/second-level-account-detail';
import CommonSideslider from '@/components/common-sideslider';

import { useI18n } from 'vue-i18n';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useTable } from '@/hooks/useTable/useTable';
import { AccountLevelEnum, searchData, secondarySearchData } from '../constants';
import { MENU_BILL_ROOT_ACCOUNT_CREATE, MENU_BILL_MAIN_ACCOUNT_CREATE } from '@/constants/menu-symbol';

export default defineComponent({
  props: {
    accountLevel: {
      type: String as PropType<AccountLevelEnum>,
      default: AccountLevelEnum.FirstLevel,
    },
  },
  setup(props) {
    const router = useRouter();
    const { t } = useI18n();

    const { columns } = useColumns(props.accountLevel);

    const isSideSliderShow = ref(false);
    const curAccount = ref<any>({});

    const { CommonTable } = useTable({
      tableOptions: {
        columns: [
          {
            label: props.accountLevel === AccountLevelEnum.FirstLevel ? '一级帐号名称' : '二级帐号名称',
            field: 'name',
            render: ({ data }: any) => (
              <Button
                text
                theme='primary'
                onClick={() => {
                  curAccount.value = data;
                  isSideSliderShow.value = true;
                }}>
                {data.name}
              </Button>
            ),
          },
          ...columns,
        ],
      },
      searchOptions: {
        searchData: props.accountLevel === AccountLevelEnum.FirstLevel ? searchData : secondarySearchData,
      },
      requestOption: {
        type: props.accountLevel === AccountLevelEnum.FirstLevel ? 'account/root_accounts' : 'account/main_accounts',
        sortOption: { sort: 'created_at', order: 'DESC' },
        dataPath: 'data.details',
      },
    });

    const handleCreate = () => {
      router.push({
        name:
          props.accountLevel === AccountLevelEnum.FirstLevel
            ? MENU_BILL_ROOT_ACCOUNT_CREATE
            : MENU_BILL_MAIN_ACCOUNT_CREATE,
      });
    };

    return () => (
      <>
        <CommonTable>
          {{
            operation: () => (
              // TODO: 操作权限 —— "录入/创建"按钮需要使用 hcm-auth 组件包裹以实现操作级权限控制
              // 一级账号对应 AUTH_FIND_ROOT_ACCOUNT 相关写权限，二级账号对应 AUTH_FIND_MAIN_ACCOUNT 相关写权限
              <Button theme='primary' onClick={handleCreate}>
                {props.accountLevel === AccountLevelEnum.FirstLevel ? t('录入一级账号') : t('创建二级账号')}
              </Button>
            ),
          }}
        </CommonTable>

        <CommonSideslider
          v-model:isShow={isSideSliderShow.value}
          width={640}
          title={props.accountLevel === AccountLevelEnum.FirstLevel ? t('一级账号详情') : t('二级账号详情')}
          noFooter={true}>
          {props.accountLevel === AccountLevelEnum.FirstLevel ? (
            <FirstLevelAccountDetail accountId={curAccount.value.id} />
          ) : (
            <SecondLevelAccountDetail accountId={curAccount.value.id} />
          )}
        </CommonSideslider>
      </>
    );
  },
});
