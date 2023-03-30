import { defineComponent, onMounted, ref  } from 'vue';
import { Button } from 'bkui-vue';
import { useI18n } from 'vue-i18n';
import { useCommonStore } from '@/store';
import { useRoute } from 'vue-router';

import permissions from '@/assets/image/403.png';
import './403.scss';

export default defineComponent({
  setup() {
    const { t } = useI18n();
    const route = useRoute();
    const commonStore = useCommonStore();
    const authUrl = ref('');
    const urlLoading = ref<boolean>(false);

    onMounted(async () => {
      let urlKey: any = route.params.id;
      if (urlKey?.includes('iaas_resource_operate')) {
        urlKey = 'iaas_resource_operate';
      } else if (urlKey?.includes('resource_find')) {
        urlKey = 'resource_find';
      }
      const { authVerifyData } = commonStore;
      if (authVerifyData) {       // 权限矩阵数据
        const params = authVerifyData.urlParams[urlKey];    // 获取权限链接需要的参数
        if (params) {
          urlLoading.value = true;
          const res = await commonStore.authActionUrl(params);    // 列表的权限申请地址
          authUrl.value = res;
          urlLoading.value = false;
          console.log('res.data', res);
        }
      }
    });

    // 打开一个新窗口
    const handlePermissionJump = () => {
      window.open(authUrl.value);
    };
    return {
      handlePermissionJump,
      authUrl,
      urlLoading,
      t,
    };
  },

  render() {
    return (
      <div class="forbid-layout">
        <img src={permissions} alt="403" />
        <h2>{this.t('抱歉，您暂无该功能的权限')}</h2>
        <p class="mt10">{this.t('您还没有该功能的权限，可以点击下方的"申请功能权限"获得权限')}</p>
        <div>
        <Button class="mt20" theme="primary"
        loading={this.urlLoading}
        onClick={this.handlePermissionJump}>{this.t('申请权限')}</Button>
        </div>
      </div>
    );
  },
});
