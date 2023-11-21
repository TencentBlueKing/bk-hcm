import { defineComponent, ref, PropType } from 'vue';
import { Search } from 'bkui-vue/lib/icon';

import './index.scss';

export default defineComponent({
  name: 'search-input',
  emits: ['search', 'update:modelValue'],
  props: {
    width: {
      type: Number as PropType<number>,
      default: 300
    },
    placeholder: {
      type: String as PropType<string>,
      default: '请输入'
    },
    modelValue: {
      type: String as PropType<string>,
    },
  },
  setup(props, ctx) {

    const inputVal = ref('')

    const triggerSearch = () => {
      ctx.emit('update:modelValue', inputVal.value);
      ctx.emit('search');
    }

    return () => (
      <bk-input
        v-model={inputVal}
        style={{'width': `${props.width}px`}}
        clearable={true}
        onClear={() => triggerSearch}
        onInput={() => triggerSearch}>
        {{
          suffix: () => <Search class="search-input-icon" />
        }}
      </bk-input>
    );
  },
});