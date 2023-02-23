import { defineComponent, onMounted, ref  } from 'vue';
import { Button } from 'bkui-vue';
import { useI18n } from 'vue-i18n';
import { useCommonStore } from '@/store';
import permissions from '@/assets/image/403.png';
import './403.scss';

export default defineComponent({
  setup() {
    const { t } = useI18n();
    const commonStore = useCommonStore();
    const authUrl = ref('');
    onMounted(async () => {
      const params = commonStore.authVerifyParams;
      if (params) {
        const res = await commonStore.authActionUrl(params);    // 列表的权限申请地址
        authUrl.value = res;
        console.log('res.data', res);
      }
    });

    // 打开一个新窗口
    const handlePermissionJump = () => {
      window.open(authUrl.value);
    };
    return {
      handlePermissionJump,
      authUrl,
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
        <Button class="mt20" theme="primary" onClick={this.handlePermissionJump}>{this.t('申请权限')}</Button>
        </div>
      </div>
    );
  },
});
