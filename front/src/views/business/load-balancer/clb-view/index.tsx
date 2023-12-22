import { defineComponent, ref, provide, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import './index.scss';
import allVendors from '@/assets/image/all-vendors.png';
import DynamicTree from '../components/dynamic-tree';
import LoadBalancerDropdownMenu from '../components/clb-dropdown-menu';
import SpecificClbManager from './specific-clb-manager';
import SpecificDomainManager from './specific-domain-manager';
// import Funnel from 'bkui-vue/lib/icon/funnel';
// import AllClbsManager from './all-clbs-manager';

export default defineComponent({
  name: 'LoadBalancerView',
  setup() {
    const route = useRoute();
    const router = useRouter();
    const treeData = ref([]);
    provide('treeData', treeData);
    const isAdvancedSearchShow = ref(false);

    const searchValue = ref('');
    const searchResultCount = ref(0);
    provide('searchResultCount', searchResultCount);

    const toggleExpand = ref(false);
    const handleToggleResultExpand = (isOpen: boolean, shallow?: boolean) => {
      toggleExpand.value = !isOpen;
      treeData.value = treeData.value.map((item) => {
        item.isOpen = isOpen;
        if (!shallow && item.children?.length) {
          item.children = item.children.map((subItem: any) => {
            subItem.isOpen = isOpen;
            return subItem;
          });
        }
        return item;
      });
    };

    const componentMap = {
      clb: <SpecificClbManager/>,
      lisenter: <div>lisenter</div>,
      domain: <SpecificDomainManager/>,
    };

    watch(searchValue, () => {
      searchResultCount.value = 0;
    });

    watch(searchResultCount, (val) => {
      if (val <= 20) {
        handleToggleResultExpand(true, true);
      } else {
        handleToggleResultExpand(false);
      }
    });

    const isAllClbsSelected = ref(true);
    const handleSelectAllClbs = () => {
      isAllClbsSelected.value = !isAllClbsSelected.value;
      if (isAllClbsSelected.value) {
        router.replace({
          query: {
            ...route.query,
            type: 'all',
          },
        });
      }
    };

    const renderComponent = (type: 'clb' | 'listener' | 'domain') => {
      return componentMap[type];
    };

    return () => (
      <div class='clb-view-page'>
        <div class='left-container'>
          <div class='search-wrap'>
            <bk-input v-model={searchValue.value} type='search' clearable placeholder='搜索负载均衡名称、VIP'></bk-input>
          </div>
          <div class='tree-wrap'>
            {
              // eslint-disable-next-line no-nested-ternary
              searchValue.value
                ? (searchResultCount.value ? (
                  <div class='search-result-wrap'>
                    <span class='left-text'>共 {searchResultCount.value} 条搜索结果</span>
                    {
                      toggleExpand.value
                        ? <span class='right-text' onClick={() => handleToggleResultExpand(true)}>全部展开</span>
                        : <span class='right-text' onClick={() => handleToggleResultExpand(false)}>全部收起</span>
                    }
                  </div>
                ) : null)
                : (
                <div class={`all-clbs${isAllClbsSelected.value ? ' selected' : ''}`} onClick={handleSelectAllClbs}>
                  <div class='left-wrap'>
                    <img src={allVendors} alt='' class='prefix-icon' />
                    <span class='text'>全部负载均衡</span>
                  </div>
                  <div class='right-wrap'>
                    <div class='count'>{6654}</div>
                    <LoadBalancerDropdownMenu uuid='all' type='all' />
                  </div>
                </div>
                )
            }
            <DynamicTree searchValue={searchValue.value} />
          </div>
        </div>
        {
          isAdvancedSearchShow.value && <div class='advanced-search'>高级搜索</div>
        }
        <div class='main-container'>
          <div class='common-card-wrap'>
            {
              renderComponent('domain')
            }
          </div>
        </div>
      </div>
    );
  },
});
