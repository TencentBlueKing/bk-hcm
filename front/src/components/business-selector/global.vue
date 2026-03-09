<script lang="ts" setup>
import { ref, watch } from 'vue';
// @ts-expect-error
import AppSelect from '@blueking/app-select';
import '@blueking/app-select/dist/style.css';
import { useBusinessGlobalStore, type IBusinessItem } from '@/store/business-global';
import { useBusinessFavorite } from '@/hooks/use-business-favorite';

const props = defineProps<{
  value: number;
}>();

const emit = defineEmits<{
  change: [id: number];
}>();

const { BK_CMDB_CREATE_BIZ_URL } = window.PROJECT_CONFIG;

const businessGlobalStore = useBusinessGlobalStore();

const businessFavorite = useBusinessFavorite();

const businessList = ref<IBusinessItem[]>(businessGlobalStore.businessAuthorizedList.slice());
const selectedValue = ref({ id: props.value, name: '' });

const createDialogShow = ref(false);

const sortBusinessList = () => {
  // 按收藏排序
  businessList.value.sort((biz1, biz2) => {
    return +businessFavorite.favoriteSet.value.has(biz2.id) - +businessFavorite.favoriteSet.value.has(biz1.id);
  });
};

watch(
  () => props.value,
  async (newValue) => {
    selectedValue.value.id = newValue;
    if (!businessGlobalStore.businessAuthorizedList.some((biz) => biz.id === newValue)) return;
    await businessFavorite.get(newValue);
    sortBusinessList();
  },
);

const handleChange = async (val: IBusinessItem) => {
  if (val.id !== selectedValue.value.id) {
    emit('change', val.id);
  }
  selectedValue.value = val;
};

const handleCollect = async (event: MouseEvent, id: number, collected: boolean) => {
  event.stopPropagation();
  if (collected) {
    await businessFavorite.remove(id);
  } else {
    await businessFavorite.add(id);
  }
  sortBusinessList();
};

const handleCreate = () => {
  window.open(BK_CMDB_CREATE_BIZ_URL, '_blank');
};
</script>

<template>
  <AppSelect
    class="business-selector-global"
    theme="dark"
    :clearable="false"
    :data="businessList"
    :value="selectedValue"
    @change="handleChange"
  >
    <template #default="{ data }">
      <div class="select-item">
        <div class="item-name">{{ data.name }}</div>
        <div class="item-id">({{ data.id }})</div>
        <div class="item-collect">
          <i
            :class="[
              'hcm-icon',
              businessFavorite.favoriteSet.value.has(data.id) ? 'bkhcm-icon-collect' : 'bkhcm-icon-not-favorited',
            ]"
            @click="(event) => handleCollect(event, data.id, businessFavorite.favoriteSet.value.has(data.id))"
          />
        </div>
      </div>
    </template>
    <template #append>
      <div class="select-append">
        <bk-button class="create-button" text size="small" @click="createDialogShow = true">
          <i class="hcm-icon bkhcm-icon-plus-circle create-icon" />
          新建业务
        </bk-button>
      </div>
    </template>
  </AppSelect>

  <bk-dialog v-model:is-show="createDialogShow" dialog-type="show" width="520">
    <bk-exception
      type="building"
      scene="part"
      title="新建业务参考以下指引"
      description="业务是蓝鲸配置平台的管理空间，可以满足不同团队，不同项目的资源隔离管理需求"
    >
      <bk-button theme="primary" text @click="handleCreate">新建业务</bk-button>
    </bk-exception>
  </bk-dialog>
</template>

<style lang="scss" scoped>
.business-selector-global {
  &[data-theme='dark'] {
    // background: #434b5f;
  }
}

.select-item {
  display: flex;
  align-items: center;
  width: 100%;

  .item-name {
    color: #c4c6cc;
  }

  .item-id {
    color: #979ba5;
    margin-left: 4px;
  }

  .item-collect {
    margin-left: auto;

    .bkhcm-icon-collect {
      color: #cc933a;
      font-size: 14px;
    }
  }
}

.select-append {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;

  .create-button {
    color: #c4c6cc;

    .create-icon {
      margin-right: 4px;
    }
  }
}
</style>
