import { defineComponent, ref } from 'vue';
import cssModule from './index.module.scss';

import { Alert, Dialog } from 'bkui-vue';

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
      <Dialog title={t('确认账单')} confirmText={t('确认账单')} v-model:isShow={isShow.value}>
        <Alert class='mb12'>{t('如全部云账号已定账，可以同步到OBS')}</Alert>
        <div class={cssModule.item}>
          <span>{t('云厂商')}</span>
          <span>xxx</span>
        </div>
        <div class={cssModule.item}>
          <span>{t('一级账号ID')}</span>
          <span>xxx</span>
        </div>
        <div class={cssModule.item}>
          <span>{t('核算月份')}</span>
          <span>xxx</span>
        </div>
        <div class={cssModule.item}>
          <span>{t('上次定帐时间')}</span>
          <span>xxx</span>
        </div>
        <div class={cssModule.item}>
          <span>{t('运营产品数量')}</span>
          <span>xxx</span>
        </div>
        <div class={cssModule.item}>
          <span>{t('总金额（人民币）')}</span>
          <span class={cssModule.money}>xxx</span>
        </div>
        <div class={cssModule.item}>
          <span>{t('总金额（美金）')}</span>
          <span class={cssModule.money}>xxx</span>
        </div>
      </Dialog>
    );
  },
});
