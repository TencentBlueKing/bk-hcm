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
    disabled: {
      type: Boolean,
      default: false,
    },
    // 值类型: string/number
    valueType: {
      type: String as PropType<'string' | 'number'>,
      default: 'number',
    },
    // 当valueType='number'时, 可以设置min,max
    min: Number,
    max: Number,
    disabledTip: String,
  },
  emits: ['updateValue'],
  setup(props, { emit }) {
    const inputValue = ref('');
    const handleConfirm = () => {
      emit('updateValue', inputValue.value);
      inputValue.value = '';
    };
    return () => (
      <PopConfirm
        width={280}
        trigger='click'
        placement='bottom-start'
        extCls='batch-update-popconfirm'
        onConfirm={handleConfirm}
        disabled={props.disabled}>
        {{
          default: () => (
            <i
              class={`hcm-icon bkhcm-icon-batch-edit${props.disabled ? ' disabled' : ''}`}
              v-bk-tooltips={{
                content: props.disabledTip,
                disabled: !props.disabled,
              }}></i>
          ),
          content: () => (
            <div class='batch-update-popconfirm-content'>
              <div class='title'>批量修改{props.title}</div>
              {props.valueType === 'number' ? (
                <Input
                  v-model_number={inputValue.value}
                  type='number'
                  class='no-number-control'
                  min={props.min}
                  max={props.max}
                  placeholder={`${props.min}-${props.max}`}
                />
              ) : (
                <Input v-model={inputValue.value} />
              )}
            </div>
          ),
        }}
      </PopConfirm>
    );
  },
});
