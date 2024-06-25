import { defineComponent, ref } from 'vue';
import cssModule from './index.module.scss';

import { Alert, Dialog, Message } from 'bkui-vue';

import { useI18n } from 'vue-i18n';
import { BillsRootAccountSummary } from '@/typings/bill';
import { timeFormatter } from '@/common/util';
import { confirmBillsRootAccountSummary } from '@/api/bill';

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
        await confirmBillsRootAccountSummary({
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
      <Dialog title={t('确认账单')} confirmText={t('确认账单')} v-model:isShow={isShow.value} onConfirm={handleConfirm}>
        <Alert class='mb12'>{t('如全部云账号已定账，可以同步到OBS')}</Alert>
        <div class={cssModule.item}>
          <span>{t('云厂商')}</span>
          <span>{info.value.vendor}</span>
        </div>
        <div class={cssModule.item}>
          <span>{t('一级账号ID')}</span>
          <span>{info.value.root_account_id}</span>
        </div>
        <div class={cssModule.item}>
          <span>{t('核算月份')}</span>
          <span>
            {info.value.bill_year}年{info.value.bill_month}月
          </span>
        </div>
        <div class={cssModule.item}>
          <span>{t('上次确认时间')}</span>
          <span>{timeFormatter(info.value.updated_at)}</span>
        </div>
        <div class={cssModule.item}>
          <span>{t('业务')}</span>
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
