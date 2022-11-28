// @ts-check
import { Staff, StaffType } from '@/typings';
import { defineStore } from 'pinia';
import QueryString from 'qs';
import { shallowRef } from 'vue';
const { BK_HOST } = window.PROJECT_CONFIG;

export const useStaffStore = defineStore({
  id: 'staffStore',
  state: () => ({
    fetching: false,
    list: shallowRef([]),
  }),
  actions: {
    async fetchStaffs(type: StaffType) {
      if (this.list.length > 0 || this.fetching) return;
      this.fetching = true;
      const typeUrlMap = {
        [StaffType.RTX]: 'get_all_staff_info',
        [StaffType.MAIL]: 'get_all_ad_groups',
        [StaffType.ALL]: 'get_all_rtx_and_mail_group',
      };
      const prefix = `//${BK_HOST}/component/compapi/tof3/${typeUrlMap[type]}`;
      const params = {
        query_type: type === StaffType.RTX ? 'simple_data' : undefined,
        app_code: 'workbench',
        callback: 'callbackStaff',
      };
      const scriptTag = document.createElement('script');
      scriptTag.setAttribute('type', 'text/javascript');
      scriptTag.setAttribute('src', `${prefix}?${QueryString.stringify(params)}`);

      const headTag = document.getElementsByTagName('head')[0];
      // @ts-ignore
      window[params.callback] = ({ data, result }: { data: Staff[], result: boolean }) => {
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
