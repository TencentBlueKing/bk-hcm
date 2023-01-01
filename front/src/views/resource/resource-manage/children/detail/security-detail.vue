<script lang="ts" setup>
import DetailHeader from '../../common/header/detail-header';
import DetailTab from '../../common/tab/detail-tab';
import SecurityInfo from '../components/security/security-info.vue';
import SecurityRelate from '../components/security/security-relate.vue';
import SecurityRule from '../components/security/security-rule.vue';

import {
  useI18n,
} from 'vue-i18n';

import {
  ref,
} from 'vue';


import {
  useRoute,
} from 'vue-router';

const {
  t,
} = useI18n();

const route = useRoute();

const activeTab = ref(route.query?.activeTab);

const tabs = [
  {
    name: t('基本信息'),
    value: 'detail',
  },
  {
    name: t('关联实例'),
    value: 'relate',
  },
  {
    name: t('安全组规则'),
    value: 'rule',
  },
];

</script>

<template>
  <detail-header>
    {{t('安全组')}}：ID（xxx）
  </detail-header>

  <detail-tab
    :tabs="tabs"
    :active="activeTab"
  >
    <template #default="type">
      <security-info v-if="type === 'detail'" />
      <security-relate v-if="type === 'relate'" />
      <security-rule v-if="type === 'rule'" />
    </template>
  </detail-tab>
</template>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}
.w60 {
  width: 60px;
}
</style>
