import { PropType, defineComponent } from 'vue';
import { EditLine } from 'bkui-vue/lib/icon';
import './index.scss';

/**
 * 选中项预览组件
 */
export default defineComponent({
  name: 'SelectedItemPreviewComp',
  props: { content: String, onClick: Function as PropType<() => void> },
  setup(props) {
    return () => (
      <div class='selected-item-preview-comp'>
        <div class='content'>{props.content}</div>
        <EditLine class='edit-btn' onClick={props.onClick} />
      </div>
    );
  },
});
