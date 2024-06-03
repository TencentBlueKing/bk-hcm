import { defineComponent } from "vue";
import './index.scss';

export default defineComponent({
  setup(props, ctx) {
    return () => '云账单';
  },
});