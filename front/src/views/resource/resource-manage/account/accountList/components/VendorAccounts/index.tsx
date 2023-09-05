import { PropType, defineComponent } from 'vue';
import './index.scss';
import successAccount from '@/assets/image/success-account.png';
import failedAccount from '@/assets/image/failed-account.png';
import { VendorEnum } from '@/common/constant';

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
  },
  setup(props) {
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
                            {name.length > 12
                              ? `${name.substring(0, 10)}..`
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
