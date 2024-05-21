import http from '@/http';
// import { Department } from '@/typings';
import { defineStore } from 'pinia';
import QueryString from 'qs';
const { BK_COMPONENT_API_URL } = window.PROJECT_CONFIG;

export const useDepartmentStore = defineStore({
  id: 'departmentStore',
  state: () => ({
    departmentMap: new Map(),
  }),
  actions: {
    async fetchDepartMents(field: string, lookups: number) {
      const prefix = `${BK_COMPONENT_API_URL}/api/c/compapi/v2/usermanage/fe_list_departments/`;
      const params = {
        app_code: 'magicbox',
        no_page: true,
        lookup_field: 'parent',
        // exact_lookups: lookups,
        callback: 'callbackDepart',
      };
      const scriptTag = document.createElement('script');
      scriptTag.setAttribute('type', 'text/javascript');
      scriptTag.setAttribute('src', `${prefix}?${QueryString.stringify(params)}`);

      const headTag = document.getElementsByTagName('head')[0];
      // @ts-ignore
      window[params.callback] = ({ data, result }: { data: Staff[]; result: boolean }) => {
        if (result) {
          if (field === 'level') {
            data.forEach((item: any) => {
              this.departmentMap.set(item.id, item);
            });
          } else {
            data.forEach((item: any) => {
              this.departmentMap.set(item.id, {
                ...item,
                parent: lookups,
              });
            });
            const parent = this.departmentMap.get(lookups);
            Object.assign(parent, {
              loaded: true,
              loading: false,
            });
          }
          // return data;
        }
        headTag.removeChild(scriptTag);
        // @ts-ignore
        delete window[params.callback];
      };
      headTag.appendChild(scriptTag);
    },
    async getRootDept() {
      try {
        this.fetchDepartMents('level', 0);
      } catch (error) {
        console.error(error);
      }
    },
    async getChildDept(dept_id: number) {
      const parent = this.departmentMap.get(dept_id);
      if (parent) {
        parent.loading = true;
      }
      this.fetchDepartMents('parent', dept_id);
    },
    async getParentIds(dept_ids: string) {
      const res = await http.get('/usermanage/batch_request_department_parent_id/', {
        params: {
          dept_ids,
        },
      });
      return res;
    },
  },
});
