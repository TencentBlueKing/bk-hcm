<script setup lang="ts">
import Panel from './panel';
import { AccountLevelEnum } from './constants';
</script>

<template>
  <div class="account-list-page">
    <div class="account-list-content">
      <Panel :account-level="AccountLevelEnum.SecondLevel" />
    </div>
  </div>
</template>

<style lang="scss" scoped src="./account-list.scss"></style>
