import { useAccessSourceStore } from '@/stores';
import { Access, AccessType, Platform } from '@/typings';
import { computed, ref, watch } from 'vue';
import { useRoute } from 'vue-router';

export function useBlock(accessType: AccessType) {
  const accessSourceStore = useAccessSourceStore();
  const route = useRoute();
  const isFetching = ref(false);
  const refreshing = ref(false);

  if (accessSourceStore.accessClassifyMap === null) {
    accessSourceStore.fetchAccessSourceClassify(route.params.projectId as string);
  }

  watch(
    () => route.params.projectId,
    async (projectId) => {
      isFetching.value = true;
      await accessSourceStore.fetchAccessSource(projectId as string);
      isFetching.value = false;
    },
    {
      immediate: true,
    },
  );

  const accessList = computed(() => accessSourceStore.accessDataMap[accessType]);
  const classifyList = computed(() => accessSourceStore.accessClassifyMap?.[accessType] ?? []);

  function insertDataAccess() {
    const access = {
      isAuthed: false,
    };
    accessSourceStore.insertAccessSource({
      accessType,
      access,
    });
  }
  function updateDataAccess(index: number, classifyType: Platform) {
    accessSourceStore.updateAccessSource({
      index,
      accessType,
      value: {
        type: accessType,
        data_type: classifyType,
      },
    });
  }

  function getOauthUrlByPlatform(dataType: Platform) {
    return classifyList.value.find((item) => item.type === dataType)?.oauth_url ?? '';
  }

  function removeDataAccess(index: number, access: Access) {
    accessSourceStore.removeAccessSource({
      accessType,
      index,
      access,
    });
  }

  async function refreshAccessData(projectId: string) {
    if (refreshing.value) return;
    refreshing.value = true;
    await accessSourceStore.fetchAccessSource(projectId);
    refreshing.value = false;
  }

  return {
    isFetching,
    refreshing,
    accessList,
    classifyList,
    refreshAccessData,
    insertDataAccess,
    removeDataAccess,
    getOauthUrlByPlatform,
    updateDataAccess,
  };
}
