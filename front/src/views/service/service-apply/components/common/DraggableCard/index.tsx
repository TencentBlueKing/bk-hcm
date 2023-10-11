import { defineComponent, ref } from 'vue';
import './index.scss';
import { AngleDown, AngleUp } from 'bkui-vue/lib/icon';

export default defineComponent({
  props: {
    title: {
      required: true,
      type: String,
    },
    index: {
      required: true,
      type: Number,
    },
  },
  setup(props, { slots }) {
    const isExpand = ref(false);
    return () => (
      <div>
        <div class='draggable-card-header'>
          <div onClick={() => isExpand.value = !isExpand.value} class={'draggable-card-header-icon'}>
            {
              isExpand.value ? <AngleUp width={17} height={14}/> : <AngleDown width={17} height={14}/>
            }
          </div>
          <span class={'draggable-card-header-title'}>
            { props.title }
          </span>
          <div class={'draggable-card-header-index'}>
            { props.index }
          </div>
          <i class={'icon bk-icon icon-grag-fill mr16 draggable-card-header-draggable-btn'}></i>
        </div>
        <div class={'draggable-card-container'}>
            {
              slots.default?.()
            }
        </div>
      </div>
    );
  },
});
