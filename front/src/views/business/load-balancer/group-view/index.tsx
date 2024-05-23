import { defineComponent } from 'vue';
import { RouterView } from 'vue-router';
// import components
import TargetGroupList from './target-group-list';
import './index.scss';

export default defineComponent({
  name: 'TargetGroupView',
  setup() {
    return () => (
      <div class='group-view-page'>
        <div class='left-container'>
          <TargetGroupList />
        </div>
        <div class='main-container'>
          {/* 四级路由 */}
          <RouterView />
        </div>
      </div>
    );
  },
});
