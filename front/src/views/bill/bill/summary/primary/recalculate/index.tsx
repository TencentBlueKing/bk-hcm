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
      <Dialog title={t('重算账单')} confirmText={t('确定重算')} v-model:isShow={isShow.value}>
        <Alert class='mb12'>{t('重算账单为从云上重新同步账单数据，重新计算账单的数据')}</Alert>
        <div class={cssModule.item}>
          <span>{t('核算月份')}</span>
          <span>xxx</span>
        </div>
        <div class={cssModule.item}>
          <span>{t('最近一次核算时间')}</span>
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
