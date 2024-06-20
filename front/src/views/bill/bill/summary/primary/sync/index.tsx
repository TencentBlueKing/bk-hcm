import { defineComponent, ref } from 'vue';
import cssModule from './index.module.scss';

import { Alert, Checkbox, Dialog } from 'bkui-vue';

import { useI18n } from 'vue-i18n';

export default defineComponent({
  setup(_, { expose }) {
    const { t } = useI18n();
    const isShow = ref(false);

    const triggerShow = (v: boolean) => {
      isShow.value = v;
    };

    expose({ triggerShow });

    return () => (
      <Dialog v-model:isShow={isShow.value} width={640} title={t('账单同步')}>
        <Alert theme='warning'>{t('账单同步必须在 每个月的1号 操作，其余时间不允许同步')}</Alert>
        <section class={cssModule['sync-content-wrapper']}>
          <div class={cssModule.title}>{t('同步内容')}</div>
          <div class={cssModule.item}>
            <span>{t('总金额（人民币）')}</span>
            <span class={cssModule.money}>xxx</span>
          </div>
          <div class={cssModule.item}>
            <span>{t('总金额（美金）')}</span>
            <span class={cssModule.money}>xxx</span>
          </div>
          <div class={cssModule.item}>
            <span>{t('运营产品数量')}</span>
            <span class={cssModule.count}>xxx</span>
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
        <Checkbox>{t('已确认所有步骤正确，可以触发同步操作')}</Checkbox>
      </Dialog>
    );
  },
});
