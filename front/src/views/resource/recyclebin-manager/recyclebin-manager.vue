<template>
  <div class="template-warp">
    <div class="flex-row operate-warp justify-content-between align-items-center mb20">
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
            <bk-button
              text theme="primary"
              @click="handleJump('accountDetail', data.id)"
            >{{data?.name}}
            </bk-button>
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
              <!-- <bk-button text theme="primary" @click="handleSync(props?.data.id)">
              {{t('同步')}}
            </bk-button> -->
              <bk-button
                text theme="primary" @click="handleJump('accountDetail', props?.data.id)"
              >
                {{t('编辑')}}
              </bk-button>
            <!-- <bk-button text theme="primary" @click="handleDelete(props?.data.id, props?.data.name)">
              {{t('删除')}}
            </bk-button> -->
            </div>
          </template>
        </bk-table-column>
      </bk-table>
      <bk-dialog
        :is-show="showDeleteBox"
        :title="deleteBoxTitle"
        :theme="'primary'"
        :quick-close="false"
        @closed="showDeleteBox = false"
        @confirm="() => handleDialogConfirm('del')"
      >
        <div>{{t('删除之后无法恢复账户信息')}}</div>
      </bk-dialog>
    </bk-loading>
  </div>
</template>

<script lang="ts">
import { reactive, watch, toRefs, defineComponent } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { Message } from 'bkui-vue';
import { useAccountStore } from '@/store';
import { CloudType, AccountType } from '@/typings';
import { VENDORS } from '@/common/constant';


export default defineComponent({
  name: 'AccountManageList',
  components: {
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
      loading: true,
      dataId: 0,
      CloudType,
      AccountType,
      filter: { op: 'and', rules: [] },
      recycleTypeData: [{ name: t('主机回收'), value: 'host' }, { name: t('硬盘回收'), value: 'dist' }],
      selectedType: 'host',
    });

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

    // 是否精确
    watch(
      () => state.isAccurate,
      (val) => {
        state.filter.rules.forEach((e: any) => {
          e.op = val ? 'eq' : 'cs';
        });
      },
    );

    getListCount();
    getAccountList();


    const init = () => {
      state.pagination.current = 1;
      state.pagination.limit = 10;
      state.isAccurate = false;
      state.searchValue = [];
      getAccountList();
    };
    // 弹窗确认
    const handleDialogConfirm = async (diaType: string) => {
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
        // 重新请求列表
        init();
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

    // 删除
    const handleDelete = (id: number, name: string) => {
      state.dataId = id;
      state.deleteBoxTitle = `确认要删除${name}?`;
      state.showDeleteBox = true;
    };

    // 同步
    const handleSync = (id: number) => {
      state.dataId = id;
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

    const handleSelected = (v) => {
      state.selectedType = v;
    };

    return {
      ...toRefs(state),
      init,
      handleDialogConfirm,
      handleJump,
      handleDelete,
      handleSync,
      handlePageLimitChange,
      handlePageValueChange,
      handleSelected,
      t,
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

