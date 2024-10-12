import { defineComponent } from 'vue';
import { provideBreadcrumb } from './hooks/use-breakcrumb';
import Home from '@/views/home';
import Notice from '@/views/notice/index.vue';

const { ENABLE_NOTICE } = window.PROJECT_CONFIG;

export default defineComponent({
  setup() {
    provideBreadcrumb();
    return () => (
      <div class='full-page flex-column'>
        {ENABLE_NOTICE === 'true' && <Notice />}
        <Home class='flex-1'></Home>
      </div>
    );
  },
});
