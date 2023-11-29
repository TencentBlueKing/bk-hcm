import { defineComponent } from "vue";
import {  useRouter } from 'vue-router';
import { ArrowsLeft, AngleUpFill, EditLine } from "bkui-vue/lib/icon";

import './index.scss';

export default defineComponent({
  name: 'scheme-selector',
  setup () {

    const router = useRouter();

    const goToSchemeList = () => {
      router.push({ name: 'scheme-list' });
    }
    return () => (
      <div class="scheme-selector">
        <ArrowsLeft class="back-icon" onClick={goToSchemeList} />
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
