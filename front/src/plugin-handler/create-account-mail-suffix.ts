import { ref } from 'vue';
import isEmail from 'validator/lib/isEmail';

export const PluginHandlerMailbox = {
  suffixText: '' as any,
  isMailValid: ref(false),
  emailRules: [
    {
      trigger: 'change',
      message: '请输入正确格式的邮箱',
      validator: (val: string) => {
        const isValid = isEmail(`${val}${PluginHandlerMailbox.suffixText}`);
        PluginHandlerMailbox.isMailValid.value = isValid;
        return isValid;
      },
    },
  ],
};

export type PluginHandlerMailbox = typeof PluginHandlerMailbox;
