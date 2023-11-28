import { defineComponent, ref } from 'vue';
import { Search } from 'bkui-vue/lib/icon';

import './index.scss';

export default defineComponent({
  name: 'search-input',
  emits: ['search', 'update:modelValue'],
  props: {
    width: {
      type: Number,
      default: 300
    },
    placeholder: {
      type: String,
      default: '请输入'
    },
    modelValue: {
      type: String,
      default: ''
    },
  },
  setup(props, ctx) {

    const inputVal = ref(props.modelValue)

    const triggerSearch = () => {
      ctx.emit('update:modelValue', inputVal.value);
      ctx.emit('search');
    }

    return () => (
      <bk-input
        v-model={inputVal.value}
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