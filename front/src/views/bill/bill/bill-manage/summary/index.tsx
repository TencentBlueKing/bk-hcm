import { defineComponent } from 'vue';
import './index.scss';

export default defineComponent({
  name: 'BillSummary',
  setup() {
    return () => <div class='bill-summary-module'>summary</div>;
  },
});
