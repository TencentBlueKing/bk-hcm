import { defineComponent } from 'vue';
import { Loading } from 'bkui-vue';
// import './loading.scss';
export default defineComponent({
  name: 'LoadingComp',
  props: {
    title: {
      type: String,
      default: '',
    },
    size: {
      type: String,
      default: 'nomal',
    },
  },
  render() {
    return (
      <div class='full-width full-height flex-row align-items-center justify-content-center'>
        <Loading size={this.size} title={this.title} />
      </div>
    );
  },
});
