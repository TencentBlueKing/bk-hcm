<template>
  <div class="template-warp">
    <div class="flex-row operate-warp justify-content-between align-items-center mb20" v-if="isResourcePage">
      <div>
        <bk-button-group>
          <bk-button
            v-for="item in recycleTypeData"
            :key="item.value"
            :selected="selectedType === item.value"
            @click="handleSelected(item.value)"
          >
            {{ item.name }}
          </bk-button>
        </bk-button-group>
      </div>
      <!-- <div class="flex-row input-warp justify-content-between align-items-center">
        <bk-checkbox v-model="isAccurate" class="pr20">
          {{t('精确')}}
        </bk-checkbox>
        <bk-search-select class="bg-white w280" v-model="searchValue" :data="searchData"></bk-search-select>
      </div> -->
    </div>
    <section v-if="isResourcePage">
      <span
        v-bk-tooltips="{ content: '请勾选主机信息', disabled: selections.length }"
        @click="handleAuth('recycle_bin_manage')">
        <bk-button
          theme="primary"
          :disabled="!selections.length || !authVerifyData?.permissionAction?.recycle_bin_manage"
          @click="handleOperate('destroy')"
        >{{ t('立即销毁') }}
        </bk-button>
      </span>
      <span
        v-bk-tooltips="{ content: '请勾选主机信息', disabled: selections.length }"
        @click="handleAuth('recycle_bin_manage')">
        <bk-button
          class="ml10 mb20"
          theme="primary"
          :disabled="!selections.length || !authVerifyData?.permissionAction?.recycle_bin_manage"
          @click="handleOperate('recover')"
        >{{ t('立即恢复') }}
        </bk-button>
      </span>
    </section>
    <bk-loading
      :loading="isLoading"
    >
      <bk-table
        class="table-layout"
        :data="datas"
        remote-pagination
        :pagination="pagination"
        show-overflow-tooltip
        @page-value-change="handlePageChange"
        @page-limit-change="handlePageSizeChange"
        @selection-change="handleSelectionChange"
        row-hover="auto"
        row-key="id"
      >
        <bk-table-column
          v-if="isResourcePage"
          width="100"
          type="selection"
        />
        <bk-table-column
          label="ID"
          prop="id"
          width="120"
          sort
        />
        <bk-table-column
          :label="t('回收任务ID')"
          prop="task_id"
        >
        </bk-table-column>
        <bk-table-column
          :label="t('云厂商')"
          prop="vendor"
        >
          <template #default="props">
            {{CloudType[props?.data?.vendor]}}
          </template>
        </bk-table-column>
        <bk-table-column
          :label="t('云账号')"
          prop="account_id"
        >
        </bk-table-column>
        <bk-table-column
          :label="t('地域')"
          prop="region"
        >
          <template #default="{ data }">
            {{
              getRegionName(data?.vendor, data?.region)
            }}
          </template>
        </bk-table-column>
        <bk-table-column
          :label="t('资源实例ID')"
          prop="res_id"
        >
          <template #default="{ data }">
            <!-- <bk-button
              text theme="primary" @click="() => {
                handleShowDialog(data?.res_type, data?.res_id, data?.vendor)
              }">
              {{data?.res_id}}
            </bk-button> -->
            {{data?.res_id}}
          </template>
        </bk-table-column>
        <bk-table-column
          :label="selectedType === 'cvm' ? t('资源名称') : t('关联的主机')"
          prop="res_name"
        >
        </bk-table-column>
        <bk-table-column
          :label="t('回收执行人')"
          prop="reviser"
        >
        </bk-table-column>
        <bk-table-column
          :label="t('回收时间')"
          prop="created_at"
        >
          <template #default="{ data }">
            {{data?.created_at}}
          </template>
        </bk-table-column>
        <bk-table-column
          :label="t('状态')"
          prop="status"
        >
          <template #default="{ data }">
            {{
              t(`${RECYCLE_BIN_ITEM_STATUS[data?.status]}`) || data?.status || '--'
            }}
          </template>
        </bk-table-column>
        <bk-table-column
          v-if="isResourcePage"
          :label="t('操作')"
          :min-width="150"
        >
          <template #default="{ data }">
            <span @click="handleAuth('recycle_bin_manage')">
              <bk-button
                text theme="primary"
                :disabled="!authVerifyData?.permissionAction?.recycle_bin_manage || data?.status !== 'wait_recycle'"
                class="mr10" @click="handleOperate('destroy', [data.id])">
                {{t('立即销毁')}}
              </bk-button>
            </span>
            <span @click="handleAuth('recycle_bin_manage')">
              <bk-button
                text theme="primary" @click="handleOperate('recover', [data.id])"
                :disabled="!authVerifyData?.permissionAction?.recycle_bin_manage || data?.status !== 'wait_recycle'"
              >
                {{t('立即恢复')}}
              </bk-button>
            </span>
          </template>
        </bk-table-column>
      </bk-table>
      <bk-dialog
        :is-show="showDeleteBox"
        :title="deleteBoxTitle"
        :theme="'primary'"
        :quick-close="false"
        @closed="showDeleteBox = false"
        @confirm="handleDialogConfirm"
      >
        <div v-if="type === 'destroy'">{{`${selectedType === 'cvm' ? '销毁之后无法恢复主机信息' : '销毁之后无法从云上恢复硬盘'}`}}</div>
        <div v-else>{{t(`将恢复${selectedType === 'cvm' ? '主机' : '硬盘'}信息`)}}</div>
      </bk-dialog>
    </bk-loading>

    <bk-dialog
      :is-show="showResourceInfo"
      :title="selectedType === 'cvm' ? '主机详情' : '硬盘详情'"
      theme="primary">
      <HostInfo v-if="selectedType === 'cvm'" :data="detail" :type="vendor"></HostInfo>
      <HostDrive v-else :data="detail" :type="vendor"></HostDrive>
    </bk-dialog>

    <permission-dialog
      v-model:is-show="showPermissionDialog"
      :params="permissionParams"
      @cancel="handlePermissionDialog"
      @confirm="handlePermissionConfirm"
    ></permission-dialog>
  </div>
</template>

<script lang="ts">
import { reactive, watch, toRefs, defineComponent, ref, computed } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { Message } from 'bkui-vue';
import useQueryCommonList from '@/views/resource/resource-manage/hooks/use-query-list-common';
import { useResourceStore, useAccountStore } from '@/store';
import { CloudType, FilterType } from '@/typings';
import { VENDORS } from '@/common/constant';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import HostInfo from '@/views/resource/resource-manage/children/components/host/host-info/index.vue';
import HostDrive from '@/views/resource/resource-manage/children/components/host/host-drive.vue';
import { useVerify } from '@/hooks';
import { RECYCLE_BIN_ITEM_STATUS } from '@/constants/resource';
import { useRegionsStore } from '@/store/useRegionsStore';

export default defineComponent({
  name: 'RecyclebinManageList',
  components: {
    HostInfo,
    HostDrive,
  },
  setup() {
    const { t } = useI18n();
    const router = useRouter();
    const route = useRoute();
    const resourceStore = useResourceStore();
    const accountStore = useAccountStore();
    const fetchUrl = ref<string>('recycle_records/list');
    const { getRegionName } = useRegionsStore();

    const state = reactive({
      isAccurate: false,    // 是否精确
      searchValue: [],
      searchData: [
        {
          name: '名称',
          id: 'name',
        }, {
          name: '云厂商',
          id: 'vendor',
          children: VENDORS,
        }, {
          name: '负责人',
          id: 'managers',
        },
      ],
      showDeleteBox: false,
      deleteBoxTitle: '',
      loading: true,
      dataId: 0,
      CloudType,
      filter: { op: 'and', rules: [{ field: 'res_type', op: 'eq', value: 'cvm' }] },
      recycleTypeData: [{ name: t('主机回收'), value: 'cvm' }, { name: t('硬盘回收'), value: 'disk' }],
      selectedType: 'cvm',
      type: '',
      selectedIds: [],
      showResourceInfo: false,
      vendor: '',
      detail: {},
    });


    // hooks
    const {
      datas,
      isLoading,
      pagination,
      handlePageSizeChange,
      handlePageChange,
      getList,
    } = useQueryCommonList({ filter: state.filter as FilterType }, fetchUrl);

    const {
      selections,
      handleSelectionChange,
      resetSelections,
    } = useSelection();

    // 选择类型
    const handleSelected = (v) => {
      state.filter.rules = [{
        field: 'res_type',
        op: 'eq',
        value: v,
      }];
      state.selectedType = v;
      resetSelections();
    };

    // 是否精确
    watch(
      () => state.isAccurate,
      (val) => {
        state.filter.rules.forEach((e: any) => {
          e.op = val ? 'eq' : 'cs';
        });
      },
    );

    watch(() => route.params, (params) => {
      console.log(params.type);
      if (params.type) {
        // @ts-ignore
        state.selectedType = params.type;
        handleSelected(params.type);
      }
    }, { immediate: true });

    const isResourcePage = computed(() => {   // 资源下没有业务ID
      return !accountStore.bizs;
    });

    // 弹窗确认
    const handleDialogConfirm = async () => {
      const params: any = {
        record_ids: state.selectedIds,
      };
      try {
        if (state.type === 'destroy') {
          await resourceStore.deleteRecycledData(`${state.selectedType}s`, params);
        } else {
          await resourceStore.recoverRecycledData(`${state.selectedType}s`, params);
        }

        const operate = state.type === 'destroy' ? t('销毁') : t('恢复');
        const resourceName = state.selectedType === 'cvm' ? t('主机') : t('硬盘');
        Message({
          message: `${operate}${resourceName}成功`,
          theme: 'success',
        });
        // 重新请求列表
        getList();
      } catch (error) {
        console.log(error);
      } finally {
        state.showDeleteBox = false;
      }
    };
    // 跳转页面
    const handleJump = (routerName: string, id?: string) => {
      const routerConfig = {
        query: {},
        name: routerName,
      };
      if (id) {
        routerConfig.query = {
          id,
        };
      }
      router.push(routerConfig);
    };

    // 销毁恢复
    const handleOperate = (type: string, ids?: string[]) => {
      console.log('selections', ids, selections.value);
      state.selectedIds = ids ? ids : selections.value.map(e => e.id);
      console.log('state.selectedIds', state.selectedIds);
      state.type = type;
      state.deleteBoxTitle = `确认要 ${type === 'destroy' ? t('销毁') : t('恢复')}`;
      state.showDeleteBox = true;
    };

    // 资源详情
    const handleShowDialog = async (type: string, id: string, vendor: string) => {
      try {
        state.detail = await resourceStore.recycledResourceDetail(`${type}s`, id);
        state.vendor = vendor;
        state.showResourceInfo = true;
      } catch (error) {
        console.log(error);
      }
    };

    getList();

    // 权限hook
    const {
      showPermissionDialog,
      handlePermissionConfirm,
      handlePermissionDialog,
      handleAuth,
      permissionParams,
      authVerifyData,
    } = useVerify();

    return {
      ...toRefs(state),
      handleDialogConfirm,
      handleJump,
      handleOperate,
      handleSelected,
      isLoading,
      handlePageSizeChange,
      handlePageChange,
      pagination,
      datas,
      handleSelectionChange,
      selections,
      isResourcePage,
      handleShowDialog,
      t,
      showPermissionDialog,
      handlePermissionConfirm,
      handlePermissionDialog,
      handleAuth,
      permissionParams,
      authVerifyData,
      RECYCLE_BIN_ITEM_STATUS,
      getRegionName,
    };
  },
});
</script>
<style lang="scss">
  .sync-dialog-warp{
    height: 150px;
    .t-icon{
      height: 42px;
      width: 110px;
    }
    .logo-icon{
        height: 42px;
        width: 42px;
    }
    .arrow-icon{
      position: relative;
      flex: 1;
      overflow: hidden;
      height: 13px;
      line-height: 13px;
      .content{
        width: 130px;
        position: absolute;
        left: 200px;
        animation: 3s move infinite linear;
      }
    }
  }
@-webkit-keyframes move {
  from {
		left: 0%;
	}

	to {
		left: 100%;
	}
}

@keyframes move {
	from {
		left: 0%;
	}

	to {
		left: 100%;
	}
}
</style>

