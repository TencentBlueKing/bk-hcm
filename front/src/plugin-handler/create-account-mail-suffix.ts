import { ref } from 'vue';
export const PluginHandlerMailbox = {
  suffixText: '' as any,
  isMailValid: ref(false),
  emailRules: [
    {
      trigger: 'change',
      message: '请输入正确格式的邮箱',
      validator: (val: string) => {
        const isValid = /^([a-zA-Z0-9._%-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,6})*$/.test(val);
        PluginHandlerMailbox.isMailValid.value = isValid;
        return isValid;
      },
    },
  ],
};

export type PluginHandlerMailbox = typeof PluginHandlerMailbox;
