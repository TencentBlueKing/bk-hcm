import { ref } from "vue";

export const pluginHandler = {
  useAccountSelector: () => ({
    isAccountShow: ref(false),
    AccountSelectorCard: null as any,
  }),
  ApplicationForm: null as any,
};

export type PluginHandlerType = typeof pluginHandler;