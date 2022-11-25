<template>
  <div class="account-warp">
    <div class="operate-warp flex-row justify-content-between align-items-center mb20">
      <bk-button theme="primary" @click="toAddAccount">
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
        prop="ip"
      >
        <template #default="props">
          {{ props?.data.ip }}
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
            <bk-button text theme="primary" @click="test(props)">
              同步
            </bk-button>
            <bk-button text theme="primary" @click="test(props)">
              编辑
            </bk-button>
            <bk-button text theme="primary" @click="test(props)">
              删除
            </bk-button>
          </div>
        </template>
      </bk-table-column>
    </bk-table>
  </div>
</template>

<script lang="ts">
import { reactive, toRefs, defineComponent, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';

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
          ip: '192.168.0.1-2018-05-25 15:02:241',
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
    });

    onMounted(async () => {
      console.log(122133333);
    });
    onUnmounted(() => {
    });

    const test = (data: any) => {
      console.log(11111, data);
    };
    // 跳转到新增资源账号
    const toAddAccount = () => {
      router.push({ name: 'accountAdd' });
    };

    return {
      ...toRefs(state),
      toAddAccount,
      test,
      t,
    };
  },
});
</script>

