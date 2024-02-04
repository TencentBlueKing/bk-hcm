import { PropType, defineComponent } from 'vue';
import { useI18n } from 'vue-i18n';

export default defineComponent({
  props: {
    message: {
      type: String as PropType<string>,
    },
  },
  setup: (props) => {
    const { t } = useI18n();
    return () => (
      <bk-exception
        class='exception-wrap-item'
        type='403'
        title={t('无该应用访问权限')}
        description={props.message}></bk-exception>
    );
  },
});
