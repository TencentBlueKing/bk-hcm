import { defineComponent } from 'vue';
import empty from '@/assets/image/empty-table.svg';
import './index.scss';

export default defineComponent({
  name: 'Empty',
  props: {
    text: {
      type: String,
      default: '暂无数据',
    },
  },
  setup(props) {
    return () => (
      <div class='empty-wrap'>
        <img class='empty-img' src={empty} alt='' />
        <div class='empty-text'>{props.text}</div>
      </div>
    );
  },
});
