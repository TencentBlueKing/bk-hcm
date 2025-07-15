<script setup lang="ts"></script>

<template>
  <div class="rule-container">
    <div class="simple-info">
      <span>域名数量：{{ 45 }}</span>
      <span>默认域名：{{ 'xxx' }}</span>
    </div>
    <!-- TODO-CLB：这一层应该只需要遍历域名列表：-->
    <!-- TODO-CLB：下一层需要请求url列表：/api/v1/cloud/bizs/2005000002/vendors/tcloud/listeners/00001jq4/rules/list，注意传入 {
  "op": "and",
  "rules": [
    {
      "field": "domain",
      "op": "eq",
      "value": "doc.tencent.com"
    }
  ]
} -->
    <rule-collapse-panel />
  </div>
</template>

<style scoped lang="scss"></style>
