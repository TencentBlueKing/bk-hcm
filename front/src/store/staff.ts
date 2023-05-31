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
    async fetchStaffs() {
      if (this.fetching) return;
      this.fetching = true;
      const prefix = `${BK_COMPONENT_API_URL}/component/compapi/tof3/get_all_staff_info`;
      const params: any = {
        query_type: 'simple_data',
        app_code: 'workbench',
        callback: 'callbackStaff',
      };
      const scriptTag = document.createElement('script');
      scriptTag.setAttribute('type', 'text/javascript');
      scriptTag.setAttribute('src', `${prefix}?${QueryString.stringify(params)}`);

      const headTag = document.getElementsByTagName('head')[0];
      // @ts-ignore
      window[params.callback] = ({ data, result }: { data: any, result: boolean }) => {
        if (result) {
          this.fetching = false;
          this.list = data;
        }
        headTag.removeChild(scriptTag);
        // @ts-ignore
        delete window[params.callback];
      };
      headTag.appendChild(scriptTag);
    },
  },
});
