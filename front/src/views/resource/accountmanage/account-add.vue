<template>
  <div class="account-add-warp">
    <bk-form
      ref="formRef"
      class="example"
      :model="formModel"
      :rules="rules">
      <bk-form-item
        label="姓名"
        property="name"
      >
        <bk-input
          v-model="formModel.name"
          placeholder="请输入"
          clearable
        />
      </bk-form-item>
    </bk-form>
  </div>
</template>
<script lang="ts">
import { reactive, toRefs, defineComponent, onMounted, onUnmounted } from 'vue';
export default defineComponent({
  name: 'AccountManageAdd',
  setup() {
    const state = reactive({
      formModel: { name: '' },
      rules: {},
    });
    onMounted(async () => {
      console.log(11112334);
    });

    onUnmounted(() => {
    });
    return {
      ...toRefs(state),
    };
  },
});
</script>
