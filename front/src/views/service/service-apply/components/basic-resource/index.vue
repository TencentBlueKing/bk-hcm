<template>
  <div class="basic-resource">
    <section class="card-group">
      <h2 class="group-title">资源申请</h2>
      <div class="group-content">
        <ResourceCard :list="appList" :auth-verify-data="authVerifyData" @handle-auth-click="handleClick" />
      </div>
    </section>
    <section class="card-group">
      <h2 class="group-title">账号录入</h2>
      <div class="group-content">
        <ResourceCard :list="accountList" :auth-verify-data="authVerifyData" @handle-auth-click="handleClick" />
      </div>
    </section>
    <permission-dialog
      v-model:is-show="showPermissionDialog"
      :params="permissionParams"
      @cancel="handlePermissionDialog"
      @confirm="handlePermissionConfirm"
    ></permission-dialog>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue';
import ResourceCard from '../resource-card/index.vue';
import { useVerify } from '@/hooks';

export default defineComponent({
  name: 'BasicResource',
  components: {
    ResourceCard,
  },
  setup() {
    const appList = [
      {
        name: '主机',
        btnText: '立即申请',
        routeName: 'applyCvm',
      },
      {
        name: 'VPC',
        btnText: '立即申请',
        routeName: 'applyVPC',
      },
      {
        name: '云硬盘',
        btnText: '立即申请',
        routeName: 'applyDisk',
      },
    ];
    const accountList = ref([
      {
        name: '云账号录入',
        id: 1,
        btnText: '立即录入',
        routeName: 'applyAccount',
      },
    ]);

    const handleClick = (value: string) => {
      console.log('1111', value);
      handleAuth(value);
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
      appList,
      accountList,
      showPermissionDialog,
      handlePermissionConfirm,
      handlePermissionDialog,
      handleAuth,
      permissionParams,
      authVerifyData,
      handleClick,
    };
  },
});
</script>

<style lang="scss" scoped>
.card-group {
  margin-bottom: 20px;
  .group-title {
    font-size: 16px;
    font-weight: normal;
    padding: 0.5em 0;
    border-bottom: 1px solid rgba(0, 0, 0, 0.15);
  }
  .group-content {
    padding: 12px;
  }
}
</style>
