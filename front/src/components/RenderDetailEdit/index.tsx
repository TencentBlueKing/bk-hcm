import { Input } from 'bkui-vue';
import { defineComponent, ref, nextTick, watch } from 'vue';
import { useI18n } from 'vue-i18n';

export default defineComponent({
  props: {
    isEdit: {
      type: Boolean,
      default: false,
    },
    fromKey: {
      type: String,
      default: '',
    },
    modelValue: {
      type: [String],
      default: '1',
    },
  },
  emits: ['update:modelValue', 'change', 'input', 'blur'],
  setup(props, ctx) {
    const { t } = useI18n();
    const renderEdit = ref(false);
    const inputRef = ref<InstanceType<typeof Input>>(null);
    console.log('modelValue', props.modelValue, props.isEdit);
    const test = () => {
      // @ts-ignore
      console.log('isEdit', props.isEdit, props.renderEdit, 1113333);
      renderEdit.value = true;
      nextTick(() => {
        // @ts-ignore
        inputRef.value?.focus();
      });
    };


    watch(
      () => props.isEdit,
      () => {
        console.log('props.isEdit', props.isEdit);
        renderEdit.value = props.isEdit;
        // initChart();
      },
    );

    const handleChange = (val: string) => {
      ctx.emit('input', val);
      ctx.emit('change', val);
      ctx.emit('update:modelValue', val);
    };

    const handleBlur = () => {
      ctx.emit('blur', renderEdit.value);
    };
    return () => (
        <span>
            {renderEdit.value ? (
                <Input ref={inputRef} class="w320" placeholder={t('请输入')} modelValue={props.modelValue} onChange={handleChange} onBlur={handleBlur} />
            ) : <span>{props.modelValue}</span>}
            <i onClick={test} class={'icon hcm-icon bkhcm-icon-edit pl15 account-edit-icon'}/>
        </span>
    );
  },


});
