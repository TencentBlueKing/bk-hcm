import { defineComponent } from 'vue';
import { Table } from 'bkui-vue';
import Search from '../../components/search';
import Button from '../../components/button';
import Amount from '../../components/amount';

export default defineComponent({
  name: 'SubAccountTabPanel',
  setup() {
    const columns = [
      {
        label: '二级账号ID',
        field: 'secondary_account_id',
      },
      {
        label: '二级账号名称',
        field: 'secondary_account_name',
      },
      {
        label: '运营产品',
        field: 'operational_product',
      },
      {
        label: '组织架构',
        field: 'organization_structure',
      },
      {
        label: '已确认账单人民币（元）',
        field: 'confirmed_bill_rmb',
      },
      {
        label: '已确认账单美金（美元）',
        field: 'confirmed_bill_usd',
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
        secondary_account_id: 'SA20001',
        secondary_account_name: '二级账户一',
        operational_product: '云服务器',
        organization_structure: '部门A',
        confirmed_bill_rmb: 12000,
        confirmed_bill_usd: 1800,
        current_bill_rmb: 12500,
        current_bill_usd: 1875,
      },
      {
        secondary_account_id: 'SA20002',
        secondary_account_name: '二级账户二',
        operational_product: '大数据分析',
        organization_structure: '部门B',
        confirmed_bill_rmb: 15000,
        confirmed_bill_usd: 2250,
        current_bill_rmb: 15500,
        current_bill_usd: 2325,
      },
      {
        secondary_account_id: 'SA20003',
        secondary_account_name: '二级账户三',
        operational_product: '人工智能',
        organization_structure: '部门C',
        confirmed_bill_rmb: 18000,
        confirmed_bill_usd: 2700,
        current_bill_rmb: 18500,
        current_bill_usd: 2775,
      },
      {
        secondary_account_id: 'SA20004',
        secondary_account_name: '二级账户四',
        operational_product: '云储存',
        organization_structure: '部门D',
        confirmed_bill_rmb: 10000,
        confirmed_bill_usd: 1500,
        current_bill_rmb: 10500,
        current_bill_usd: 1575,
      },
      {
        secondary_account_id: 'SA20005',
        secondary_account_name: '二级账户五',
        operational_product: '内容分发网络',
        organization_structure: '部门A',
        confirmed_bill_rmb: 13000,
        confirmed_bill_usd: 1950,
        current_bill_rmb: 13500,
        current_bill_usd: 2025,
      },
      {
        secondary_account_id: 'SA20006',
        secondary_account_name: '二级账户六',
        operational_product: '开发工具',
        organization_structure: '部门B',
        confirmed_bill_rmb: 14000,
        confirmed_bill_usd: 2100,
        current_bill_rmb: 14500,
        current_bill_usd: 2175,
      },
      {
        secondary_account_id: 'SA20007',
        secondary_account_name: '二级账户七',
        operational_product: '物联网',
        organization_structure: '部门C',
        confirmed_bill_rmb: 16000,
        confirmed_bill_usd: 2400,
        current_bill_rmb: 16500,
        current_bill_usd: 2475,
      },
      {
        secondary_account_id: 'SA20008',
        secondary_account_name: '二级账户八',
        operational_product: '区块链',
        organization_structure: '部门D',
        confirmed_bill_rmb: 17000,
        confirmed_bill_usd: 2550,
        current_bill_rmb: 17500,
        current_bill_usd: 2625,
      },
      {
        secondary_account_id: 'SA20009',
        secondary_account_name: '二级账户九',
        operational_product: '数据库',
        organization_structure: '部门A',
        confirmed_bill_rmb: 11000,
        confirmed_bill_usd: 1650,
        current_bill_rmb: 11500,
        current_bill_usd: 1725,
      },
      {
        secondary_account_id: 'SA20010',
        secondary_account_name: '二级账户十',
        operational_product: '中间件',
        organization_structure: '部门B',
        confirmed_bill_rmb: 9000,
        confirmed_bill_usd: 1350,
        current_bill_rmb: 9500,
        current_bill_usd: 1425,
      },
    ];

    return () => (
      <>
        <Search />
        <div class='p24' style={{ height: 'calc(100% - 162px)' }}>
          <section class='flex-row align-items-center mb16'>
            <Button noSyncBtn />
            <Amount />
          </section>
          <Table style={{ maxHeight: 'calc(100% - 48px)' }} columns={columns} data={tableData} pagination />
        </div>
      </>
    );
  },
});
