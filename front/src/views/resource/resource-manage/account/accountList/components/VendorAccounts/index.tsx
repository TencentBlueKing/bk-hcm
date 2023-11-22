import { PropType, defineComponent } from 'vue';
import './index.scss';
import successAccount from '@/assets/image/success-account.png';
import failedAccount from '@/assets/image/failed-account.png';
import { VendorEnum } from '@/common/constant';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';

export default defineComponent({
  props: {
    accounts: {
      required: true,
      type: Array as PropType<
      Array<{
        vendor: VendorEnum;
        count: number;
        name: string;
        icon: any;
        accounts: any[];
        isExpand: boolean;
      }>
      >,
    },
    handleExpand: {
      required: true,
      type: Function as PropType<(vendor: VendorEnum) => void>,
    },
    handleSelect: {
      required: true,
      type: Function as PropType<(id: string) => void>,
    },
  },
  setup(props) {
    const resourceAccountStore = useResourceAccountStore();
    return () => (
      <div class={'vendor-account-list'}>
        {props.accounts.map(({ vendor, count, name, icon, accounts, isExpand }) => (
            <>
              {count > 0 ? (
                <>
                  <div
                    class={'vendor-account-menu'}
                    onClick={() => props.handleExpand(vendor as VendorEnum)}>
                    <i
                      class={
                        isExpand
                          ? 'hcm-icon bkhcm-icon-down-shape vendor-account-menu-dropdown-icon'
                          : 'hcm-icon bkhcm-icon-right-shape vendor-account-menu-dropdown-icon'
                      }></i>
                    <img src={icon} class={'vendor-icon'}></img>
                    <span class={'vendor-account-title'}>{name}</span>
                    <span class={'vendor-account-menu-count'}>{count}</span>
                  </div>
                  {isExpand
                    ? accounts.map(({ sync_status, name, id }) => (
                        <div
                          class={`vendor-account-menu-item ${resourceAccountStore.resourceAccount?.id === id ? 'actived-vendor-account-menu-item' : ''}`}
                          onClick={() => props.handleSelect(id)}
                        >
                          <img
                            src={
                              sync_status === 'sync_success'
                                ? successAccount
                                : failedAccount
                            }
                            class={'vendor-icon'}></img>
                          <span class={'vendor-account-menu-item-text'}>
                            {name.length > 22
                              ? (
                                <span v-bk-tooltips={{
                                  content: name,
                                  placement: 'right',
                                }}
                              >
                                  {`${name.substring(0, 22)}..`}
                                </span>
                              )
                              : name}
                          </span>
                        </div>
                    ))
                    : null}
                </>
              ) : null}
            </>
        ))}
      </div>
    );
  },
});
