/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */
import http from '@/http';
import { useUserStore } from '@/store';

export interface IUser {
  bk_username: string;
  login_name: string;
  full_name: string;
  display_name: string;
  data_source_type: string;
  owner_tenant_id: string;
  organization_paths: string[];
}

export default function useSearchUser() {
  const userStore = useUserStore();
  const userManageUrl = window.PROJECT_CONFIG.USER_MANAGE_URL;

  const search = async (val: string) => {
    const api = `${userManageUrl}/api/v3/open-web/tenant/users/-/search/?keyword=${val}`;
    const res = await http.get(api, {
      globalHeaders: false,
      globalError: false,
      headers: { 'X-Bk-Tenant-Id': userStore.tenantId },
    });
    if (!res) {
      console.error('fetch user failed');
      return [];
    }
    const users: IUser[] = res?.data ?? [];
    const result = users.map((item) => ({
      id: item.bk_username,
      name: item.display_name,
    }));

    return result;
  };

  const lookup = async (values: string) => {
    const api = `${userManageUrl}/api/v3/open-web/tenant/users/-/lookup/?lookups=${values}&lookup_fields=bk_username`;
    const res = await http.get(api, {
      globalHeaders: false,
      globalError: false,
      headers: { 'X-Bk-Tenant-Id': userStore.tenantId },
    });
    const users: IUser[] = res?.data ?? [];
    return users;
  };

  return {
    search,
    lookup,
  };
}
