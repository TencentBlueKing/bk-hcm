import { PropType, defineComponent } from 'vue';
import { useI18n } from 'vue-i18n';
import { isChinese } from '@/language/i18n';

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
        title={t('无该业务权限')}
        description={
          isChinese ? `请联系 ${props.message} 开通` : `Please contact ${props.message}.`
        }>
      </bk-exception>
    );
  },
});
