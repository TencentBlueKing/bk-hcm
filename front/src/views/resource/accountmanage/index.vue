<template>
  <div class="template-warp">
    <div class="flex-row operate-warp justify-content-between align-items-center mb20">
      <div @click="handleAuth('account_import')">
        <bk-button
          theme="primary" @click="handleJump('accountAdd')"
          :disabled="!authVerifyData.permissionAction.account_import">
          {{t('新增') }}
        </bk-button>
      </div>
      <div class="flex-row input-warp justify-content-between align-items-center">
        <bk-checkbox v-model="isAccurate" class="pr20">
          {{t('精确')}}
        </bk-checkbox>
        <bk-search-select class="bg-white w280" v-model="searchValue" :data="searchData"></bk-search-select>
      </div>
    </div>
    <bk-loading
      :loading="loading"
    >
      <bk-table
        class="table-layout"
        :data="tableData"
        remote-pagination
        :pagination="pagination"
        @page-value-change="handlePageValueChange"
        @page-limit-change="handlePageLimitChange"
        row-hover="auto"
      >
        <bk-table-column
          label="ID"
          prop="id"
          sort
        />
        <bk-table-column
          :label="t('名称')"
          prop="name"
        >
          <template #default="{ data }">
            <div @click="handleAuth('account_edit')">
              <bk-button
                text theme="primary"
                @click="handleJump('accountDetail', data.id)"
                :disabled="!authVerifyData.permissionAction?.account_edit">{{data?.name}}</bk-button>
            </div>
          </template>
        </bk-table-column>
        <bk-table-column
          :label="t('云厂商')"
          prop="vendor"
        >
          <template #default="props">
            {{CloudType[props.data.vendor]}}
          </template>
        </bk-table-column>
        <bk-table-column
          :label="t('类型')"
          prop="type"
        >
          <template #default="{ data }">
            {{AccountType[data?.type]}}
          </template>
        </bk-table-column>
        <bk-table-column
          :label="t('负责人')"
          prop="managers"
        >
          <template #default="{ data }">
            {{data.managers?.join(',')}}
          </template>
        </bk-table-column>
        <bk-table-column
          :label="t('余额')"
          prop="price"
        >
          <template #default="{ data }">
            {{data?.price || '--'}}{{data?.price_unit}}
          </template>
        </bk-table-column>
        <bk-table-column
          :label="t('创建时间')"
          prop="created_at"
        >
          <template #default="{ data }">
            {{data.created_at}}
          </template>
        </bk-table-column>
        <bk-table-column
          :label="t('备注')"
          prop="memo"
        />
        <bk-table-column
          :label="t('操作')"
        >
          <template #default="props">
            <div class="operate-button">
              <bk-button text theme="primary" @click="handleOperate(props?.data.id, 'sync')">
                {{t('同步')}}
              </bk-button>
              <div @click="handleAuth('account_edit')">
                <bk-button
                  text theme="primary" @click="handleJump('accountDetail', props?.data.id)"
                  :disabled="!authVerifyData.permissionAction.account_edit">
                  {{t('编辑')}}
                </bk-button>
              </div>
              <bk-button class="ml15" text theme="primary" @click="handleOperate(props?.data.id, 'del')">
                {{t('删除')}}
              </bk-button>
            </div>
          </template>
        </bk-table-column>
      </bk-table>
      <bk-dialog
        :is-show="showDeleteBox"
        :title="deleteBoxTitle"
        :theme="'primary'"
        :dialog-type="'show'"
      >
        <div v-if="type === 'del'">
          {{t('删除之后无法恢复账户信息')}}
        </div>
        <div v-else>
          <div v-if="btnLoading">{{t('同步中...')}}</div>
          <div v-else>{{t('确认同步该账号')}}</div>
        </div>

        <div class="flex-row btn-warp">
          <bk-button
            class="mr10 dialog-button"
            theme="primary"
            :loading="btnLoading"
            @click="handleDialogConfirm(type)"
          >确认</bk-button>
          <bk-button
            class="mr10 dialog-button"
            @click="handleDialogCancel"
          >取消</bk-button>
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
import { VENDORS } from '@/common/constant';
import { useVerify } from '@/hooks';
import permissionDialog from '@/components/permission-dialog';


export default defineComponent({
  name: 'AccountManageList',
  components: {
    permissionDialog,
  },
  setup() {
    const { t } = useI18n();
    const router = useRouter();
    const accountStore = useAccountStore();

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
      tableData: [],
      pagination: {
        count: 0,
        current: 1,
        limit: 10,
      },
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

    const showPermissionDialog = ref(false);    // 无权限弹窗

    onMounted(async () => {
      /* 获取账号列表接口 */
      // getListCount(); // 数量
      // init(); // 列表
    });

    // 权限hook
    const {
      handlePermissionConfirm,
      handlePermissionDialog,
      handleAuth,
      permissionParams,
      authVerifyData,
    } = useVerify(showPermissionDialog);

    console.log('authVerifyData', authVerifyData);

    // 请求获取列表的总条数
    const getListCount = async () => {
      const params = {
        filter: state.filter,
        page: {
          count: true,
        },
      };
      const res = await accountStore.getAccountList(params);
      state.pagination.count = res?.data.count || 0;
    };

    const getAccountList = async () => {
      state.loading = true;
      try {
        const params = {
          filter: state.filter,
          page: {
            count: false,
            limit: state.pagination.limit,
            start: state.pagination.limit * (state.pagination.current - 1),
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
      (val) => {
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
        state.pagination = {
          count: 0,
          current: 1,
          limit: 10,
        };
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
      state.pagination.current = 1;
      state.pagination.limit = 10;
      state.isAccurate = false;
      state.searchValue = [];
      getAccountList();
    };
    // 弹窗确认
    const handleDialogConfirm = async (diaType: string) => {
      state.btnLoading = true;
      try {
        if (diaType === 'del') {    // 删除
          await accountStore.accountDelete(state.dataId);
        } else if (diaType === 'sync') {    // 同步
          await accountStore.accountSync(state.dataId);
        }
        Message({
          message: t(diaType === 'del' ? '删除成功' : '同步成功'),
          theme: 'success',
        });
        state.btnLoading = false;
        // 重新请求列表
        init();
      } catch (error) {
        console.log(error);
      } finally {
        // state.showDeleteBox = false;
      }
    };

    const handleDialogCancel = () => {
      state.showDeleteBox = false;
      state.btnLoading = false;
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

    // 处理翻页
    const handlePageLimitChange = (limit: number) => {
      state.pagination.limit = limit;
      getAccountList();
    };

    const handlePageValueChange = (value: number) => {
      state.pagination.current = value;
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
    };
  },
});
</script>
<style lang="scss">
.operate-button{
  display: flex;
}
.btn-warp{
  margin-top: 30px;
  justify-content: end;
}
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

