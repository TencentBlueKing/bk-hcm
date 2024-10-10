<script setup lang="ts">
import { onMounted, ref } from 'vue';
import ReleaseNote from '@blueking/release-note';
import '@blueking/release-note/vue3/vue3.css';
import http from '@/http';
import { localStorageActions, timeFromNow } from '@/common/util';

interface ReleaseNote {
  version: string;
  time: string;
  is_current: boolean;
}
type ReleaseNoteList = ReleaseNote[];

const showSyncReleaseNote = ref(false);
const releaseNoteDetail = ref('');
const syncReleaseList = ref<ReleaseNoteList>([]);
const syncReleaseNoteLoading = ref(false);
const current = ref('');

const handleSelectRelease = async (changeLog: ReleaseNote) => {
  syncReleaseNoteLoading.value = true;
  try {
    const res = await http.get(`/api/v1/web/changelog/${changeLog.version}`);
    releaseNoteDetail.value = res.data;
  } finally {
    syncReleaseNoteLoading.value = false;
  }
};

onMounted(() => {
  const getReleaseNoteList = async () => {
    const { data }: { data: ReleaseNoteList } = await http.get('/api/v1/web/changelogs');

    syncReleaseList.value = data?.sort((a, b) => timeFromNow(a.time) - timeFromNow(b.time)) || [];
    // 最新版本，若没有，取列表第一个
    current.value = data?.find((item) => item.is_current)?.version || data?.[0].version || '';

    // 判断上次访问系统的版本号与当前最新版本是否一致，不一致则默认弹出一次版本日志
    if (
      syncReleaseList.value.length &&
      localStorageActions.get('bk-hcm-release-note-version', (value) => value) !== current.value
    ) {
      // 更新版本日志的version
      localStorageActions.set('bk-hcm-release-note-version', current.value);
      showSyncReleaseNote.value = true;
    }
  };

  // 加载版本日志列表
  getReleaseNoteList();
});
</script>

<template>
  <aside class="header-release-note">
    <i class="hcm-icon bkhcm-icon-release-note" @click="showSyncReleaseNote = true"></i>
  </aside>
  <ReleaseNote
    v-model:show="showSyncReleaseNote"
    title-key="version"
    sub-title-key="time"
    :current="current"
    :detail="releaseNoteDetail"
    :list="syncReleaseList"
    :loading="syncReleaseNoteLoading"
    @selected="handleSelectRelease"
  />
</template>

<style scoped lang="scss">
.header-release-note {
  margin-right: 25px;
  font-size: 16px;

  &:hover {
    color: #fff;
    cursor: pointer;
  }
}
</style>
