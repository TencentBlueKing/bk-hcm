import { defineComponent, ref } from 'vue';
import cssModule from './index.module.scss';

import { Alert, Dialog, Message } from 'bkui-vue';

import { useI18n } from 'vue-i18n';
import { BillsRootAccountSummary } from '@/typings/bill';
import { timeFormatter } from '@/common/util';
import { reAccountBillsRootAccountSummary } from '@/api/bill';

export default defineComponent({
  props: { reloadTable: Function },
  setup(props, { expose }) {
    const { t } = useI18n();

    const isShow = ref(false);
    const isLoading = ref(false);
    const info = ref<BillsRootAccountSummary>(null);

    const triggerShow = (v: boolean, data: BillsRootAccountSummary) => {
      isShow.value = v;
      info.value = data;
    };

    const handleConfirm = async () => {
      isLoading.value = true;
      try {
        await reAccountBillsRootAccountSummary({
          bill_year: info.value.bill_year,
          bill_month: info.value.bill_month,
          root_account_id: info.value.root_account_id,
        });
        Message({ theme: 'success', message: '提交成功' });
        isShow.value = false;
        props.reloadTable();
      } finally {
        isLoading.value = false;
      }
    };

    expose({ triggerShow });

    return () => (
      <Dialog
        title={t('重算账单')}
        confirmText={t('确定重算')}
        v-model:isShow={isShow.value}
        onConfirm={handleConfirm}
        isLoading={isLoading.value}>
        <Alert class='mb12'>{t('重算账单为从云上重新同步账单数据，重新计算账单的数据')}</Alert>
        <div class={cssModule.item}>
          <span>{t('核算月份')}</span>
          <span>
            {info.value.bill_year}年{info.value.bill_month}月
          </span>
        </div>
        <div class={cssModule.item}>
          <span>{t('最近一次核算时间')}</span>
          <span>{timeFormatter(info.value.updated_at)}</span>
        </div>
        <div class={cssModule.item}>
          <span>{t('运营产品数量')}</span>
          <span>{info.value.product_num}</span>
        </div>
        <div class={cssModule.item}>
          <span>{t('总金额（人民币）')}</span>
          <span class={cssModule.money}>￥{info.value.current_month_rmb_cost}</span>
        </div>
        <div class={cssModule.item}>
          <span>{t('总金额（美金）')}</span>
          <span class={cssModule.money}>＄{info.value.current_month_cost}</span>
        </div>
      </Dialog>
    );
  },
});
