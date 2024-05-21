<template>
  <div class="template-warp">
    <div class="flex-row operate-warp justify-content-between align-items-center mb20">
      <div @click="handleAuth('account_import')">
        <!-- <bk-button
          theme="primary" @click="handleJump('accountAdd')"
          :disabled="!authVerifyData.permissionAction.account_import">
          {{t('新增') }}
        </bk-button> -->
      </div>
      <div class="flex-row input-warp justify-content-between align-items-center">
        <bk-checkbox v-model="isAccurate" class="pr20">
          {{ t('精确') }}
        </bk-checkbox>
        <bk-search-select
          class="bg-white w280"
          :conditions="[]"
          v-model="searchValue"
          :data="searchData"
        ></bk-search-select>
      </div>
    </div>
    <bk-loading :loading="loading">
      <bk-table
        class="table-layout"
        :data="tableData"
        remote-pagination
        :pagination="{
          count: pageCount,
          limit: memoPageSize,
          current: memoPageIndex,
        }"
        show-overflow-tooltip
        @page-value-change="handlePageValueChange"
        @page-limit-change="handlePageLimitChange"
        row-hover="auto"
        row-key="id"
      >
        <bk-table-column label="ID" width="120" prop="id" sort />
        <bk-table-column :label="t('名称')" prop="name" sort>
          <template #default="props">
            <bk-button text theme="primary" @click="handleJump('accountDetail', props?.row.id, true)">
              {{ props?.row.name }}
            </bk-button>
          </template>
        </bk-table-column>

        <bk-table-column label="账号类型" prop="type" sort>
          <template #default="props">
            {{ AccountType[props?.row.type] }}
          </template>
        </bk-table-column>

        <bk-table-column :label="t('云厂商')" prop="vendor" sort>
          <template #default="props">
            {{ CloudType[props?.row?.vendor] }}
          </template>
        </bk-table-column>

        <bk-table-column label="站点类型" prop="site" sort>
          <template #default="props">
            {{ SITE_TYPE_MAP[props?.row.site] }}
          </template>
        </bk-table-column>
        <bk-table-column label="所属业务" prop="bk_biz_ids" sort>
          <template #default="props">
            {{
              props?.row?.bk_biz_ids
                ?.join(',')
                ?.split(',')
                ?.map((v: string) => getNameFromBusinessMap(+v))
                ?.join(',')
            }}
          </template>
        </bk-table-column>
        <bk-table-column :label="t('负责人')" prop="managers" sort>
          <template #default="props">
            {{ props?.row.managers?.join(',') }}
          </template>
        </bk-table-column>
        <bk-table-column label="创建人" prop="creator" sort></bk-table-column>
        <bk-table-column label="修改人" prop="reviser" sort></bk-table-column>
        <!-- <bk-table-column
          :label="t('余额')"
          prop="price"
        >
          <template #default="{ data }">
            {{data?.price || '--'}}{{data?.price_unit}}
          </template>
        </bk-table-column> -->
        <bk-table-column :label="t('创建时间')" prop="created_at" sort>
          <template #default="props">
            {{ timeFormatter(props?.row.created_at) }}
          </template>
        </bk-table-column>
        <bk-table-column label="修改时间" prop="updated_at" sort>
          <template #default="props">
            {{ timeFormatter(props?.row.updated_at) }}
          </template>
        </bk-table-column>
        <bk-table-column :label="t('备注')" prop="memo" />
        <bk-table-column :label="t('操作')">
          <template #default="props">
            <div class="operate-button">
              <div @click="handleAuth('account_edit')">
                <bk-button
                  text
                  theme="primary"
                  @click="handleJump('accountDetail', props?.row.id)"
                  :disabled="!authVerifyData.permissionAction.account_edit"
                >
                  {{ t('编辑') }}
                </bk-button>
              </div>
              <bk-button class="ml15" text theme="primary" @click="handleOperate(props?.row.id, 'del')">
                {{ t('删除') }}
              </bk-button>
              <bk-button
                text
                theme="primary"
                @click="handleOperate(props?.row.id, 'sync')"
                v-if="props?.row?.type === 'resource'"
              >
                {{ t('同步') }}
              </bk-button>
            </div>
          </template>
        </bk-table-column>
      </bk-table>
      <bk-dialog :is-show="showDeleteBox" :title="deleteBoxTitle" :theme="'primary'" :dialog-type="'show'">
        <div v-if="type === 'del'">
          {{ t('删除之后无法恢复账户信息') }}
        </div>
        <div v-else>
          <div v-if="btnLoading">{{ t('同步中...') }}</div>
          <div v-else>{{ t('同步该账号下的资源，点击确定后，立即触发同步任务') }}</div>
        </div>

        <div class="flex-row btn-warp">
          <bk-button
            class="mr10 dialog-button"
            theme="primary"
            :loading="btnLoading"
            @click="handleDialogConfirm(type)"
          >
            确认
          </bk-button>
          <bk-button class="mr10 dialog-button" @click="handleDialogCancel">取消</bk-button>
        </div>
      </bk-dialog>
    </bk-loading>

    <permission-dialog
      v-model:is-show="showPermissionDialog"
      :params="permissionParams"
      @cancel="handlePermissionDialog"
      @confirm="handlePermissionConfirm"
    ></permission-dialog>
  </div>
</template>

<script lang="ts">
import { reactive, watch, toRefs, defineComponent, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useAccountStore } from '@/store';
import rightArrow from '@/assets/image/right-arrow.png';
import { Message } from 'bkui-vue';
import { CloudType, AccountType } from '@/typings';
import { ACCOUNT_TYPES, SITE_TYPES, SITE_TYPE_MAP, VENDORS } from '@/common/constant';
import { useVerify } from '@/hooks';
import { useMemoPagination, DEFAULT_PAGE_INDEX, DEFAULT_PAGE_SIZE } from '@/hooks/useMemoPagination';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { timeFormatter } from '@/common/util';

export default defineComponent({
  name: 'AccountManageList',
  setup() {
    const { t } = useI18n();
    const router = useRouter();
    const accountStore = useAccountStore();
    const { getNameFromBusinessMap, businessMap } = useBusinessMapStore();
    const { setMemoPageSize, setMemoPageIndex, memoPageIndex, memoPageSize, memoPageStart } = useMemoPagination();
    const state = reactive({
      isAccurate: false, // 是否精确
      searchValue: [],
      searchData: [
        {
          name: '名称',
          id: 'name',
        },
        {
          name: '账号类型',
          id: 'type',
          children: ACCOUNT_TYPES,
        },
        {
          name: '云厂商',
          id: 'vendor',
          children: VENDORS,
        },
        {
          name: '站点类型',
          id: 'site',
          children: SITE_TYPES,
        },
        {
          name: '负责人',
          id: 'managers',
        },
        {
          name: '创建人',
          id: 'creator',
        },
        {
          name: '修改人',
          id: 'reviser',
        },
      ],
      tableData: [],
      showDeleteBox: false,
      deleteBoxTitle: '',
      syncTitle: t('同步'),
      rightArrow,
      loading: true,
      dataId: null,
      CloudType,
      AccountType,
      filter: { op: 'and', rules: [] },
      type: '',
      btnLoading: false,
    });

    const pageCount = ref(0);

    onMounted(async () => {
      /* 获取账号列表接口 */
      // getListCount(); // 数量
      // init(); // 列表
    });

    // 权限hook
    const {
      showPermissionDialog,
      handlePermissionConfirm,
      handlePermissionDialog,
      handleAuth,
      permissionParams,
      authVerifyData,
    } = useVerify();

    // 请求获取列表的总条数
    const getListCount = async () => {
      const params = {
        filter: state.filter,
        page: {
          count: true,
        },
      };
      const res = await accountStore.getAccountList(params);
      pageCount.value = res?.data.count || 0;
    };

    const getAccountList = async () => {
      state.loading = true;
      try {
        const params = {
          filter: state.filter,
          page: {
            count: false,
            limit: memoPageSize.value,
            start: memoPageStart.value,
            sort: 'created_at',
            order: 'DESC',
          },
        };
        const res = await accountStore.getAccountList(params);
        state.tableData = res.data.details;
      } catch (error) {
        console.log(error);
      } finally {
        state.loading = false;
      }
    };

    // 搜索数据
    watch(
      () => state.searchValue,
      (val, oldVal) => {
        console.log('val', val);
        state.filter.rules = val.reduce((p, v) => {
          if (v.type === 'condition') {
            state.filter.op = v.id || 'and';
          } else {
            console.log('v.values[0].id', v.values[0].id);
            if (v.id === 'managers') {
              p.push({
                field: v.id,
                op: 'json_contains',
                value: v.values[0].id,
              });
            } else {
              p.push({
                field: v.id,
                op: state.isAccurate ? 'eq' : 'cs',
                value: v.values[0].id,
              });
            }
          }
          return p;
        }, []);
        pageCount.value = 0;
        if (oldVal !== undefined) {
          setMemoPageIndex(DEFAULT_PAGE_INDEX);
          setMemoPageSize(DEFAULT_PAGE_SIZE);
        }
        /* 获取账号列表接口 */
        getListCount(); // 数量
        getAccountList(); // 列表
      },
      {
        deep: true,
        immediate: true,
      },
    );

    // 是否精确
    watch(
      () => state.isAccurate,
      (val) => {
        state.filter.rules.forEach((e: any) => {
          e.op = val ? 'eq' : 'cs';
        });
      },
    );

    const init = () => {
      setMemoPageIndex(DEFAULT_PAGE_INDEX);
      setMemoPageSize(DEFAULT_PAGE_SIZE);
      state.isAccurate = false;
      state.searchValue = [];
      getAccountList();
    };
    // 弹窗确认
    const handleDialogConfirm = async (diaType: string) => {
      state.btnLoading = true;
      try {
        if (diaType === 'del') {
          // 删除
          await accountStore.accountDelete(state.dataId);
        } else if (diaType === 'sync') {
          // 同步
          await accountStore.accountSync(state.dataId);
        }
        Message({
          message: t(diaType === 'del' ? '删除成功' : t('本次同步任务触发成功。如需再次同步，请在20分钟后重试')),
          theme: 'success',
        });
        state.btnLoading = false;
        // 重新请求列表
        init();
      } catch (error) {
        console.log(error);
      } finally {
        state.btnLoading = false;
        state.showDeleteBox = false;
      }
    };

    const handleDialogCancel = () => {
      state.showDeleteBox = false;
      state.btnLoading = false;
    };
    // 跳转页面
    const handleJump = (routerName: string, id?: string, isDetail?: boolean) => {
      const routerConfig = {
        query: {},
        name: routerName,
      };
      if (id) {
        routerConfig.query = {
          accountId: id,
          isDetail,
        };
      }
      router.push(routerConfig);
    };

    // 删除
    const handleOperate = async (id: number, type: string) => {
      state.dataId = id;
      state.type = type;
      if (type === 'del') {
        try {
          await accountStore.accountDeleteValidate(state.dataId);
          state.deleteBoxTitle = '确认删除';
          state.showDeleteBox = true;
        } catch (error) {
          console.log(error);
        }
      } else {
        state.deleteBoxTitle = '确认同步';
        state.showDeleteBox = true;
      }
    };

    const handlePageLimitChange = (limit: number) => {
      setMemoPageSize(limit);
      setMemoPageIndex(DEFAULT_PAGE_INDEX);
      getAccountList();
    };

    const handlePageValueChange = (value: number) => {
      setMemoPageIndex(value);
      getAccountList();
    };

    return {
      ...toRefs(state),
      init,
      showPermissionDialog,
      handlePermissionDialog,
      handlePermissionConfirm,
      handleDialogConfirm,
      handleJump,
      handleOperate,
      handleAuth,
      permissionParams,
      authVerifyData,
      handlePageLimitChange,
      handlePageValueChange,
      handleDialogCancel,
      t,
      pageCount,
      memoPageIndex,
      memoPageSize,
      SITE_TYPE_MAP,
      getNameFromBusinessMap,
      businessMap,
      timeFormatter,
    };
  },
});
</script>
<style lang="scss">
.operate-button {
  display: flex;
}
.btn-warp {
  margin-top: 30px;
  justify-content: end;
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
