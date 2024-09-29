import { defineComponent, onMounted, onUnmounted } from 'vue';
import Home from '@/views/home';
import Notice from '@/views/notice/index.vue';

const { ENABLE_NOTICE } = window.PROJECT_CONFIG;

export default defineComponent({
  setup() {
    // const router = useRouter();
    // 设置 rem
    const calcRem = () => {
      const doc = window.document;
      const docEl = doc.documentElement;
      const designWidth = 1580; // 默认设计图宽度
      const maxRate = 2560 / designWidth;
      const minRate = 1280 / designWidth;
      const clientWidth = docEl.getBoundingClientRect().width || window.innerWidth;
      const flexibleRem = Math.max(Math.min(clientWidth / designWidth, maxRate), minRate) * 100;
      docEl.style.fontSize = `${flexibleRem}px`;
      // 项目中没有做 rem 适配, 所以这里直接设置 rem 为 14px, 解决 MagicBox 适配 rem 失效问题
      docEl.style.fontSize = '14px';
    };
    onMounted(() => {
      calcRem();
      window.addEventListener('resize', calcRem, false);
    });
    onUnmounted(() => {
      window.removeEventListener('resize', calcRem, false);
    });
    return () => (
      <div class='full-page flex-column'>
        {ENABLE_NOTICE === 'true' && <Notice />}
        <Home class='flex-1'></Home>
      </div>
    );
  },
});
