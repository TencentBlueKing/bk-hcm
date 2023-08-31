import { defineComponent, onMounted, reactive } from 'vue';
import './index.scss';
import tcloudVendor from '@/assets/image/vendor-tcloud.png';
import awsVendor from '@/assets/image/vendor-aws.png';
import azureVendor from '@/assets/image/vendor-azure.png';
import gcpVendor from '@/assets/image/vendor-gcp.png';
import huaweiVendor from '@/assets/image/vendor-huawei.png';
import successAccount from '@/assets/image/success-account.png';
import failedAccount from '@/assets/image/failed-account.png';
import { getAllAccounts } from './getAllAcounts';
import { VendorEnum } from '@/common/constant';

export default defineComponent({
  setup() {
    const Vendors = reactive({
      [VendorEnum.TCLOUD]: {
        count: 0,
        name: '腾讯云',
        icon: tcloudVendor,
        accounts: [],
        isExpand: false,
      },
      [VendorEnum.AWS]: {
        count: 0,
        name: '亚马逊云',
        icon: awsVendor,
        accounts: [],
        isExpand: false,
      },
      [VendorEnum.AZURE]: {
        count: 0,
        name: '微软云',
        icon: azureVendor,
        accounts: [],
        isExpand: false,
      },
      [VendorEnum.GCP]: {
        count: 0,
        name: '谷歌云',
        icon: gcpVendor,
        accounts: [],
        isExpand: false,
      },
      [VendorEnum.HUAWEI]: {
        count: 0,
        name: '华为云',
        icon: huaweiVendor,
        accounts: [],
        isExpand: false,
      },
    });

    onMounted(async () => {
      const resArr = await getAllAccounts();
      for (const { data } of resArr) {
        console.log(data.count, data?.details?.[0]?.vendor);
        if (data?.details?.length) {
          const vendor = data?.details?.[0]?.vendor;
          if (Vendors[vendor]) {
            Vendors[vendor].count = data.details.length;
            Vendors[vendor].accounts = data.details;
          }
        }
      }
    });

    const handleExpand = (vendor: VendorEnum) => {
      Vendors[vendor].isExpand = !Vendors[vendor].isExpand;
    };

    return () => (
      <div class={'vendor-account-list'}>
        {Object.entries(Vendors).map(([vendor, { count, name, icon, accounts, isExpand }]) => (
            <>
              <div
                class={'vendor-account-menu'}
                onClick={() => handleExpand(vendor as VendorEnum)}
              >
                <i
                  class={
                    isExpand
                      ? 'icon bk-icon icon-down-shape vendor-account-menu-dropdown-icon'
                      : 'icon bk-icon icon-right-shape vendor-account-menu-dropdown-icon'
                  }></i>
                <img src={icon} class={'vendor-icon'}></img>
                <span class={'vendor-account-title'}>{name}</span>
                <span class={'vendor-account-menu-count'}>{count}</span>
              </div>
              {isExpand
                ? accounts.map(({ sync_status, name }) => (
                    <div class={'vendor-account-menu-item'}>
                      <img
                        src={
                          sync_status === 'sync_success'
                            ? successAccount
                            : failedAccount
                        }
                        class={'vendor-icon'}></img>
                      <span class={'vendor-account-menu-item-text'}>
                        {name}
                      </span>
                    </div>
                ))
                : null}
            </>
        ))}
      </div>
    );
  },
});
