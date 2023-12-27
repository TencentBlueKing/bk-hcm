import { PropType, defineComponent, ref } from 'vue';
import { PopConfirm, Input } from 'bkui-vue';
import './index.scss';

export default defineComponent({
  name: 'BatchUpdatePopConfirm',
  props: {
    title: {
      type: String as PropType<string>,
      required: true,
    },
  },
  emits: ['updateValue'],
  setup(props, { emit }) {
    const inputVal = ref('');
    const handleConfirm = () => {
      emit('updateValue', inputVal.value);
      inputVal.value = '';
    };
    return () => (
      <PopConfirm
        width={280}
        trigger='click'
        placement='right'
        extCls='batch-update-popconfirm'
        onConfirm={handleConfirm}>
        {{
          default: () => <i class='hcm-icon bkhcm-icon-batch-edit'></i>,
          content: () => (
            <div class='batch-update-popconfirm-content'>
              <div class='title'>批量删除{props.title}</div>
              <Input v-model={inputVal.value}></Input>
            </div>
          ),
        }}
      </PopConfirm>
    );
  },
});
