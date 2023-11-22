import { defineComponent } from "vue";

import './index.scss';

export default defineComponent({
  name: 'network-heat-map',
  setup () {
    return () => (
      <div class="network-heat-map">
        <h3>网络热力分析</h3>
      </div>
    )
  },
});
