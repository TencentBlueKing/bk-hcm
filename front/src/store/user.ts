import { ref } from 'vue';
import { defineStore } from 'pinia';
import http, { jsonp } from '@/http';

export interface IUserItem {
  username: string;
  display_name: string;
  domain?: string;
  logo?: string;
  category_id?: string;
  id?: number;
  category_name?: string;
}

export interface ISearchResponse {
  code: number;
  data: {
    count?: number;
    results: IUserItem[];
  };
  message?: string;
  [key: string]: any;
}

export const useUserStore = defineStore('user', () => {
  const username = ref('');
  const searchLoading = ref(false);
  const userList = ref<IUserItem[]>([]);
  const memberDefaultList = ref<string[]>([]);

  // 获取当前用户信息
  const userInfo = async () => {
    const res = await http.get('/api/v1/web/users');
    username.value = res.data.username;
    memberDefaultList.value.push(res.data.username);
  };

  const setMemberDefaultList = (list: string[]) => {
    memberDefaultList.value = list;
  };

  const searchUseBK = (value: string) => {
    const api = `${window.PROJECT_CONFIG.BK_COMPONENT_API_URL}/api/c/compapi/v2/usermanage/fs_list_users`;
    const params = {
      app_code: 'bk-magicbox',
      page: 1,
      page_size: 50,
      fuzzy_lookups: value,
    };
    return jsonp<ISearchResponse>(api, params);
  };

  const searchUseOther = (value: string): Promise<ISearchResponse> => {
    return http.get('/api/user/list', { user: value });
  };

  const getBKUser = (value: string[]) => {
    const api = `${window.PROJECT_CONFIG.BK_COMPONENT_API_URL}/api/c/compapi/v2/usermanage/fs_list_users`;
    // TODO: 100限制
    const params = {
      app_code: 'bk-magicbox',
      page: 1,
      page_size: 100,
      exact_lookups: value.join(','),
    };
    return jsonp<ISearchResponse>(api, params);
  };

  const getOtherUser = (value: string[]): Promise<ISearchResponse> => {
    return http.get('/api/user/list', { users: value });
  };

  // 适配未接入bk用户管理
  const searchUserFn = window.PROJECT_CONFIG.BK_COMPONENT_API_URL ? searchUseBK : searchUseOther;
  const getUserFn = window.PROJECT_CONFIG.BK_COMPONENT_API_URL ? getBKUser : getOtherUser;

  // 通过用户名精确获取用户，用于根据用户名，得到完全用户信息
  const getUserByName = async (value: string[]) => {
    const res = await getUserFn(value);
    const list = res?.data?.results || [];

    for (const user of list) {
      if (!userList.value.some((item) => item.username === user.username)) {
        userList.value.push(user);
      }
    }

    return list;
  };

  // 通过关键字模糊搜索用户，用于人员选择器
  const search = async (value: string) => {
    searchLoading.value = true;
    try {
      const res = await searchUserFn(value);
      return res?.data?.results || [];
    } finally {
      searchLoading.value = false;
    }
  };

  return {
    username,
    searchLoading,
    getUserByName,
    userList,
    userInfo,
    memberDefaultList,
    setMemberDefaultList,
    search,
  };
});
