import { defineComponent } from 'vue';
import './index.scss';
import tcloudVendor from '@/assets/image/vendor-tcloud.png';
// import awsVendor from '@/assets/image/vendor-aws.png';
// import azureVendor from '@/assets/image/vendor-azure.png';
// import gcpVendor from '@/assets/image/vendor-gcp.png';
// import huaweiVendor from '@/assets/image/vendor-huawei.png';
import successAccount from '@/assets/image/success-account.png';

export default defineComponent({
  setup() {
    return () => (
      <div class={'vendor-account-list'}>
        <div class={'vendor-account-menu'}>
          <icon class={'icon bk-icon icon-down-shape vendor-account-menu-dropdown-icon'}></icon>
          <img src={tcloudVendor} class={'vendor-icon'}></img>
          <span class={'vendor-account-title'}>
            腾讯云
          </span>
          <span class={'vendor-account-menu-count'}>
            232
          </span>
        </div>
        <div class={'vendor-account-menu-item'}>
          <img src={successAccount} class={'account-icon'}></img>
          云账号1
        </div>
        <div class={'vendor-account-menu-item'}>
        <img src={successAccount}></img>
          云账号2
        </div>
      </div>
    );
  },
});
