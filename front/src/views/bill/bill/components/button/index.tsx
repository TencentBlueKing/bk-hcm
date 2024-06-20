import { PropType, Ref, defineComponent } from 'vue';
import cssModule from './index.module.scss';
import { Button } from 'bkui-vue';
import { useI18n } from 'vue-i18n';

export default defineComponent({
  props: { noSyncBtn: Boolean, billSyncDialogRef: Object as PropType<Ref> },
  setup(props) {
    const { t } = useI18n();

    const handleSync = () => {
      props.billSyncDialogRef.value.triggerShow(true);
    };

    return () => (
      <>
        {!props.noSyncBtn && (
          <Button theme='primary' class={cssModule.button} onClick={handleSync}>
            {t('同步')}
          </Button>
        )}
        <Button class={cssModule.button}>{t('导出')}</Button>
      </>
    );
  },
});
