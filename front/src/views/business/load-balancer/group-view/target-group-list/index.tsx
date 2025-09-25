import { ComputedRef, defineComponent, inject, onMounted, onUnmounted, ref, watch, nextTick } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useLoadBalancerStore, useAccountStore, useBusinessStore, ITargetGroupDetail } from '@/store';
import useMoreActionDropdown from '@/hooks/useMoreActionDropdown';
import { useSingleList } from '@/hooks/useSingleList';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { QueryRuleOPEnum } from '@/typings';
import { MENU_BUSINESS_TARGET_GROUP_DETAILS, MENU_BUSINESS_TARGET_GROUP_OVERVIEW } from '@/constants/menu-symbol';
import { IAuthSign } from '@/common/auth-service';
import bus from '@/common/bus';

import { Input, Message, OverflowTitle } from 'bkui-vue';
import allIcon from '@/assets/image/all-lb.svg';
import mubiaoIcon from '@/assets/image/mubiao.svg';
import Confirm from '@/components/confirm';
import './index.scss';

export default defineComponent({
  name: 'TargetGroupList',
  setup() {
    const clbCreateAuthSign = inject<ComputedRef<IAuthSign | IAuthSign[]>>('clbCreateAuthSign');
    const clbOperationAuthSign = inject<ComputedRef<IAuthSign | IAuthSign[]>>('clbOperationAuthSign');
    // use hooks
    const { getBusinessApiPath } = useWhereAmI();
    const router = useRouter();
    const route = useRoute();
    // use stores
    const loadBalancerStore = useLoadBalancerStore();
    const accountStore = useAccountStore();
    const businessStore = useBusinessStore();

    // 搜索相关
    const searchValue = ref('');
    const targetGroupListRef = ref(null);
    const activeTargetGroupId = ref(''); // 当前选中的目标组id
    const allTargetGroupsItem = { type: 'all', isDropdownListShow: false }; // 全部目标组item

    // 获取目标组列表
    const rules = ref([]);
    const { dataList, pagination, handleScrollEnd, handleRefresh, isDataLoad } = useSingleList({
      url: `/api/v1/cloud/${getBusinessApiPath()}target_groups/list`,
      rules: () => rules.value,
      immediate: !loadBalancerStore.tgSearchTarget,
      rollRequestConfig: { enabled: true, limit: 500 },
      success: () => handleListLoaded(),
    });

    const handleListLoaded = () => {
      nextTick(() => {
        const target = targetGroupListRef.value.getElementsByClassName('selected')[0];
        target?.parentNode?.scrollTo({
          top: target?.offsetTop - 200,
          behavior: 'auto',
        });
      });
    };

    // handler - 切换目标组
    const handleTypeChange = (targetGroup?: ITargetGroupDetail) => {
      const { id = '', vendor } = targetGroup || {};
      // // 如果两个目标组id相同，则不做切换
      if (id === activeTargetGroupId.value) return;
      // 设置当前选中的目标组id
      activeTargetGroupId.value = id;
      loadBalancerStore.setTargetGroupId(id);
      // 导航
      router.push({
        name: id ? MENU_BUSINESS_TARGET_GROUP_DETAILS : MENU_BUSINESS_TARGET_GROUP_OVERVIEW,
        query: { ...route.query, type: id ? route.query.type : undefined, vendor },
        params: { id: id || undefined },
      });
    };

    // 删除目标组
    const handleDeleteTargetGroup = (node: any) => {
      const { id, name } = node;
      Confirm('请确定删除目标组', `将删除目标组【${name}】`, async () => {
        await businessStore.deleteTargetGroups({ bk_biz_id: accountStore.bizs, ids: [id] });
        Message({ message: '删除成功', theme: 'success' });
        // 重新拉取目标组list
        handleRefresh();
        // 跳转至全部目标组下
        handleTypeChange();
      });
    };

    // more-action
    const typeMenuMap = {
      all: [{ label: '新增目标组', handler: () => bus.$emit('addTargetGroup') }],
      specific: [{ label: '删除', handler: handleDeleteTargetGroup }],
    };
    const { showDropdownList, currentPopBoundaryNodeKey } = useMoreActionDropdown(typeMenuMap);

    const handleSearch = (val: string) => {
      // 设置搜索条件
      if (val) {
        rules.value = [{ field: 'name', op: QueryRuleOPEnum.CIS, value: val }];
      } else {
        rules.value = [];
      }
      // 拉取搜索结果
      handleRefresh();
    };

    watch(
      () => route.params.id,
      (val) => {
        // 高亮状态保持
        if (!val) activeTargetGroupId.value = '';
        else activeTargetGroupId.value = val as string;
      },
      { immediate: true },
    );

    watch(
      () => loadBalancerStore.tgSearchTarget,
      (val) => {
        if (!val) return;
        searchValue.value = val;
        handleSearch(val);
      },
      {
        immediate: true,
      },
    );

    onMounted(() => {
      bus.$on('refreshTargetGroupList', handleRefresh);
    });

    const handleScroll = ({ target }: Event) => {
      const { scrollTop, scrollHeight, clientHeight } = target as HTMLDivElement;
      if (!isDataLoad.value && scrollTop + clientHeight >= scrollHeight) {
        handleScrollEnd();
      }
    };

    onUnmounted(() => {
      bus.$off('refreshTargetGroupList');
      loadBalancerStore.setTgSearchTarget('');
    });

    return () => (
      <div class='target-group-list' ref={targetGroupListRef}>
        <div class='search-wrap'>
          <Input
            v-model={searchValue.value}
            type='search'
            clearable
            placeholder='搜索目标组'
            onChange={handleSearch}
            onClear={() => loadBalancerStore.setTgSearchTarget('')}
          />
        </div>
        <div class='group-list-wrap'>
          <div
            class={`all-groups-wrap${!activeTargetGroupId.value ? ' selected' : ''}`}
            onClick={() => handleTypeChange()}>
            <div class='base-info'>
              <img src={allIcon} alt='' class='prefix-icon' />
              <span class='text'>全部目标组</span>
            </div>
            <div class='ext-info'>
              <div class='count'>{pagination.count}</div>
              <hcm-auth class='more-action' sign={clbCreateAuthSign.value}>
                {{
                  default: ({ noPerm }: { noPerm: boolean }) => (
                    <bk-button text disabled={noPerm} onClick={(e: Event) => showDropdownList(e, allTargetGroupsItem)}>
                      <i class='hcm-icon bkhcm-icon-more-fill'></i>
                    </bk-button>
                  ),
                }}
              </hcm-auth>
            </div>
          </div>
          <div class='group-list' onScroll={handleScroll}>
            {dataList.value.map((item) => {
              return (
                <div
                  key={item.id}
                  class={`group-item-wrap${activeTargetGroupId.value === item.id ? ' selected' : ''}`}
                  onClick={() => handleTypeChange(item)}>
                  <OverflowTitle type='tips' class='base-info'>
                    <img src={mubiaoIcon} alt='' class='prefix-icon' />
                    <span class='text'>{item.name}</span>
                  </OverflowTitle>
                  <div class={`ext-info${currentPopBoundaryNodeKey.value === item.id ? ' show-dropdown' : ''}`}>
                    <div class='count'>{item.count}</div>
                    <hcm-auth class='more-action' sign={clbOperationAuthSign.value}>
                      {{
                        default: ({ noPerm }: { noPerm: boolean }) => (
                          <bk-button
                            text
                            disabled={noPerm}
                            onClick={(e: Event) => showDropdownList(e, { type: 'specific', ...item })}>
                            <i class='hcm-icon bkhcm-icon-more-fill'></i>
                          </bk-button>
                        ),
                      }}
                    </hcm-auth>
                  </div>
                </div>
              );
            })}
            {isDataLoad.value && (
              <bk-loading loading size='small'>
                <div style='width: 100%; height: 36px' />
              </bk-loading>
            )}
          </div>
        </div>
      </div>
    );
  },
});
