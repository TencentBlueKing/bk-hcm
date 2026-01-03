<script setup lang="ts">
import { ArrowsRight } from 'bkui-vue/lib/icon';
import routerAction from '@/router/utils/action';
import { GLOBAL_BIZS_KEY } from '@/common/constant';

const props = defineProps<{ bkBizId: number }>();

const handleGotoAddPlan = () => {
  routerAction.open({ path: '/business/resource-plan/add', query: { [GLOBAL_BIZS_KEY]: props.bkBizId } });
};
</script>

<template>
  <bk-alert theme="warning" class="device-type-plan-empty-alert">
    <template #title>当前预测数据为空，无可申领机型，处理建议如下</template>
    <template #default>
      <p class="desc-item">
        1.&nbsp;
        <b>增加预测：</b>
        如需本周申领，到货日期选择本周
        <bk-button text theme="primary" class="link-button" @click="handleGotoAddPlan">
          去增加预测
          <arrows-right class="icon-right" />
        </bk-button>
      </p>
      <p class="desc-item">
        2.&nbsp;
        <b>切换需求类型：</b>
        可优先选择「滚服项目」和「小额绿通」，资源限量供给，额度有限
      </p>
    </template>
  </bk-alert>
</template>

<style scoped lang="scss">
.device-type-plan-empty-alert {
  border: none;
  color: #4d4f56;
  line-height: 20px;

  .desc-item {
    display: flex;
    align-items: center;
  }

  .link-button {
    margin-left: 4px;
  }

  .icon-right {
    font-size: 22px;
  }
}
</style>
