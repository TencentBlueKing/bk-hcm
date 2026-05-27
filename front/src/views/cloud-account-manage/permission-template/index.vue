<script setup lang="ts">
import { ref, computed, inject, type Ref, watch, reactive, useTemplateRef } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { Plus } from 'bkui-vue/lib/icon';
import { Message } from 'bkui-vue';
import { ModelPropertySearch, ModelPropertyColumn } from '@/model/typings';
import useSearchQs from '@/hooks/use-search-qs';
import usePage from '@/hooks/use-page';
import { VendorEnum } from '@/common/constant';
import { transformFlatCondition } from '@/utils/search';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { usePermissionTemplateStore } from '@/store/cloud-account-manage/permission-template';
import routerAction from '@/router/utils/action';
import { MENU_SERVICE_TICKET_DETAILS } from '@/constants/menu-symbol';
import type { ISearchCondition } from '@/views/cloud-account-manage/permission-template/typings';
import Search from './children/list/search/search.vue';
import { SearchConditionFactory } from './children/list/search/condition-factory';
import DataList from './children/list/data-list/data-list.vue';
import { TableColumnFactory } from './children/list/data-list/column-factory';
import type { IPermissionTemplateItem } from '@/store/cloud-account-manage/permission-template';
import CreateForm from './children/form/create.vue';
import EditForm from './children/form/edit.vue';
import Details from './children/details/details.vue';
import DeleteDialog from './children/delete-dialog.vue';
import {
  AUTH_BIZ_CREATE_PERMISSION_TEMPLATE,
  AUTH_BIZ_UPDATE_PERMISSION_TEMPLATE,
  AUTH_BIZ_DELETE_PERMISSION_TEMPLATE,
} from '@/constants/auth-symbols';
import { getTypeData } from '@/views/cloud-account-manage/permission-template/utils';
const currentVendor = inject<Ref<VendorEnum>>('currentVendor', ref(VendorEnum.TCLOUD));
const permissionTemplateStore = usePermissionTemplateStore();
const route = useRoute();
const router = useRouter();
const { getBizsId } = useWhereAmI();

const { pagination, getPageParams } = usePage();

const bizId = computed(() => getBizsId());

const searchModel = computed(() => SearchConditionFactory.createModel(currentVendor.value));
const searchFields = computed<ModelPropertySearch[]>(() => searchModel.value.getProperties());
const condition = ref<ISearchCondition>({});

const columnModel = TableColumnFactory.createModel(currentVendor.value);
const dataListColumns = computed<ModelPropertyColumn[]>(() => columnModel.getProperties());
const templateList = ref<IPermissionTemplateItem[]>([]);

const searchQs = useSearchQs({ key: 'filter', properties: searchFields.value });
const sortParams = ref<{ sort: string; order: string }>({ sort: 'created_at', order: 'DESC' });

const createState = reactive({
  isShow: false,
  data: null,
});

const editState = reactive({
  isShow: false,
  data: null,
});

const detailsState = reactive({
  isShow: false,
  data: null,
});

const deleteState = reactive({
  isShow: false,
  data: null as IPermissionTemplateItem | null,
});

const createFormRef = useTemplateRef<typeof CreateForm>('createFormRef');
const editFormRef = useTemplateRef<typeof EditForm>('editFormRef');

const fetchList = async () => {
  const { list = [], count } = await permissionTemplateStore.getPermissionTemplateList(
    bizId.value,
    currentVendor.value,
    {
      ...transformFlatCondition(condition.value, searchFields.value),
      page: getPageParams(pagination, sortParams.value),
    },
  );
  pagination.count = count;
  templateList.value = list;
};

watch(
  () => route.query,
  async (query) => {
    condition.value = searchQs.get(query);

    pagination.current = Number(query.page) || 1;
    pagination.limit = Number(query.limit) || pagination.limit;

    sortParams.value = {
      sort: (query.sort || 'created_at') as string,
      order: (query.order || 'DESC') as string,
    };

    if (!query?.id) {
      await fetchList();
    }
  },
  { immediate: true },
);

const handleSearch = (condition: ISearchCondition) => {
  searchQs.set(condition);
};

const handleReset = () => {
  searchQs.clear();
};

const handleCreate = () => {
  createState.isShow = true;
  createState.data = {
    type: '1',
  };
};
const handleEdit = (row: IPermissionTemplateItem) => {
  editState.isShow = true;
  editState.data = { ...row, type: '1' };
};

const handleCreateSubmit = async () => {
  await createFormRef.value.validate();
  const formData = createFormRef.value.getFormData();
  const { type, policy_document, ...postData } = formData;
  const result = await permissionTemplateStore.createPermissionTemplate(bizId.value, currentVendor.value, postData);
  Message({ theme: 'success', message: '新建成功' });
  if (result.id) {
    routerAction.redirect({
      name: MENU_SERVICE_TICKET_DETAILS,
      query: {
        id: result.id,
        type: 'account',
      },
    });
  }
};

const handleEditSubmit = async () => {
  await editFormRef.value.validate();
  const formData = editFormRef.value.getFormData();
  const { changedFormData } = editFormRef.value;
  const { type, policy_document, account_id, ...postData } = changedFormData;
  const result = await permissionTemplateStore.updatePermissionTemplate(bizId.value, currentVendor.value, {
    id: formData.id,
    ...postData,
  });
  Message({ theme: 'success', message: '编辑成功' });
  if (result.id) {
    routerAction.redirect({
      name: MENU_SERVICE_TICKET_DETAILS,
      query: {
        id: result.id,
        type: 'account',
      },
    });
  }
};

const handleViewDetails = (row: IPermissionTemplateItem) => {
  detailsState.data = row;
  detailsState.isShow = true;
  // 将 id 写入 URL query，支持分享/刷新/浏览器后退
  router.replace({ query: { ...route.query, id: row.id, _t: undefined } });
};

// 弹窗被用户手动关闭时，同步移除 URL 中的 id
watch(
  () => detailsState.isShow,
  (val) => {
    if (!val && route.query.id) {
      const query = { ...route.query };
      delete query.id;
      router.replace({ query });
    }
  },
);

// 监听 route.query.id，使用详情接口加载并打开弹窗
watch(
  () => route.query.id,
  async (id) => {
    if (!id) {
      detailsState.isShow = false;
      detailsState.data = null;
      return;
    }
    try {
      const detail = await permissionTemplateStore.getPermissionTemplateDetail(
        bizId.value,
        currentVendor.value,
        id as string,
      );
      if (detail) {
        detailsState.data = detail;
        detailsState.isShow = true;
      } else {
        Message({ theme: 'warning', message: '未找到权限模板数据' });
        router.replace({ query: { ...route.query, id: undefined } });
      }
    } catch (error) {
      console.error('获取权限模板详情失败:', error);
      Message({ theme: 'error', message: '获取权限模板详情失败' });
      router.replace({ query: { ...route.query, id: undefined } });
    }
  },
  { immediate: true },
);

const handleDelete = (row: IPermissionTemplateItem) => {
  deleteState.isShow = true;
  deleteState.data = row;
};

const handleDeleteConfirm = async () => {
  const result = await permissionTemplateStore.deletePermissionTemplate(bizId.value, currentVendor.value, {
    id: deleteState.data.id,
  });
  deleteState.isShow = false;
  if (result?.id) {
    routerAction.redirect({
      name: MENU_SERVICE_TICKET_DETAILS,
      query: {
        id: result.id,
        type: 'account',
      },
    });
  }
};
</script>

<template>
  <div class="permission-template-list">
    <Search :fields="searchFields" :condition="condition" @search="handleSearch" @reset="handleReset" />

    <div class="table-panel">
      <div class="toolbar">
        <hcm-auth :sign="{ type: AUTH_BIZ_CREATE_PERMISSION_TEMPLATE, relation: [bizId] }" v-slot="{ noPerm }">
          <bk-button theme="primary" :disabled="noPerm" @click="handleCreate">
            <Plus style="font-size: 22px" />
            新建权限模板
          </bk-button>
        </hcm-auth>
      </div>
      <DataList
        v-bkloading="{ loading: permissionTemplateStore.listLoading }"
        :columns="dataListColumns"
        :list="templateList"
        :pagination="pagination"
        @view-details="handleViewDetails"
        @delete="handleDelete"
        @edit="handleEdit"
      />
    </div>
  </div>

  <bk-sideslider v-model:is-show="createState.isShow" render-directive="if" width="640" title="新建权限模板">
    <template #default>
      <CreateForm ref="createFormRef" :data="createState.data" />
    </template>
    <template #footer>
      <div class="sideslider-footer">
        <bk-button theme="primary" :loading="permissionTemplateStore.createLoading" @click="handleCreateSubmit">
          提交
        </bk-button>
        <bk-button @click="createState.isShow = false">取消</bk-button>
      </div>
    </template>
  </bk-sideslider>

  <bk-sideslider v-model:is-show="editState.isShow" render-directive="if" width="640" title="编辑权限模板">
    <template #default>
      <EditForm ref="editFormRef" :data="editState.data" />
    </template>
    <template #footer>
      <div class="sideslider-footer">
        <bk-button
          theme="primary"
          :disabled="!editFormRef?.isChanged"
          :loading="permissionTemplateStore.updateLoading"
          @click="handleEditSubmit"
        >
          提交
        </bk-button>
        <bk-button @click="editState.isShow = false">取消</bk-button>
      </div>
    </template>
  </bk-sideslider>

  <DeleteDialog
    v-model="deleteState.isShow"
    :data="deleteState.data"
    :loading="permissionTemplateStore.deleteLoading"
    @confirm="handleDeleteConfirm"
  />

  <bk-sideslider v-model:is-show="detailsState.isShow" render-directive="if" width="640">
    <template #header>
      <div class="sideslider-details-header">
        <span>权限模板详情</span>
        <div class="actions">
          <hcm-auth :sign="{ type: AUTH_BIZ_UPDATE_PERMISSION_TEMPLATE, relation: [bizId] }" v-slot="{ noPerm }">
            <bk-button
              outline
              theme="primary"
              :disabled="noPerm || !getTypeData(detailsState.data).isCloudCustom"
              @click="handleEdit(detailsState.data)"
              v-bk-tooltips="{
                content: '仅云自定义模板可编辑',
                disabled: getTypeData(detailsState.data).isCloudCustom,
              }"
            >
              编辑
            </bk-button>
          </hcm-auth>
          <hcm-auth :sign="{ type: AUTH_BIZ_DELETE_PERMISSION_TEMPLATE, relation: [bizId] }" v-slot="{ noPerm }">
            <bk-button
              outline
              :disabled="
                noPerm ||
                getTypeData(detailsState.data).isCloudPreset ||
                detailsState.data.associated_sub_account_count > 0
              "
              @click="handleDelete(detailsState.data)"
              v-bk-tooltips="{
                content: getTypeData(detailsState.data).isCloudPreset ? '云系统预设不可删除' : '有三级账号关联不可删除',
                disabled: !(
                  getTypeData(detailsState.data).isCloudPreset || detailsState.data.associated_sub_account_count > 0
                ),
              }"
            >
              删除
            </bk-button>
          </hcm-auth>
        </div>
      </div>
    </template>
    <template #default>
      <div class="sideslider-details-content">
        <Details :data="detailsState.data" />
      </div>
    </template>
  </bk-sideslider>
</template>

<style lang="scss" scoped>
.permission-template-list {
  height: 100%;

  .table-panel {
    background: #fff;
    border-radius: 2px;
    box-shadow: 0 2px 4px 0 #1919290d;
    margin: 24px;
    padding: 16px 24px;
  }

  .toolbar {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 16px;
  }
}

.sideslider-footer {
  display: flex;
  align-items: center;
  gap: 6px;

  .bk-button {
    min-width: 88px;
  }
}

.sideslider-details-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  height: 100%;
  padding-right: 24px;

  .actions {
    display: flex;
    align-items: center;
    gap: 8px;
  }
}

.sideslider-details-content {
  height: calc(100vh - 52px);
  padding: 24px;
  background: #f5f7fa;
}
</style>
