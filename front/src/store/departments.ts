import http from '@/http';
// import { Department } from '@/typings';
import { defineStore } from 'pinia';

export const useDepartmentStore = defineStore({
  id: 'departmentStore',
  state: () => ({
    departmentMap: new Map(),
  }),
  actions: {
    async getRootDept() {
      try {
        const res = await http.get('/usermanage/list_first_level_department/');

        res.forEach((item: any) => {
          this.departmentMap.set(item.id, item);
        });
      } catch (error) {
        console.error(error);
      }
    },
    async getChildDept(dept_id: number) {
      try {
        const parent = this.departmentMap.get(dept_id);
        if (parent) {
          parent.loading = true;
        }
        const res = await http.get('/usermanage/retrieve_department/', {
          params: {
            dept_id,
          },
        });

        res.children.forEach((item: any) => {
          this.departmentMap.set(item.id, {
            ...item,
            parent: dept_id,
          });
        });
        if (parent) {
          Object.assign(parent, {
            loaded: true,
            loading: false,
          });
        }
      } catch (error) {
        console.error(error);
      }
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

