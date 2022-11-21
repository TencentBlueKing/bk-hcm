import {
  defineComponent,
  onMounted,
  onUnmounted,
} from 'vue';
import { useI18n } from 'vue-i18n';

export default defineComponent({
  setup() {
    onMounted(() => {
    });

    onUnmounted(() => {
    });

    const { t } = useI18n();

    return () => (
        <span class="test">{t('你好世界')}</span>
    );
  },
});
