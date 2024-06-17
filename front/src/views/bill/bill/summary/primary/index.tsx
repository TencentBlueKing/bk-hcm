import { defineComponent } from 'vue';
import { SearchSelect, Table } from 'bkui-vue';
import { useI18n } from 'vue-i18n';
import Button from '../../components/button';
import Amount from '../../components/amount';

export default defineComponent({
  name: 'PrimaryAccountTabPanel',
  setup() {
    const { t } = useI18n();

    const columns = [
      {
        label: '一级账号ID',
        field: 'primary_account_id',
      },
      {
        label: '一级账号名称',
        field: 'primary_account_name',
      },
      {
        label: '账号状态',
        field: 'account_status',
      },
      {
        label: '账单同步（人名币-元）当月',
        field: 'current_month_rmb',
      },
      {
        label: '账单同步（人名币-元）上月',
        field: 'last_month_rmb',
      },
      {
        label: '账单同步（人名币-元）环比',
        field: 'month_on_month_rmb',
      },
      {
        label: '账单同步（美金-美元）当月',
        field: 'current_month_usd',
      },
      {
        label: '账单同步（美金-美元）上月',
        field: 'last_month_usd',
      },
      {
        label: '账单同步（美金-美元）环比',
        field: 'month_on_month_usd',
      },
      {
        label: '当前账单人名币（元）',
        field: 'current_bill_rmb',
      },
      {
        label: '当前账单美金（美元）',
        field: 'current_bill_usd',
      },
      {
        label: '调账人名币（元）',
        field: 'adjustment_rmb',
      },
      {
        label: '调账美金（美元）',
        field: 'adjustment_usd',
      },
    ];

    const tableData = [
      {
        primary_account_id: 'PA10001',
        primary_account_name: '账户一',
        account_status: '活跃',
        current_month_rmb: 5000,
        last_month_rmb: 4800,
        month_on_month_rmb: '4.17%',
        current_month_usd: 750,
        last_month_usd: 720,
        month_on_month_usd: '4.17%',
        current_bill_rmb: 4900,
        current_bill_usd: 740,
        adjustment_rmb: 100,
        adjustment_usd: 15,
      },
      {
        primary_account_id: 'PA10002',
        primary_account_name: '账户二',
        account_status: '冻结',
        current_month_rmb: 6500,
        last_month_rmb: 6300,
        month_on_month_rmb: '3.17%',
        current_month_usd: 970,
        last_month_usd: 945,
        month_on_month_usd: '2.65%',
        current_bill_rmb: 6400,
        current_bill_usd: 960,
        adjustment_rmb: 100,
        adjustment_usd: 10,
      },
      {
        primary_account_id: 'PA10003',
        primary_account_name: '账户三',
        account_status: '活跃',
        current_month_rmb: 7000,
        last_month_rmb: 7100,
        month_on_month_rmb: '-1.41%',
        current_month_usd: 1050,
        last_month_usd: 1065,
        month_on_month_usd: '-1.41%',
        current_bill_rmb: 6900,
        current_bill_usd: 1035,
        adjustment_rmb: 100,
        adjustment_usd: 15,
      },
      {
        primary_account_id: 'PA10004',
        primary_account_name: '账户四',
        account_status: '活跃',
        current_month_rmb: 5300,
        last_month_rmb: 5200,
        month_on_month_rmb: '1.92%',
        current_month_usd: 795,
        last_month_usd: 780,
        month_on_month_usd: '1.92%',
        current_bill_rmb: 5200,
        current_bill_usd: 780,
        adjustment_rmb: 100,
        adjustment_usd: 15,
      },
      {
        primary_account_id: 'PA10005',
        primary_account_name: '账户五',
        account_status: '冻结',
        current_month_rmb: 4500,
        last_month_rmb: 4700,
        month_on_month_rmb: '-4.26%',
        current_month_usd: 675,
        last_month_usd: 705,
        month_on_month_usd: '-4.26%',
        current_bill_rmb: 4400,
        current_bill_usd: 660,
        adjustment_rmb: 100,
        adjustment_usd: 15,
      },
      {
        primary_account_id: 'PA10006',
        primary_account_name: '账户六',
        account_status: '活跃',
        current_month_rmb: 6000,
        last_month_rmb: 5800,
        month_on_month_rmb: '3.45%',
        current_month_usd: 900,
        last_month_usd: 870,
        month_on_month_usd: '3.45%',
        current_bill_rmb: 5900,
        current_bill_usd: 885,
        adjustment_rmb: 100,
        adjustment_usd: 15,
      },
      {
        primary_account_id: 'PA10007',
        primary_account_name: '账户七',
        account_status: '活跃',
        current_month_rmb: 5000,
        last_month_rmb: 4800,
        month_on_month_rmb: '4.17%',
        current_month_usd: 750,
        last_month_usd: 720,
        month_on_month_usd: '4.17%',
        current_bill_rmb: 4900,
        current_bill_usd: 740,
        adjustment_rmb: 100,
        adjustment_usd: 15,
      },
      {
        primary_account_id: 'PA10008',
        primary_account_name: '账户八',
        account_status: '冻结',
        current_month_rmb: 6500,
        last_month_rmb: 6300,
        month_on_month_rmb: '3.17%',
        current_month_usd: 970,
        last_month_usd: 945,
        month_on_month_usd: '2.65%',
        current_bill_rmb: 6400,
        current_bill_usd: 960,
        adjustment_rmb: 100,
        adjustment_usd: 10,
      },
      {
        primary_account_id: 'PA10009',
        primary_account_name: '账户九',
        account_status: '活跃',
        current_month_rmb: 7000,
        last_month_rmb: 7100,
        month_on_month_rmb: '-1.41%',
        current_month_usd: 1050,
        last_month_usd: 1065,
        month_on_month_usd: '-1.41%',
        current_bill_rmb: 6900,
        current_bill_usd: 1035,
        adjustment_rmb: 100,
        adjustment_usd: 15,
      },
      {
        primary_account_id: 'PA10010',
        primary_account_name: '账户十',
        account_status: '活跃',
        current_month_rmb: 5300,
        last_month_rmb: 5200,
        month_on_month_rmb: '1.92%',
        current_month_usd: 795,
        last_month_usd: 780,
        month_on_month_usd: '1.92%',
        current_bill_rmb: 5200,
        current_bill_usd: 780,
        adjustment_rmb: 100,
        adjustment_usd: 15,
      },
    ];

    return () => (
      <div class='full-height p24'>
        <section class='flex-row align-items-center'>
          <Button />
          <div class='flex-row align-items-center ml-auto'>
            <SearchSelect class='w500 mr24' />
            <div>{t('操作记录')}</div>
          </div>
        </section>
        <Amount class='mt16 mb16' />
        <Table style={{ maxHeight: 'calc(100% - 88px)' }} columns={columns} data={tableData} pagination />
      </div>
    );
  },
});
