import { defineComponent } from 'vue';
import { Table } from 'bkui-vue';
import Button from '../../components/button';
import Amount from '../../components/amount';

export default defineComponent({
  name: 'OperationProductTabPanel',
  setup() {
    const columns = [
      {
        label: '运营产品',
        field: 'operational_product',
      },
      {
        label: '组织架构',
        field: 'organization_structure',
      },
      {
        label: '账单同步人民币（元）',
        field: 'synced_bill_rmb',
      },
      {
        label: '账单同步美金（美元）',
        field: 'synced_bill_usd',
      },
      {
        label: '当前账单人民币（元）',
        field: 'current_bill_rmb',
      },
      {
        label: '当前账单美金（美元）',
        field: 'current_bill_usd',
      },
    ];

    const tableData = [
      {
        operational_product: '云服务器',
        organization_structure: '部门A',
        synced_bill_rmb: 12000,
        synced_bill_usd: 1800,
        current_bill_rmb: 12500,
        current_bill_usd: 1875,
      },
      {
        operational_product: '大数据分析',
        organization_structure: '部门B',
        synced_bill_rmb: 15000,
        synced_bill_usd: 2250,
        current_bill_rmb: 15500,
        current_bill_usd: 2325,
      },
      {
        operational_product: '人工智能',
        organization_structure: '部门C',
        synced_bill_rmb: 18000,
        synced_bill_usd: 2700,
        current_bill_rmb: 18500,
        current_bill_usd: 2775,
      },
    ];

    return () => (
      <div class='full-height p24'>
        <section class='flex-row align-items-center mb16'>
          <Button noSyncBtn />
          <Amount />
        </section>
        <Table style={{ maxHeight: 'calc(100% - 48px)' }} columns={columns} data={tableData} pagination />
      </div>
    );
  },
});
