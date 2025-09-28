<template>
  <div class="recycle-manager-page">
    <!-- <bk-button-group>
      <bk-button
        v-for="item in recycleTypeData"
        :key="item.value"
        :selected="selectedType === item.value"
        @click="handleSelected(item.value)"
      >
        {{ item.name }}
      </bk-button>
    </bk-button-group> -->
    <bk-tab type="card-grid" v-model:active="selectedType">
      <template #setting>
        <div
          class="setting-icon-container"
          @click="() => (isSettingDialogShow = true)"
          v-show="resourceAccountStore.resourceAccount?.id"
        >
          <i class="hcm-icon bkhcm-icon-shezhi" diasbled></i>
        </div>
      </template>
      <bk-tab-panel v-for="item in recycleTypeData" :key="item.value" :name="item.value" :label="item.name">
        <section class="header-container">
          <span
            v-bk-tooltips="{
              content: `请勾选${selectedType === 'cvm' ? '主机' : '硬盘'}信息`,
              disabled: selections.length,
            }"
            @click="handleAuth('recycle_bin_manage')"
          >
            <bk-button
              :disabled="!selections.length || !authVerifyData?.permissionAction?.recycle_bin_manage"
              @click="handleOperate('destroy')"
            >
              {{ t('立即销毁') }}
            </bk-button>
          </span>
          <span
            v-bk-tooltips="{
              content: `请勾选${selectedType === 'cvm' ? '主机' : '硬盘'}信息`,
              disabled: selections.length,
            }"
            @click="handleAuth('recycle_bin_manage')"
          >
            <bk-button
              class="ml8"
              :disabled="!selections.length || !authVerifyData?.permissionAction?.recycle_bin_manage"
              @click="handleOperate('recover')"
            >
              {{ t('立即恢复') }}
            </bk-button>
          </span>
          <SearchSelect
            class="w500 common-search-selector"
            v-model="searchVal"
            :data="searchData"
            value-behavior="need-key"
          />
        </section>
        <bk-loading :loading="isLoading" opacity="1">
          <bk-table
            :key="selectedType"
            class="table-layout"
            :data="datas"
            remote-pagination
            :pagination="pagination"
            @page-value-change="handlePageChange"
            @page-limit-change="handlePageSizeChange"
            @selection-change="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable)"
            @select-all="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable, true)"
            row-hover="auto"
            row-key="id"
            :is-row-select-enable="isRowSelectEnable"
            show-overflow-tooltip
          >
            <bk-table-column width="30" min-width="30" type="selection" />
            <bk-table-column :label="`${selectedType === 'cvm' ? '主机' : '硬盘'}ID`" prop="cloud_res_id" sort>
              <template #default="props">
                <bk-button
                  theme="primary"
                  text
                  @click="
                    () =>
                      handleClick(props?.data?.vendor, props?.data?.res_id, selectedType === 'cvm' ? 'host' : 'drive')
                  "
                >
                  {{ props?.data?.cloud_res_id }}
                </bk-button>
              </template>
            </bk-table-column>
            <bk-table-column label="名称" prop="res_name">
              <template #default="{ cell }">
                {{ cell || '--' }}
              </template>
            </bk-table-column>
            <!-- <bk-table-column
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
            </bk-table-column> -->
            <bk-table-column label="所属主机" prop="detail_cvm_id" v-if="selectedType === 'disk'">
              <template #default="{ data }">
                {{ data?.detail?.cvm_id || '--' }}
                <i
                  class="hcm-icon bkhcm-icon-link related-cvm-link"
                  v-if="data?.detail?.cvm_id"
                  @click="() => handleLink(data.detail.cvm_id)"
                />
              </template>
            </bk-table-column>
            <bk-table-column :label="t('地域')" prop="region">
              <template #default="{ data }">
                {{ getRegionName(data?.vendor, data?.region) }}
              </template>
            </bk-table-column>
            <!-- <bk-table-column
              :label="t('资源实例ID')"
              prop="res_id"
            >
              <template #default="{ data }">
                <bk-button
                  text theme="primary" @click="() => {
                    handleShowDialog(data?.res_type, data?.res_id, data?.vendor)
                  }">
                  {{data?.res_id}}
                </bk-button>
                {{data?.res_id}}
              </template>
            </bk-table-column>
            <bk-table-column
              :label="selectedType === 'cvm' ? t('资源名称') : t('关联的主机')"
              prop="res_name"
            >
            </bk-table-column> -->
            <bk-table-column label="回收人" prop="reviser"></bk-table-column>
            <bk-table-column label="进入回收站时间" prop="created_at" :sort="true">
              <template #default="{ cell }">
                {{ timeFormatter(cell) }}
              </template>
            </bk-table-column>
            <bk-table-column label="过期时间" prop="recycled_at" :sort="true">
              <template #default="{ cell }">
                <bk-tag theme="danger">
                  {{ moment(cell).fromNow() }}
                </bk-tag>
                {{ timeFormatter(cell) }}
              </template>
            </bk-table-column>
            <bk-table-column v-if="isResourcePage" :label="t('操作')" :min-width="150">
              <template #default="{ data }">
                <span @click="handleAuth('recycle_bin_manage')">
                  <bk-button
                    text
                    class="mr10"
                    theme="primary"
                    @click="handleOperate('destroy', [data.id])"
                    v-bk-tooltips="generateTooltipsOptions(data)"
                    :disabled="
                      !authVerifyData?.permissionAction?.recycle_bin_manage ||
                      data?.recycle_type === 'related' ||
                      (whereAmI === Senarios.resource && data?.bk_biz_id !== -1)
                    "
                  >
                    销毁
                  </bk-button>
                </span>
                <span @click="handleAuth('recycle_bin_manage')">
                  <bk-button
                    text
                    theme="primary"
                    @click="handleOperate('recover', [data.id])"
                    v-bk-tooltips="generateTooltipsOptions(data)"
                    :disabled="
                      !authVerifyData?.permissionAction?.recycle_bin_manage ||
                      data?.recycle_type === 'related' ||
                      (whereAmI === Senarios.resource && data?.bk_biz_id !== -1)
                    "
                  >
                    恢复
                  </bk-button>
                </span>
              </template>
            </bk-table-column>
          </bk-table>
        </bk-loading>
      </bk-tab-panel>
    </bk-tab>

    <bk-dialog
      :is-show="showDeleteBox"
      :title="deleteBoxTitle"
      :theme="'primary'"
      :quick-close="false"
      @closed="showDeleteBox = false"
      @confirm="handleDialogConfirm"
    >
      <div v-if="type === 'destroy'">
        {{ `${selectedType === 'cvm' ? '销毁之后无法恢复主机信息' : '销毁之后无法从云上恢复硬盘'}` }}
      </div>
      <div v-else>{{ t(`将恢复${selectedType === 'cvm' ? '主机' : '硬盘'}信息`) }}</div>
    </bk-dialog>

    <bk-dialog :is-show="showResourceInfo" :title="selectedType === 'cvm' ? '主机详情' : '硬盘详情'" theme="primary">
      <HostInfo v-if="selectedType === 'cvm'" :data="detail" :type="vendor"></HostInfo>
      <HostDrive v-else :data="detail" :type="vendor"></HostDrive>
    </bk-dialog>

    <bk-dialog
      :is-show="isSettingDialogShow"
      title="回收站配置"
      @closed="() => (isSettingDialogShow = false)"
      @confirm="handleSettingConfirm"
      theme="primary"
      :is-loading="isSettingDialogLoading"
    >
      保留时长
      <bk-select v-model="recycleReserveTime" class="mt6">
        <bk-option v-for="item in RESERVE_TIME_SET" :key="item.value" :value="item.value" :label="item.label" />
      </bk-select>
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
import { reactive, watch, toRefs, defineComponent, ref, computed, onMounted } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { Message, SearchSelect } from 'bkui-vue';
import useQueryCommonList from '@/views/resource/resource-manage/hooks/use-query-list-common';
import { useResourceStore, useAccountStore } from '@/store';
import { CloudType, FilterType, QueryRuleOPEnum } from '@/typings';
import { VENDORS, VendorEnum } from '@/common/constant';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import HostInfo from '@/views/resource/resource-manage/children/components/host/host-info/index.vue';
import HostDrive from '@/views/resource/resource-manage/children/components/host/host-drive.vue';
import { useVerify } from '@/hooks';
import { RECYCLE_BIN_ITEM_STATUS } from '@/constants/resource';
import { useRegionsStore } from '@/store/useRegionsStore';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import moment from 'moment';
import http from '@/http';
import { useWhereAmI, Senarios } from '@/hooks/useWhereAmI';
import { timeFormatter } from '@/common/util';

export default defineComponent({
  name: 'RecyclebinManageList',
  components: {
    HostInfo,
    HostDrive,
    SearchSelect,
  },
  setup() {
    const { t } = useI18n();
    const router = useRouter();
    const route = useRoute();
    const resourceStore = useResourceStore();
    const accountStore = useAccountStore();
    const fetchUrl = ref<string>('recycle_records/list');
    const { getRegionName } = useRegionsStore();
    const resourceAccountStore = useResourceAccountStore();
    const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;
    const { whereAmI } = useWhereAmI();
    const searchVal = ref([]);
    const searchData = [
      {
        name: 'ID',
        id: 'res_id',
      },
    ];
    const isRowSelectEnable = ({ row, isCheckAll }: any) => {
      if (isCheckAll) return true;
      // if (whereAmI.value === Senarios.resource && row.id) {
      //   return row.bk_biz_id === -1;
      // }
      return isCurRowSelectEnable(row);
    };

    const isCurRowSelectEnable = (row: any) => {
      return !(row?.recycle_type === 'related' || (whereAmI.value === Senarios.resource && row?.bk_biz_id !== -1));
    };

    const state = reactive({
      isAccurate: false, // 是否精确
      searchValue: [],
      searchData: [
        {
          name: '名称',
          id: 'name',
        },
        {
          name: '云厂商',
          id: 'vendor',
          children: VENDORS,
        },
        {
          name: '负责人',
          id: 'managers',
        },
      ],
      showDeleteBox: false,
      deleteBoxTitle: '',
      loading: true,
      dataId: 0,
      CloudType,
      filter: {
        op: 'and',
        rules: [
          { field: 'res_type', op: 'eq', value: 'cvm' },
          { field: 'status', op: 'eq', value: 'wait_recycle' },
        ],
      },
      recycleTypeData: [
        { name: t('主机回收'), value: 'cvm' },
        { name: t('硬盘回收'), value: 'disk' },
      ],
      selectedType: 'cvm',
      type: '',
      selectedIds: [],
      showResourceInfo: false,
      vendor: '',
      detail: {},
    });

    const isSettingDialogShow = ref(false);
    const isSettingDialogLoading = ref(false);
    const recycleReserveTime = ref(48);
    const RESERVE_TIME_SET = new Array(8)
      .fill(0)
      .map((_val, idx) => idx)
      .map((num) => ({
        label: `${num}${num > 0 ? '天' : ''}`,
        value: num * 24,
      }));

    // hooks
    const { datas, isLoading, pagination, handlePageSizeChange, handlePageChange, getList } = useQueryCommonList(
      { filter: state.filter as FilterType },
      fetchUrl,
    );

    const { selections, handleSelectionChange, resetSelections } = useSelection();

    // 确定回收站保留时长
    const handleSettingConfirm = async () => {
      isSettingDialogLoading.value = true;
      try {
        await http.patch(
          `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/accounts/${resourceAccountStore.resourceAccount?.id}`,
          {
            account_id: resourceAccountStore.resourceAccount?.id,
            recycle_reserve_time: recycleReserveTime.value,
          },
        );
        Message({
          theme: 'success',
          message: '配置成功',
        });
        isSettingDialogShow.value = false;
        resourceAccountStore.resourceAccount.recycle_reserve_time = recycleReserveTime.value;
      } finally {
        isSettingDialogLoading.value = false;
      }
    };

    const handleLink = (cvmId: string) => {
      route.query.cvm = cvmId;
      state.selectedType = 'cvm';
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

    watch(
      () => resourceAccountStore.resourceAccount,
      (account) => {
        if (whereAmI.value !== Senarios.resource) return;
        const idx = state.filter.rules.findIndex(({ field }) => field === 'account_id');
        if (!account?.id) {
          if (idx > -1) state.filter.rules.splice(idx);
          return;
        }
        const rule = {
          field: 'account_id',
          op: QueryRuleOPEnum.EQ,
          value: account.id,
        };
        if (idx === -1) state.filter.rules.push(rule);
        else state.filter.rules[idx] = rule;

        recycleReserveTime.value = resourceAccountStore.resourceAccount.recycle_reserve_time;
      },
      {
        immediate: true,
        deep: true,
      },
    );

    watch(
      () => route.query.cvm,
      (cvm) => {
        if (Array.isArray(cvm)) return;
        if (!cvm) {
          searchVal.value = [];
          return;
        }
        searchVal.value = [
          {
            id: 'res_id',
            name: 'ID',
            values: [
              {
                id: cvm,
                name: cvm,
              },
            ],
          },
        ];
      },
      {
        immediate: true,
      },
    );

    watch(
      () => searchVal.value,
      (vals) => {
        const idx = state.filter.rules.findIndex(({ field }) => field === 'res_id');
        if (idx !== -1) state.filter.rules.splice(idx, 1);
        if (!vals.length) return;
        state.filter.rules = state.filter.rules.concat(
          Array.isArray(vals)
            ? vals.map((val: any) => ({
                field: val.id,
                op: QueryRuleOPEnum.EQ,
                value: val.values[0].id,
              }))
            : [],
        );
      },
      {
        immediate: true,
      },
    );

    watch(
      () => route.query.type,
      (type) => {
        if (type === 'disk') state.selectedType = 'disk';
        if (type === 'cvm') state.selectedType = 'cvm';
      },
      {
        immediate: true,
      },
    );

    watch(
      () => state.selectedType,
      (type) => {
        const rule = {
          field: 'res_type',
          op: 'eq',
          value: type,
        };
        const idx = state.filter.rules.findIndex(({ field }) => field === 'res_type');
        if (idx === -1) state.filter.rules.push(rule);
        else state.filter.rules[idx] = rule;
        state.selectedType = type;
        router.replace({
          path: whereAmI.value === Senarios.business ? '/business/recyclebin' : '/resource/resource/recycle',
          query: {
            ...route.query,
            type,
            cvm: type === 'disk' ? undefined : route.query.cvm,
          },
        });
        resetSelections();
      },
      {
        immediate: true,
      },
    );

    onMounted(() => {
      moment.updateLocale('zh-cn', {
        relativeTime: {
          future: '余%s',
          past: '%s前',
          s: '%d秒',
          ss: '%d秒',
          m: '1分钟',
          mm: '%d分钟',
          h: '1小时',
          hh: '%d小时',
          d: '1天',
          dd: '%d天',
          w: '1周',
          ww: '%d周',
          M: '1月',
          MM: '%d月',
          y: '1年',
          yy: '%d年',
        },
      });
    });

    const isResourcePage = computed(() => {
      // 资源下没有业务ID
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
    const generateTooltipsOptions = (data: any) => {
      if (data?.recycle_type === 'related') {
        return {
          content: '该硬盘随主机回收，不可单独操作',
          disabled: data?.recycle_type !== 'related',
        };
      }
      if (data?.bk_biz_id !== -1) {
        return {
          content: '该硬盘仅可在业务下操作',
          disabled: data?.bk_biz_id === -1,
        };
      }
      return {
        disabled: true,
      };
    };
    const handleOperate = (type: string, ids?: string[]) => {
      state.selectedIds = ids ? ids : selections.value.map((e) => e.id);
      state.type = type;
      state.deleteBoxTitle = `确认要 ${type === 'destroy' ? t('销毁') : t('恢复')}`;
      state.showDeleteBox = true;
    };

    // 资源详情
    const handleShowDialog = async (type: string, id: string, vendor: string) => {
      state.detail = await resourceStore.recycledResourceDetail(`${type}s`, id);
      state.vendor = vendor;
      state.showResourceInfo = true;
    };

    const handleClick = (vendor: VendorEnum, id: number, type: 'drive' | 'host') => {
      const routeInfo: any = {
        query: {
          id,
          type: vendor,
        },
      };
      // 业务下
      if (route.path.includes('business')) {
        routeInfo.query.bizs = accountStore.bizs;
        Object.assign(routeInfo, {
          name: `${type}BusinessDetail`,
        });
      } else {
        Object.assign(routeInfo, {
          name: 'resourceDetail',
          params: {
            type,
          },
        });
      }
      router.push(routeInfo);
    };

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
      generateTooltipsOptions,
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
      isRowSelectEnable,
      showPermissionDialog,
      handlePermissionConfirm,
      handlePermissionDialog,
      handleAuth,
      permissionParams,
      authVerifyData,
      RECYCLE_BIN_ITEM_STATUS,
      getRegionName,
      isSettingDialogShow,
      recycleReserveTime,
      RESERVE_TIME_SET,
      handleSettingConfirm,
      moment,
      timeFormatter,
      handleClick,
      isSettingDialogLoading,
      resourceAccountStore,
      handleLink,
      searchData,
      searchVal,
      whereAmI,
      Senarios,
      isCurRowSelectEnable,
    };
  },
});
</script>
<style lang="scss" scoped>
.recycle-manager-page {
  :deep(.bk-tab) {
    height: calc(100vh - 200px);
    .bk-tab-header-item {
      height: 42px;
    }
    .bk-tab-content {
      padding: 16px 24px;
      height: calc(100% - 42px);

      .bk-nested-loading {
        height: calc(100% - 48px);
        .bk-table {
          max-height: 100%;
        }
      }
    }
  }
}
.operate-warp {
  :deep(.bk-tab-header) {
    line-height: normal !important;

    .bk-tab-header-item {
      padding: 0 24px;
    }
  }
}
.sync-dialog-warp {
  height: 150px;
  .t-icon {
    height: 42px;
    width: 110px;
  }
  .logo-icon {
    height: 42px;
    width: 42px;
  }
  .arrow-icon {
    position: relative;
    flex: 1;
    overflow: hidden;
    height: 13px;
    line-height: 13px;
    .content {
      width: 130px;
      position: absolute;
      left: 200px;
      animation: 3s move infinite linear;
    }
  }
}
.setting-icon-container {
  width: 32px;
  height: 32px;
  background: #ffffff;
  box-shadow: 0 2px 4px 0 #1919290d;
  border-radius: 2px;
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: pointer;
}
.mt6 {
  margin-top: 6px;
}
.related-cvm-link {
  margin-left: 4px;
  cursor: pointer;
  color: #3a84ff;
}
.header-container {
  display: flex;
  justify-content: space-between;
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
