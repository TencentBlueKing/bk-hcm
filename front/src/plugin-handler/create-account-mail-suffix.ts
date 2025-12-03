import { ref } from 'vue';

const EMAIL_REGEXP =
  /^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$/;

export const PluginHandlerMailbox = {
  suffixText: '' as any,
  isMailValid: ref(false),
  emailRules: [
    {
      trigger: 'change',
      message: '请输入正确格式的邮箱',
      validator: (val: string) => {
        const isValid = EMAIL_REGEXP.test(`${val}${PluginHandlerMailbox.suffixText}`);
        PluginHandlerMailbox.isMailValid.value = isValid;
        return isValid;
      },
    },
  ],
};

export type PluginHandlerMailbox = typeof PluginHandlerMailbox;
