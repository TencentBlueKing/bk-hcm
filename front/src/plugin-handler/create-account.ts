import { ref } from 'vue';
export const PluginHandlerMailbox = {
  isuffix: '' as any,
  isMailRules: ref(false),
  emailRules: [
    {
      trigger: 'change',
      message: '请输入正确格式的邮箱',
      validator: (val: string) => {
        const isValid = /^([a-zA-Z0-9._%-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,6})*$/.test(val);
        PluginHandlerMailbox.isMailRules.value = isValid;
        return isValid;
      },
    },
  ],
};

export type PluginHandlerMailbox = typeof PluginHandlerMailbox;
