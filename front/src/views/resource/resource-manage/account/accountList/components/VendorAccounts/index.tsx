import { PropType, defineComponent, onMounted, ref, watch } from 'vue';
import './index.scss';
import successAccount from '@/assets/image/success-account.png';
import failedAccount from '@/assets/image/failed-account.png';
import { VendorEnum } from '@/common/constant';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import { useRoute } from 'vue-router';
import { useResourceStore } from '@/store';

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
        hasNext: boolean;
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
    checkIsExpand: {
      required: true,
      type: Function as PropType<(vendor: VendorEnum) => boolean>,
    },
    getVendorAccountList: {
      require: true,
      type: Function as PropType<(vendor: VendorEnum) => void>,
    },
  },
  setup(props) {
    const resourceAccountStore = useResourceAccountStore();
    const resourceStore = useResourceStore();
    const route = useRoute();

    const loadingRef = ref([]);

    watch(() => route.query.accountId, (newVal, oldVal) => {
      if (!oldVal && newVal) {
        // 如果是从全部账号下进入详情页, 此时点击账号id是没有 oldVal 的. 如果 newVal 对应的厂商在账号列表中没有展开, 那么将之展开即可.
        const { vendorOfCurrentResource } = resourceStore;
        if (props.checkIsExpand(vendorOfCurrentResource)) return;
        props.handleExpand(vendorOfCurrentResource);
      }
    });

    onMounted(() => {
      const observer = new IntersectionObserver((entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            props.getVendorAccountList(entry.target.dataset.vendor);
          }
        });
      });
      loadingRef.value.forEach((vnode) => {
        observer.observe(vnode.$el);
      });
    });

    return () => (
      <div class={'vendor-account-list'}>
        {props.accounts.map(({ vendor, count, name, icon, accounts, isExpand, hasNext }) => (
          count > 0 && <div class='vendor-wrap'>
              <div class={`vendor-item-wrap${isExpand ? ' sticky' : ''}`} onClick={() => props.handleExpand(vendor)}>
                <i class={`icon bk-icon vendor-account-menu-dropdown-icon${isExpand ? ' icon-down-shape' : ' icon-right-shape'}`}></i>
                {icon}
                <div class='vendor-title'>{name}</div>
                <div class='vendor-account-count'>{count}</div>
              </div>
              <div class={`account-list-wrap${isExpand ? ' expand' : ''}`}>
                {
                  accounts.map(({ sync_status, name, id }) => (
                    <div class={`account-item${resourceAccountStore.resourceAccount?.id === id ? ' active' : ''}`} key={id} onClick={() => props.handleSelect(id)}>
                      <img src={sync_status === 'sync_success' ? successAccount : failedAccount} alt='' class='sync-status-icon' />
                      <span class='account-text'>
                        {
                          name.length > 22 ? (
                            <span v-bk-tooltips={{ content: name, placement: 'right' }}>{`${name.substring(0, 22)}..`}</span>
                          ) : (name)
                        }
                      </span>
                    </div>
                  ))
                }
                {
                  hasNext && <bk-loading ref={(ref: any) => loadingRef.value.push(ref)} data-vendor={vendor} size="small" loading><div style="width: 100%; height: 36px" /></bk-loading>
                }
              </div>
            </div>
        ))}
      </div>
    );
  },
});
