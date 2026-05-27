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
    valueType: {
      type: String as PropType<'string' | 'number'>,
      default: 'number',
    },
    min: Number,
    max: Number,
    disabledTip: String,
  },
  emits: ['updateValue'],
  setup(props, { emit, slots }) {
    const inputValue = ref<any>('');
    const handleConfirm = () => {
      emit('updateValue', inputValue.value);
      if (Array.isArray(inputValue.value)) {
        inputValue.value = [];
      } else {
        inputValue.value = '';
      }
    };

    const renderDefaultInput = () => {
      if (props.valueType === 'number') {
        return (
          <Input
            v-model_number={inputValue.value}
            type='number'
            class='no-number-control'
            min={props.min}
            max={props.max}
            placeholder={`${props.min}-${props.max}`}
          />
        );
      }
      return <Input v-model={inputValue.value} />;
    };

    return () => (
      <PopConfirm
        width={280}
        trigger='click'
        placement='bottom-start'
        extCls='batch-update-popconfirm'
        popoverOptions={slots.content ? { disableOutsideClick: true } : {}}
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
              {slots.content
                ? slots.content({ value: inputValue.value, updateValue: (v: any) => (inputValue.value = v) })
                : renderDefaultInput()}
            </div>
          ),
        }}
      </PopConfirm>
    );
  },
});
