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
      <!-- <div class="flex-row input-warp justify-content-between align-items-center">
        <bk-checkbox v-model="isAccurate" class="pr20">
          {{t('精确')}}
        </bk-checkbox>
        <bk-search-select class="bg-white w280" v-model="searchValue" :data="searchData"></bk-search-select>
      </div> -->
    </div>
    <bk-button
      theme="primary"
      @click="handleOperate('destroy')"
    >{{ t('立即销毁') }}</bk-button>
    <bk-button
      class="ml20 mb20"
      @click="handleOperate('recover')"
    >{{ t('撤销恢复') }}</bk-button>
    <bk-loading
      :loading="isLoading"
    >
      <bk-table
        class="table-layout"
        :data="tableData"
        remote-pagination
        :pagination="pagination"
        @page-value-change="handlePageChange"
        @page-limit-change="handlePageSizeChange"
        row-hover="auto"
      >
        <bk-table-column
          label="ID"
          prop="id"
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
            {{CloudType[props.data.vendor]}}
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
        </bk-table-column>
        <bk-table-column
          :label="t('实例ID')"
          prop="res_id"
        >
        </bk-table-column>
        <bk-table-column
          :label="selectedType === 'cvms' ? t('资源名称') : t('关联的主机')"
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
            {{data.created_at}}
          </template>
        </bk-table-column>
        <bk-table-column
          :label="t('过期时间')"
          prop="created_at"
        >
          <template #default="{ data }">
            {{data.created_at}}
          </template>
        </bk-table-column>
        <bk-table-column
          :label="t('操作')"
        >
          <template #default="props">
            <div class="operate-button">
              <bk-button text theme="primary" @click="handleOperate('destroy')">
                {{t('立即销毁')}}
              </bk-button>
              <bk-button
                text theme="primary" @click="handleOperate('recover')"
              >
                {{t('撤销恢复')}}
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
        <!-- <div>{{t('删除之后无法恢复账户信息')}}</div> -->
      </bk-dialog>
    </bk-loading>
  </div>
</template>

<script lang="ts">
import { reactive, watch, toRefs, defineComponent } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { Message } from 'bkui-vue';
import useQueryCommonList from '@/views/resource/resource-manage/hooks/use-query-list-common';
import { useResourceStore } from '@/store';
import { CloudType, AccountType } from '@/typings';
import { VENDORS } from '@/common/constant';


export default defineComponent({
  name: 'AccountManageList',
  components: {
  },
  setup() {
    const { t } = useI18n();
    const router = useRouter();
    const resourceStore = useResourceStore();

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
      showDeleteBox: false,
      deleteBoxTitle: '',
      loading: true,
      dataId: 0,
      CloudType,
      AccountType,
      filter: { op: 'and', rules: [{ field: 'res_type', op: 'eq', value: 'cvm' }] },
      recycleTypeData: [{ name: t('主机回收'), value: 'cvms' }, { name: t('硬盘回收'), value: 'disks' }],
      selectedType: 'cvms',
      type: '',
    });


    // hooks
    const {
      datas,
      isLoading,
      pagination,
      handlePageSizeChange,
      handlePageChange,
      getList,
    } = useQueryCommonList({ filter: state.filter }, 'recycle_records/list');

    // 是否精确
    watch(
      () => state.isAccurate,
      (val) => {
        state.filter.rules.forEach((e: any) => {
          e.op = val ? 'eq' : 'cs';
        });
      },
    );

    // 弹窗确认
    const handleDialogConfirm = async () => {
      const params: any = {
        ids: [],
      };
      try {
        if (state.type === 'destroy') {
          await resourceStore.deleteRecycledData(state.selectedType, params);
        } else {
          await resourceStore.recoverRecycledData(state.selectedType, params);
        }

        const operate = state.type === 'destroy' ? t('销毁') : t('回收');
        const resourceName = state.selectedType === 'cvms' ? t('主机') : t('硬盘');
        Message({
          message: `${operate}${resourceName}成功`,
          theme: 'success',
        });
        // 重新请求列表
        // init();
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
    const handleOperate = (type: string) => {
      state.type = type;
      state.deleteBoxTitle = `确认要 ${type === 'destroy' ? t('销毁') : t('恢复')}`;
      state.showDeleteBox = true;
    };

    const handleSelected = (v) => {
      state.filter.rules = [{
        field: 'res_type',
        op: 'eq',
        value: v,
      }];
      getList();
      state.selectedType = v;
    };

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

