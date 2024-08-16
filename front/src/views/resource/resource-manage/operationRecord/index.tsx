import { computed, defineComponent, reactive, ref, watch } from 'vue';
import { Button, OverflowTitle, Sideslider } from 'bkui-vue';
import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import './index.scss';
import { CloudType, QueryRuleOPEnum, RulesItem } from '@/typings';
import { useTable } from '@/hooks/useTable/useTable';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import useColumns from '../hooks/use-columns';
import { timeFormatter } from '@/common/util';
import { useRouter } from 'vue-router';
import { useAccountStore } from '@/store';
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';

export default defineComponent({
  setup() {
    const businessMapStore = useBusinessMapStore();
    const accountStore = useAccountStore();
    const { isResourcePage } = useWhereAmI();
    const router = useRouter();

    // 资源类型 tab 选项
    const resourceTypes = [
      { label: 'all', text: '全部' },
      { label: 'security_group', text: '安全组' },
      { label: 'load_balancer', text: '负载均衡' },
    ];
    // 当前选中的资源类型，默认为全部
    const activeResourceType = ref('all');

    // 表格
    const { columns, settings } = useColumns('operationRecord');
    const searchRule = reactive<RulesItem>({ field: 'res_type', op: QueryRuleOPEnum.EQ, value: 'all' });
    const tableColumns = [
      ...columns,
      {
        label: '操作',
        render: ({ data }: { data: any }) => (
          <div class='operation-cell'>
            <Button text theme='primary' onClick={() => showOperationDetail(data)}>
              查看详情
            </Button>
            {/* <Button text theme='primary'>
              再次购买
            </Button> */}
          </div>
        ),
      },
    ];
    const searchData = computed(() => {
      const base = [
        {
          name: '资源类型',
          id: 'res_type',
        },
        {
          name: '资源名称',
          id: 'res_name',
        },
        {
          name: '操作方式',
          id: 'action',
        },
        {
          name: '操作来源',
          id: 'source',
        },
        {
          name: '所属业务',
          id: 'bk_biz_id',
        },
        {
          name: '云账号',
          id: 'account_id',
        },
        {
          name: '操作人',
          id: 'operator',
        },
      ] as ISearchItem[];
      // 资源下, 不展示所属业务选项
      !isResourcePage && base.splice(4, 1);
      // 如果当前 tab 为负载均衡, 则展示任务类型选项(异步任务详情入口)
      activeResourceType.value === 'load_balancer' &&
        base.push({ name: '任务类型', id: 'detail.data.res_flow.flow_id', children: [{ name: '异步任务', id: '' }] });
      return base;
    });

    const { CommonTable } = useTable({
      searchOptions: {
        searchData: () => searchData.value,
        extra: {
          placeholder: '请输入',
        },
      },
      tableOptions: {
        columns: tableColumns,
        extra: {
          settings: settings.value,
        },
      },
      requestOption: {
        type: 'audits',
        sortOption: { sort: 'created_at', order: 'DESC' },
        filterOption: {
          rules: [searchRule],
          deleteOption: { field: 'res_type', flagValue: 'all' },
        },
      },
    });

    // 操作详情
    const isRecordDetailShow = ref(false);
    const currentDetailInfo = ref(null);
    const showOperationDetail = (listItem: any) => {
      if (
        ['load_balancer', 'url_rule', 'listener', 'url_rule_domain', 'target_group'].includes(listItem.res_type) &&
        listItem?.detail?.data?.res_flow?.flow_id
      ) {
        router.push({
          path: `/${isResourcePage ? 'resource' : 'business'}/record/detail`,
          query: {
            record_id: listItem.id,
            name: listItem.res_name,
            flow: listItem.detail.data.res_flow.flow_id,
            res_id: listItem.res_id,
            bizs: accountStore.bizs,
          },
        });
        return;
      }
      isRecordDetailShow.value = true;
      currentDetailInfo.value = listItem;
    };
    const detailInfoItemOptions = computed(() => [
      {
        label: '资源类型',
        field: 'res_type',
      },
      {
        label: '云厂商',
        field: 'vendor',
        renderValue: (cell: string) => {
          return CloudType[cell] || '--';
        },
      },
      {
        label: '云账号',
        field: 'account_id',
      },
      {
        label: '业务',
        field: 'bk_biz_id',
        renderValue: (cell: number) => {
          return businessMapStore.businessMap.get(cell) || '未分配';
        },
      },
      {
        label: '实例ID',
        field: 'res_id',
      },
      {
        label: '实例名称',
        field: 'res_name',
      },
      {
        label: '云资源ID',
        field: 'cloud_res_id',
      },
      {
        label: '操作方式',
        field: 'action',
      },
      {
        label: '操作人',
        field: 'operator',
      },
      {
        label: '操作时间',
        field: 'created_at',
        renderValue: (cell: string) => timeFormatter(cell),
      },
      {
        label: '请求ID',
        field: 'rid',
      },
      {
        label: '操作来源',
        field: 'source',
      },
    ]);

    watch(activeResourceType, (val) => {
      if (['load_balancer'].includes(val)) {
        searchRule.value = ['load_balancer', 'url_rule', 'listener', 'url_rule_domain', 'target_group'];
        searchRule.op = QueryRuleOPEnum.IN;
      } else {
        searchRule.value = val;
        searchRule.op = QueryRuleOPEnum.EQ;
      }
    });

    return () => (
      <div class={`operation-record-module${isResourcePage ? ' resource-apply' : ''}`}>
        <div class='common-card-wrap'>
          <CommonTable>
            {{
              operation: () => (
                <BkRadioGroup v-model={activeResourceType.value} type='capsule' class='resource-radio-group'>
                  {resourceTypes.map(({ label, text }) => (
                    <BkRadioButton label={label}>{text}</BkRadioButton>
                  ))}
                </BkRadioGroup>
              ),
            }}
          </CommonTable>
        </div>
        <Sideslider
          v-model:isShow={isRecordDetailShow.value}
          title='操作详情'
          width={670}
          class='record-detail-sideslider'>
          <div class='detail-info-container'>
            {detailInfoItemOptions.value.map(({ label, field, renderValue }) => (
              <div key={label} class='info-item'>
                <span class='item-label'>{label}</span>:
                <span class='item-value'>
                  <OverflowTitle type='tips' popoverOptions={{ theme: 'light' }}>
                    {renderValue ? renderValue(currentDetailInfo.value[field]) : currentDetailInfo.value[field]}
                  </OverflowTitle>
                </span>
              </div>
            ))}
          </div>
          <div class='detail-json-container'>
            <pre>
              <code>{JSON.stringify(currentDetailInfo.value.detail.data, null, 2)}</code>
            </pre>
          </div>
        </Sideslider>
      </div>
    );
  },
});
