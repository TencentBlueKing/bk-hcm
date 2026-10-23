<script lang="ts" setup>
const { VERSION } = window.PROJECT_CONFIG;
</script>

<template>
  <div class="hcm-footer">
    Copyright © {{ new Date().getFullYear() }} Tencent BlueKing. All Rights Reserved. {{ VERSION }}
  </div>
</template>

<style lang="scss" scoped>
.hcm-footer {
  display: flex;
  width: 100%;
  height: 52px;
  align-items: center;
  justify-content: center;
}
</style>
