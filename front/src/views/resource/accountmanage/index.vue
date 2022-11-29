<template>
  <div class="account-warp">
    <div class="operate-warp flex-row justify-content-between align-items-center mb20">
      <bk-button theme="primary" @click="handleJump('accountAdd')">
        {{t('新增')}}
      </bk-button>
      <div class="input-warp flex-row justify-content-between align-items-center">
        <bk-checkbox v-model="value" class="pr20">
          {{t('精确')}}
        </bk-checkbox>
        <bk-search-select class="bg-white w280" v-model="searchValue" :data="searchData"></bk-search-select>
      </div>
    </div>
    <bk-table
      class="table-layout"
      :data="tableData"
      :pagination="pagination"
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
        <template #default="props">
          <bk-button
            text theme="primary"
            @click="handleJump('accountDetail', props?.data.id)">{{ props?.data.name }}</bk-button>
        </template>
      </bk-table-column>
      <bk-table-column
        :label="t('云厂商')"
        prop="source"
      />
      <bk-table-column
        :label="t('类型')"
        prop="create_time"
      />
      <bk-table-column
        :label="t('负责人')"
        prop="create_time"
      />
      <bk-table-column
        :label="t('余额')"
        prop="create_time"
      />
      <bk-table-column
        :label="t('创建时间')"
        prop="create_time"
      />
      <bk-table-column
        label="备注"
        prop="create_time"
      />
      <bk-table-column
        label="操作"
      >
        <template #default="props">
          <div class="operate-button">
            <bk-button text theme="primary" @click="handleSync">
              同步
            </bk-button>
            <bk-button text theme="primary" @click="handleJump('accountAdd', props?.data.id)">
              编辑
            </bk-button>
            <bk-button text theme="primary" @click="handleDelete(props?.data.id, props?.data.name)">
              删除
            </bk-button>
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
      @confirm="() => test('test')"
    >
      <div>删除之后无法恢复账户信息</div>
    </bk-dialog>

    <bk-dialog
      :is-show="showSyncBox"
      :title="syncTitle"
      :theme="'primary'"
      :quick-close="false"
      @closed="showSyncBox = false"
      @confirm="() => test('test')"
    >
      <div class="sync-dialog-warp">
        <div class="flex-row justify-content-between align-items-center">
          <img class="t-icon" :src="tcloudSrc" />
          <div class="arrow-icon flex-row align-items-center">
            <img class="content" :src="rightArrow" />
          </div>
          <img class="logo-icon" :src="logo" />
        </div>
        <div class="text-center pt20 bg-default">同步中...</div>
      </div>
    </bk-dialog>
  </div>
</template>

<script lang="ts">
import { reactive, toRefs, defineComponent, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import logo from '@/assets/image/logo.png';
import rightArrow from '@/assets/image/right-arrow.png';
import tcloud from '@/assets/image/tcloud.png';

export default defineComponent({
  name: 'AccountManageList',
  setup() {
    const { t } = useI18n();
    const router = useRouter();

    const state = reactive({
      value: false,
      searchValue: '',
      searchData: [
        {
          name: '名称',
          id: 'name',
        }, {
          name: '云厂商',
          id: 'type',
        }, {
          name: '负责人',
          id: 'user',
        },
      ],
      tableData: [
        {
          id: 1,
          name: 'qcloud-for-lol',
          source: 'QQ',
          status: '创建中',
          create_time: '2018-05-25 15:02:241',
          selected: false,
        },
      ],
      pagination: {
        count: 1,
        limit: 10,
      },
      showDeleteBox: false,
      deleteBoxTitle: '',
      syncTitle: '同步',
      showSyncBox: false,
      logo,
      rightArrow,
      tcloudSrc: tcloud,
    });

    onMounted(async () => {
      console.log(122133333);
    });
    onUnmounted(() => {
    });

    const test = (data: any) => {
      console.log(11111, data);
      state.showDeleteBox = false;
    };
    // 跳转页面
    const handleJump = (routerName: string, id?: number) => {
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
      state.deleteBoxTitle = `确认要删除${name}?`;
      state.showDeleteBox = true;
    };

    const handleSync = () => {
      state.showSyncBox = true;
    };

    return {
      ...toRefs(state),
      test,
      handleJump,
      handleDelete,
      handleSync,
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

