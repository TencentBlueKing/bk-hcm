<template>
  <div class="basic-resource">
    <section class="card-group">
      <h2 class="group-title">资源申请</h2>
      <div class="group-content">
        <ResourceCard
          :list="appList"
        />
      </div>
    </section>
    <section class="card-group">
      <h2 class="group-title">账号录入</h2>
      <div class="group-content">
        <ResourceCard
          :list="accountList"
        />
      </div>
    </section>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue';
import ResourceCard from '../resource-card/index.vue';

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
    const accountList = ref([{
      name: '云账号录入',
      id: 1,
      btnText: '立即申请',
      routeName: 'applyAccount',
    }]);

    return {
      appList,
      accountList,
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
    padding: .5em 0;
    border-bottom: 1px solid rgba(0, 0, 0, .15);
  }
  .group-content {
    padding: 12px;
  }
}
</style>
