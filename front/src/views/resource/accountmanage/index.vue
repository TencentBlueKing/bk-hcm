<template>
  <div class="account-warp">
    <div class="operate-warp flex-row justify-content-between align-items-center">
      <bk-button theme="primary">
        新增
      </bk-button>
      <div class="input-warp flex-row justify-content-between align-items-center">
        <bk-checkbox v-model="value" class="pr20">
          精确
        </bk-checkbox>
        <bk-search-select class="bg-white w280" v-model="searchValue" :data="searchData"></bk-search-select>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { reactive, toRefs, defineComponent, onMounted, onUnmounted } from 'vue';
import { useUser } from '@/store/index';

export default defineComponent({
  name: 'AccountManageList',
  setup() {
    const state = reactive({
      value: false,
      searchValue: '',
      searchData: [
        {
          name: '名称',
          id: 'name',
        }, {
          name: '云厂商',
          id: 'type',
        }, {
          name: '负责人',
          id: 'user',
        },
      ],
    });
    onMounted(async () => {
      const user = useUser();
      const res = await user.test();
      console.log(12213, res);
    });

    onUnmounted(() => {
    });

    return {
      ...toRefs(state),
    };
  },
});
</script>

