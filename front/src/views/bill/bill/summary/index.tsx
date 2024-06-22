import { defineComponent } from 'vue';
import { RouterView } from 'vue-router';

export default defineComponent({
  name: 'BillSummary',
  setup() {
    return () => (
      <div class='bill-summary-module'>
        <RouterView />
      </div>
    );
  },
});
