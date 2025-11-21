<script setup lang="ts">
import { ref, computed } from 'vue';
import Search from './search';
import Table from './table';
import { useVerify } from '@/hooks';
import ErrorPage from '@/views/error-pages/403';
import { InfoLine } from 'bkui-vue/lib/icon';
import { useDissolveQuotaStore } from '@/store/dissolve/quota';
import { Message } from 'bkui-vue';
import { AUTH_FIND_DISSOLVE, AUTH_UPDATE_DISSOLVE } from '@/constants/auth-symbols';

const dissolveStore = useDissolveQuotaStore();

const { authVerifyData } = useVerify();
const moduleNames = ref<string[]>([]);

// 在search中，模块名是 `which_stages__module_name` 的格式，这里需要提取出 module_name
const moduleNameList = computed(() => {
  const names = moduleNames.value.map((item) => item.split('__')[1]).filter(Boolean);
  return [...new Set(names)];
});

const isShowConfig = ref(false);
const formData = ref({
  host_apply_time: '',
  approval_limit: '',
});

const handleConfirm = () => {
  dissolveStore
    .upsertDissolveConfig(formData.value)
    .then(() => {
      Message({
        theme: 'success',
        message: '提交成功',
      });
      isShowConfig.value = false;
    })
    .catch((e) => {
      Message({
        theme: 'error',
        message: `提交失败 ${e}`,
      });
    });
};

const handleShowConfig = async () => {
  formData.value = await dissolveStore.getDissolveConfig();
  isShowConfig.value = true;
};
</script>

<template>
  <ErrorPage
    v-if="!authVerifyData.permissionAction.service_resource_dissolve_find"
    url-key-id="biz_ziyan_resource_dissolve"
  />

  <section v-else class="home">
    <Search v-model:module-names="moduleNames"></Search>
    <Table :module-names="moduleNameList"></Table>

    <hcm-auth v-slot="{ noPerm }" :sign="[{ type: AUTH_FIND_DISSOLVE }, { type: AUTH_UPDATE_DISSOLVE }]">
      <template v-if="!noPerm">
        <Teleport defer to="#breadcrumbExtra">
          <bk-button theme="primary" @click="handleShowConfig">裁撤配置</bk-button>
        </Teleport>

        <bk-dialog v-model:is-show="isShowConfig" title="裁撤配置" width="600">
          <bk-loading title="提交中" :loading="dissolveStore.upsertDissolveConfigLoading">
            <div class="dissolve-config mt25">
              <p class="tips mb15">
                <InfoLine fill="#3a84ff" width="20" height="20" />
                <span class="ml6">配置机房裁撤的全局参数</span>
              </p>

              <bk-form class="form" form-type="vertical" :model="formData">
                <bk-form-item property="host_apply_time" required>
                  <template #label>
                    <p class="mb6" style="display: inline-block">裁撤开始时间</p>
                  </template>
                  <hcm-form-datetime
                    v-model="formData.host_apply_time"
                    style="width: 100%"
                    append-to-body
                    clearable
                    placeholder="选择裁撤开始时间"
                    font-size="medium"
                    type="datetime"
                  />
                </bk-form-item>

                <bk-form-item property="approval_limit" required>
                  <template #label>
                    <p class="tips-2 mb6">
                      <span>自动审批配置</span>
                      <InfoLine
                        class="ml5"
                        fill="#c7c7c7"
                        width="16"
                        height="16"
                        v-bk-tooltips="{ content: '申领核数达到裁撤核数的比例，则触发人工审核' }"
                      />
                    </p>
                  </template>
                  <bk-input
                    v-model="formData.approval_limit"
                    style="font-size: 14px"
                    type="number"
                    prefix="触发审批值"
                    suffix="%"
                    :precision="2"
                    :max="100"
                    :min="0"
                  />
                </bk-form-item>
              </bk-form>
            </div>
          </bk-loading>

          <template #footer>
            <div class="btns">
              <bk-button theme="primary" class="mr10" @click="handleConfirm">确定</bk-button>
              <bk-button @click="isShowConfig = false">取消</bk-button>
            </div>
          </template>
        </bk-dialog>
      </template>
    </hcm-auth>
  </section>
</template>

<style lang="scss" scoped>
.home {
  padding: 24px;
}

.tips {
  display: flex;
  align-items: center;

  span {
    color: #c4c6cc;
  }
}

.tips-2 {
  display: inline-flex;
  align-items: center;
}
</style>
