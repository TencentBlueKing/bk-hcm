import { ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { Button, Tag } from 'bkui-vue';
// import stores
import { useBusinessMapStore } from '@/store/useBusinessMap';
// import hooks
import { useI18n } from 'vue-i18n';
import { useTable } from '@/hooks/useTable/useTable';
// import utils
import { timeFormatter } from '@/common/util';
import bus from '@/common/bus';
// import constants
import { QueryRuleOPEnum, RulesItem } from '@/typings';

export default () => {
  // use hooks
  const route = useRoute();
  const { t } = useI18n();
  // use stores
  const businessMapStore = useBusinessMapStore();
  // define data
  const columns = ref([
    {
      label: t('账号名称'),
      field: 'name',
      render: ({ data }: any) => {
        return (
          <>
            {data?.name}
            {data?.account_type !== '' && (
              <Tag theme={data?.account_type === 'current_account' ? 'info' : 'success'} class='users-list-bk-tag'>
                {data?.account_type === 'current_account' ? t('当前账号') : t('主账号')}
              </Tag>
            )}
          </>
        );
      },
    },
    {
      label: t('账号 ID'),
      field: 'id',
    },
    {
      label: t('所属业务'),
      field: 'bk_biz_ids',
      render: ({ data }: any) =>
        data?.bk_biz_ids.length > 0
          ? data?.bk_biz_ids
              .map((bk_biz_id: number) => {
                return businessMapStore.getNameFromBusinessMap(bk_biz_id);
              })
              ?.join(',')
          : '--',
    },
    {
      label: t('备注'),
      field: 'memo',
      render: ({ cell }: any) => cell || '--',
    },
    {
      label: t('负责人'),
      field: 'managers',
      render: ({ data }: any) => data?.managers?.join(',') || '--',
    },
    {
      label: t('更新人'),
      field: 'reviser',
    },
    {
      label: t('更新时间'),
      field: 'updated_at',
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
    {
      label: t('操作'),
      field: 'operation',
      render: ({ data }: any) => (
        <Button text theme='primary' onClick={() => bus.$emit('handleModifyAccount', data)}>
          编辑
        </Button>
      ),
    },
  ]);
  const filterRules = ref<RulesItem>({ op: QueryRuleOPEnum.EQ, field: 'account_id', value: route.query.accountId }); // 初始化过滤条件
  // use hooks
  const { CommonTable, getListData } = useTable({
    searchOptions: {
      searchData: [{ name: '账号 ID', id: 'id' }],
    },
    tableOptions: {
      columns: columns.value,
    },
    requestOption: {
      type: 'sub_accounts',
      filterOption: {
        rules: [filterRules.value],
      },
    },
  });
  watch(
    () => route.query.accountId as string,
    (val) => {
      // 更新筛选条件, 触发请求
      filterRules.value.value = val;
    },
  );
  return { CommonTable, getListData };
};
