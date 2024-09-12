import { Input } from 'bkui-vue';
import { defineComponent, ref, watch } from 'vue';
import cssModule from './index.module.scss';
export default defineComponent({
  props: {
    suffixText: String,
    isMailValid: Boolean,
    formModel: Object,
  },
  emits: ['changeEmail'],
  setup(props, { expose, emit }) {
    // 表单input部分
    const email = ref('');

    const isNameValid = ref(false);
    const changeNameValid = (value: boolean) => {
      isNameValid.value = value;
    };

    watch(
      () => email.value,
      () => {
        emit('changeEmail', email.value);
      },
    );

    expose({
      changeNameValid,
    });

    return () => (
      <>
        <div class='flex-row'>
          <Input v-model={email.value} suffix={props.suffixText} />
        </div>
        <p class={cssModule['email-tip']}>
          <i class={['hcm-icon', 'bkhcm-icon-alert', cssModule['email-tip-icon'], cssModule['hcm-icon']]}></i>
          请确保邮箱已按指引配置，否则后续帐号将无法创建
        </p>
      </>
    );
  },
});
