import { defineComponent } from "vue";
import { ArrowsLeft, AngleUpFill, EditLine } from "bkui-vue/lib/icon";

import './index.scss';

export default defineComponent({
  name: 'scheme-selector',
  setup () {
    return () => (
      <div class="scheme-selector">
        <ArrowsLeft class="back-icon" />
        <div class="scheme-name">北美部署方案</div>
        <AngleUpFill class="arrow-icon" />
        <div class="edit-btn">
          <EditLine class="edit-icon" />
          编辑
        </div>
      </div>
    )
  },
});
