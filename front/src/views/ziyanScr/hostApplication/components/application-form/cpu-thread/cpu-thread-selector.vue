<script setup lang="ts">
import { CpuThreadSwitch } from './use-cpu-thread';

const model = defineModel<number>();
</script>

<template>
  <bk-form-item>
    <bk-radio-group v-model="model">
      <bk-radio-button :label="CpuThreadSwitch.ENABLED">开启</bk-radio-button>
      <bk-radio-button :label="CpuThreadSwitch.DISABLED">关闭</bk-radio-button>
    </bk-radio-group>
    <div class="tips">
      <span v-if="model === CpuThreadSwitch.ENABLED || model === undefined">
        默认开启超线程，如有特殊使用场景，可关闭超线程
      </span>
      <span class="disabled-text" v-else-if="model === CpuThreadSwitch.DISABLED">
        <em class="text-danger">已选择关闭</em>
        ，仅业务有特殊场景选择此功能
      </span>
      。参考文档：
      <a href="https://cloud.tencent.com/document/product/213/103798" target="_blank">
        https://cloud.tencent.com/document/product/213/103798
      </a>
    </div>
  </bk-form-item>
</template>

<style scoped lang="scss">
.tips {
  display: flex;
  align-items: center;
  font-size: 12px;
  color: #979ba5;

  .disabled-text {
    display: inline-flex;
    align-items: center;
  }

  .text-danger {
    color: $danger-color;
    font-style: normal;
  }

  a {
    color: #3a84ff;
  }
}
</style>
