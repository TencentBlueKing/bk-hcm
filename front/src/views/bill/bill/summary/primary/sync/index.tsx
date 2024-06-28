import { computed, defineComponent, ref, watchEffect } from 'vue';
import cssModule from './index.module.scss';

import { Alert, Button, Checkbox, Dialog, Message } from 'bkui-vue';
import VendorRadioGroup from '@/components/vendor-radio-group';

import { useI18n } from 'vue-i18n';
import { VendorEnum } from '@/common/constant';
import { reqBillsRootAccountSummaryList, reqBillsRootAccountSummarySum, syncRecordsBills } from '@/api/bill';
import { QueryRuleOPEnum } from '@/typings';
import { BillsRootAccountSummaryState, BillsSummarySum } from '@/typings/bill';

export default defineComponent({
  props: {
    billYear: { type: Number, required: true },
    billMonth: { type: Number, required: true },
  },
  setup(props, { expose }) {
    const { t } = useI18n();

    const isShow = ref(false);
    const isLoading = ref(false);
    const vendor = ref(VendorEnum.AZURE);
    const syncInfo = ref<BillsSummarySum>(null);

    const isChecked = ref(false);
    const isConfirmAllBills = ref(false);
    const canSyncBills = computed(() => isChecked.value && isConfirmAllBills.value);

    const triggerShow = (v: boolean) => {
      isShow.value = v;
    };

    const getVendorSyncInfo = async (vendor: VendorEnum) => {
      const res = await reqBillsRootAccountSummarySum({
        bill_year: props.billYear,
        bill_month: props.billMonth,
        filter: { op: QueryRuleOPEnum.AND, rules: [{ field: 'vendor', op: QueryRuleOPEnum.EQ, value: vendor }] },
      });
      syncInfo.value = res.data;
    };

    // 只有当某个云厂商下所有一级账号账单都处于确认状态后，才能进行同步
    const checkVendorAllBillsAreConfirmed = async (vendor: VendorEnum) => {
      const res = await reqBillsRootAccountSummaryList({
        bill_year: props.billYear,
        bill_month: props.billMonth,
        filter: { op: QueryRuleOPEnum.AND, rules: [{ field: 'vendor', op: QueryRuleOPEnum.EQ, value: vendor }] },
        // 一级账号一般不会超过500个
        page: { count: false, start: 0, limit: 500 },
      });
      isConfirmAllBills.value = !res.data.details.find((item) => item.state !== BillsRootAccountSummaryState.confirmed);
    };

    const handleConfirm = async () => {
      isLoading.value = true;
      try {
        await syncRecordsBills({ bill_year: props.billYear, bill_month: props.billMonth, vendor: vendor.value });
        Message({ theme: 'success', message: t('提交同步请求成功') });
        triggerShow(false);
      } finally {
        isLoading.value = false;
      }
    };

    watchEffect(() => {
      if (!isShow.value) return;
      getVendorSyncInfo(vendor.value);
      checkVendorAllBillsAreConfirmed(vendor.value);
    });

    expose({ triggerShow });

    return () => (
      <Dialog
        v-model:isShow={isShow.value}
        width={800}
        title={t('账单同步')}
        isLoading={isLoading.value}
        onConfirm={handleConfirm}>
        {{
          default: () => (
            <>
              <Alert theme='warning'>{t('账单同步必须在 每个月的1号 操作，其余时间不允许同步')}</Alert>
              <section class={cssModule['vendor-wrapper']}>
                <div class={cssModule.title}>{t('云厂商')}</div>
                <VendorRadioGroup v-model={vendor.value} size='small' />
              </section>
              <section class={cssModule['sync-content-wrapper']}>
                <div class={cssModule.title}>{t('同步内容')}</div>
                <div class={cssModule.item}>
                  <span>{t('总金额（人民币）')}</span>
                  <span class={cssModule.money}>￥{syncInfo.value?.cost_map?.USD?.RMBCost || 0}</span>
                </div>
                <div class={cssModule.item}>
                  <span>{t('总金额（美金）')}</span>
                  <span class={cssModule.money}>＄{syncInfo.value?.cost_map?.USD?.Cost || 0}</span>
                </div>
                <div class={cssModule.item}>
                  <span>{'业务数量'}</span>
                  <span class={cssModule.count}>{syncInfo.value?.count || 0}</span>
                </div>
              </section>
              <Alert theme='info' class={cssModule.mb12}>
                {t('在操作前，请确保当前账单核对无误后，再进行同步操作。检查的步骤如下：')}
                <br />
                {t('1.检查XX步骤')}
                <br />
                {t('2.检查XX步骤')}
                <br />
                {t('3.检查XX步骤')}
              </Alert>
              <Checkbox v-model={isChecked.value}>{t('已确认所有步骤正确，可以触发同步操作')}</Checkbox>
            </>
          ),
          footer: () => (
            <>
              <Button
                class={cssModule.button}
                theme='primary'
                disabled={!canSyncBills.value}
                onClick={handleConfirm}
                v-bk-tooltips={{
                  content: isChecked.value
                    ? t('当前云厂商下所有一级账号账单有未确认的账单，无法同步')
                    : t('请勾选确认复选框'),
                  disabled: canSyncBills.value,
                }}>
                {t('同步')}
              </Button>
              <Button class={cssModule.button} onClick={() => triggerShow(false)}>
                {t('取消')}
              </Button>
            </>
          ),
        }}
      </Dialog>
    );
  },
});
