import { PropType, defineComponent, ref } from 'vue';
import './index.scss';

import { Button } from 'bkui-vue';
import { useTable } from '@/hooks/useTable/useTable';
import ImportBillDetailDialog from '../ImportBillDetailDialog';

import { useI18n } from 'vue-i18n';
import { VendorEnum } from '@/common/constant';

export default defineComponent({
  name: 'BillDetailRenderTable',
  props: { vendor: String as PropType<VendorEnum> },
  setup(props) {
    const { t } = useI18n();

    const importBillDetailDialogRef = ref();

    const { CommonTable } = useTable({
      tableOptions: {
        columns: [
          {
            label: '核算年月',
            field: 'accounting_period',
          },
          {
            label: '业务分类',
            field: 'business_category',
          },
          {
            label: '事业群',
            field: 'business_group',
          },
          {
            label: '业务部门',
            field: 'department',
          },
          {
            label: '规划产品',
            field: 'planned_product',
          },
          {
            label: '运营产品',
            field: 'operational_product',
          },
          {
            label: '主帐号ID',
            field: 'main_account_id',
          },
          {
            label: '子帐号ID',
            field: 'sub_account_id',
          },
        ],
        reviewData: [
          {
            accounting_period: '2023-01',
            business_category: '零售',
            business_group: '集团一',
            department: '市场部',
            planned_product: '新款手表',
            operational_product: '智能手环',
            main_account_id: 'MA12345',
            sub_account_id: 'SA12345',
          },
          {
            accounting_period: '2023-02',
            business_category: '电子',
            business_group: '集团二',
            department: '研发部',
            planned_product: '新款手机',
            operational_product: '笔记本电脑',
            main_account_id: 'MA12346',
            sub_account_id: 'SA12346',
          },
          {
            accounting_period: '2023-03',
            business_category: '服务',
            business_group: '集团三',
            department: '客服部',
            planned_product: 'VIP服务',
            operational_product: '标准服务',
            main_account_id: 'MA12347',
            sub_account_id: 'SA12347',
          },
        ],
      },
      searchOptions: {
        disabled: true,
      },
      requestOption: {
        type: '',
      },
    });

    return () => (
      <div class='bill-detail-render-table-container'>
        <CommonTable>
          {{
            operation: () => (
              <>
                {props.vendor === VendorEnum.ZENLAYER && (
                  <Button onClick={() => importBillDetailDialogRef.value.triggerShow(true)}>{t('导入')}</Button>
                )}
                <Button>{t('导出')}</Button>
              </>
            ),
          }}
        </CommonTable>
        <ImportBillDetailDialog ref={importBillDetailDialogRef} />
      </div>
    );
  },
});
