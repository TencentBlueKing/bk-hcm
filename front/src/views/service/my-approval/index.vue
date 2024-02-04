<template>
  <div class="template-warp">
    <div>我的审批</div>
  </div>
</template>

<script lang="ts">
import { reactive, watch, toRefs, defineComponent, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import logo from '@/assets/image/logo.png';
import { useAccountStore } from '@/store';
import rightArrow from '@/assets/image/right-arrow.png';
import tcloud from '@/assets/image/tcloud.png';
import { Message } from 'bkui-vue';
import { CloudType, AccountType } from '@/typings';
import { VENDORS } from '@/common/constant';

export default defineComponent({
  name: 'MyApproval',
  setup() {
    const { t } = useI18n();
    const router = useRouter();
    const accountStore = useAccountStore();

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
      tableData: [],
      pagination: {
        count: 0,
        current: 1,
        limit: 10,
      },
      showDeleteBox: false,
      deleteBoxTitle: '',
      syncTitle: t('同步'),
      showSyncBox: false,
      logo,
      rightArrow,
      tcloudSrc: tcloud,
      loading: true,
      dataId: null,
      CloudType,
      AccountType,
      filter: { op: 'and', rules: [] },
    });

    onMounted(async () => {
      /* 获取账号列表接口 */
      // getListCount(); // 数量
      // init(); // 列表
    });
    onUnmounted(() => {});

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
        state.filter.rules = val.reduce((p, v) => {
          if (v.type === 'condition') {
            state.filter.op = v.id || 'and';
          } else {
            p.push({
              field: v.id,
              op: state.isAccurate ? 'eq' : 'cs',
              value: v.values[0].id,
            });
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
      try {
        if (diaType === 'del') {
          // 删除
          await accountStore.accountDelete(state.dataId);
        } else if (diaType === 'sync') {
          // 同步
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
        state.showSyncBox = false;
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
      state.showSyncBox = true;
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
      handleDialogConfirm,
      handleJump,
      handleDelete,
      handleSync,
      handlePageLimitChange,
      handlePageValueChange,
      t,
    };
  },
});
</script>
