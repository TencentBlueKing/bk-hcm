import { defineComponent, ref } from 'vue';
import { Button } from 'bkui-vue';
import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import './index.scss';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '../hooks/use-columns';

export default defineComponent({
  setup() {
    const { columns, settings } = useColumns('operationRecord');
    const tableColumns = [
      ...columns,
      {
        label: '操作',
        render: () => (
          <div class='operation-cell'>
            <Button text theme='primary'>
              查看详情
            </Button>
            <Button text theme='primary'>
              再次购买
            </Button>
          </div>
        ),
      },
    ];
    const { CommonTable } = useTable({
      columns: tableColumns,
      searchData: [],
      type: '',
      tableData: [
        {
          actionTime: '2023-04-01 10:00:00',
          resourceName: 'Server-01',
          cloudResourceId: 'id-12345',
          resourceType: 'Virtual Machine',
          operationMethod: 'Create',
          operationSource: 'Web console',
          cloudProvider: 'AWS',
          cloudAccount: 'aws-account-001',
          taskStatus: 'success',
          operator: 'Alice',
        },
        {
          actionTime: '2023-04-01 11:00:00',
          resourceName: 'Database-02',
          cloudResourceId: 'id-67890',
          resourceType: 'Database',
          operationMethod: 'Update',
          operationSource: 'API',
          cloudProvider: 'Azure',
          cloudAccount: 'azure-account-002',
          taskStatus: 'partial_success',
          operator: 'Bob',
        },
        {
          actionTime: '2023-04-01 12:00:00',
          resourceName: 'Storage-03',
          cloudResourceId: 'id-54321',
          resourceType: 'Storage',
          operationMethod: 'Delete',
          operationSource: 'Script',
          cloudProvider: 'Google Cloud',
          cloudAccount: 'gcloud-account-003',
          taskStatus: 'fail',
          operator: 'Charlie',
        },
      ],
      tableExtraOptions: {
        settings: settings.value,
      },
    });

    // 当前选中的资源类型，默认为全部
    const activeResourceType = ref('all');
    return () => (
      <div class='operation-record-module'>
        <CommonTable>
          {{
            operation: () => (
              <BkRadioGroup v-model={activeResourceType.value} type='capsule' class='resource-radio-group'>
                <BkRadioButton label='all'>全部</BkRadioButton>
                <BkRadioButton label='security'>安全组</BkRadioButton>
              </BkRadioGroup>
            ),
          }}
        </CommonTable>
      </div>
    );
  },
});
