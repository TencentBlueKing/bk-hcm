// @ts-check
import { defineStore } from 'pinia';
import QueryString from 'qs';
import { shallowRef } from 'vue';
const { BK_COMPONENT_API_URL } = window.PROJECT_CONFIG;

export const useStaffStore = defineStore({
  id: 'staffStore',
  state: () => ({
    fetching: false,
    list: shallowRef([]),
  }),
  actions: {
    async fetchStaffs(name?: string) {
      if (this.fetching) return;
      this.fetching = true;
      const prefix = `${BK_COMPONENT_API_URL}/api/c/compapi/v2/usermanage/fs_list_users`;
      const params: any = {
        app_code: 'bk-magicbox',
        page: 1,
        page_size: 200,
        callback: 'callbackStaff',
      };
      if (name) {
        params.fuzzy_lookups = name;
      }
      const scriptTag = document.createElement('script');
      scriptTag.setAttribute('type', 'text/javascript');
      scriptTag.setAttribute('src', `${prefix}?${QueryString.stringify(params)}`);

      const headTag = document.getElementsByTagName('head')[0];
      // @ts-ignore
      window[params.callback] = ({ data, result }: { data: any; result: boolean }) => {
        if (result) {
          this.fetching = false;
          // this.list = [...data.results, ...this.list];
          this.list = data.results;
        }
        headTag.removeChild(scriptTag);
        // @ts-ignore
        delete window[params.callback];
      };
      headTag.appendChild(scriptTag);
    },
  },
});
