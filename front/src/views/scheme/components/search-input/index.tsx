import { defineComponent, ref } from 'vue';
import { Search } from 'bkui-vue/lib/icon';

import './index.scss';

export default defineComponent({
  name: 'SearchInput',
  props: {
    width: {
      type: Number,
      default: 300,
    },
    placeholder: {
      type: String,
      default: '请输入',
    },
    modelValue: {
      type: String,
      default: '',
    },
  },
  emits: ['search', 'update:modelValue'],
  setup(props, ctx) {
    const inputVal = ref(props.modelValue);

    const handleInput = (val: string) => {
      if (val === '') {
        triggerSearch();
      }
    };

    const triggerSearch = () => {
      ctx.emit('update:modelValue', inputVal.value);
      ctx.emit('search');
    };

    return () => (
      <bk-input
        v-model={inputVal.value}
        style={{ width: `${props.width}px` }}
        clearable={true}
        placeholder={props.placeholder}
        onClear={triggerSearch}
        onEnter={triggerSearch}
        onInput={handleInput}>
        {{
          suffix: () => <Search class='search-input-icon' />,
        }}
      </bk-input>
    );
  },
});
