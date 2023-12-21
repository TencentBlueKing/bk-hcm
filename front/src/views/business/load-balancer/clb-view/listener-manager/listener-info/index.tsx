import { defineComponent } from 'vue';
import './index.scss';

export default defineComponent({
  name: 'ListenerInfo',
  setup() {
    return () => (
      <div class="hello-world">
        <h1>监听器基本信息</h1>
      </div>
    );
  },
});
