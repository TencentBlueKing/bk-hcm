import { defineComponent, onMounted, onUnmounted } from 'vue';
import { useI18n } from 'vue-i18n';
import bus from '@/common/bus';

export default defineComponent({
  setup() {
    onMounted(() => {});

    onUnmounted(() => {});

    const { t } = useI18n();

    console.error(bus);
    bus.$on('foo', (e) => {
      console.error(e);
    });

    // setTimeout(() => {
    //   bus.$emit('foo', 'ddd');
    // }, 3000);

    // bus.$emit('foo', 42);

    // return () => (
    //     <span class="test">{t('你好世界')}</span>
    // );
    return {
      t,
    };
  },
  render() {
    return <span class='test'>{this.t('你好世界')}</span>;
  },
});
