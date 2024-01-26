import { computed, defineComponent, reactive, ref, watch } from 'vue';
import { Button, OverflowTitle, Sideslider } from 'bkui-vue';
import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import './index.scss';
import { CloudType, QueryRuleOPEnum, RulesItem } from '@/typings';
import { useTable } from '@/hooks/useTable/useTable';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import useColumns from '../hooks/use-columns';

export default defineComponent({
  setup() {
    const businessMapStore = useBusinessMapStore();
    const { isResourcePage } = useWhereAmI();

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
    const searchData = [
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
        name: '任务状态',
        id: 'task_status',
      },
      {
        name: '操作人',
        id: 'operator',
      },
    ];
    !isResourcePage && searchData.splice(4, 1);
    const { CommonTable } = useTable({
      searchOptions: {
        searchData,
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

    // 当前选中的资源类型，默认为全部
    const activeResourceType = ref('all');

    // 操作详情
    const isRecordDetailShow = ref(false);
    const currentDetailInfo = ref(null);
    const showOperationDetail = (detail: any) => {
      isRecordDetailShow.value = true;
      currentDetailInfo.value = detail;
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

    watch(activeResourceType, val => (searchRule.value = val));

    return () => (
      <div class={`operation-record-module${isResourcePage ? ' resource-apply' : ''}`}>
        <div class='common-card-wrap'>
          <CommonTable>
            {{
              operation: () => (
                <BkRadioGroup v-model={activeResourceType.value} type='capsule' class='resource-radio-group'>
                  <BkRadioButton label='all'>全部</BkRadioButton>
                  <BkRadioButton label='security_group'>安全组</BkRadioButton>
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
