import { defineComponent } from "vue";

import './index.scss';

export default defineComponent({
  name: 'idc-map-display',
  setup () {
    return () => (
      <div class="idc-map-display">
        <h3>地图展示</h3>
      </div>
    )
  },
});
