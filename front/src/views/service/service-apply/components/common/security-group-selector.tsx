import { computed, defineComponent, PropType, ref, TransitionGroup, watch } from 'vue';
import { Button, Checkbox, Dialog, Exception, Input, Loading, Table } from 'bkui-vue';
import { BkButtonGroup } from 'bkui-vue/lib/button';
import { EditLine, Plus } from 'bkui-vue/lib/icon';
import { type UseDraggableReturn, VueDraggable } from 'vue-draggable-plus';
import './security-group-selector.scss';

import DraggableCard from './DraggableCard';

import { useI18n } from 'vue-i18n';
import { cloneDeep } from 'lodash';
import { useResourceStore } from '@/store';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { SECURITY_GROUP_RULE_TYPE, VendorEnum } from '@/common/constant';
import { QueryRuleOPEnum } from '@/typings';
import http from '@/http';

export default defineComponent({
  props: {
    modelValue: String as PropType<string | string[]>,
    bizId: Number as PropType<number | string>,
    accountId: String as PropType<string>,
    region: String as PropType<string>,
    multiple: Boolean as PropType<boolean>,
    vendor: String as PropType<string>,
    vpcId: String as PropType<string>,
    onSelectedChange: Function as PropType<(val: string[]) => void>,
  },
  emits: ['update:modelValue'],
  setup(props) {
    const { t } = useI18n();
    const resourceStore = useResourceStore();
    const { isServicePage, whereAmI, getBusinessApiPath } = useWhereAmI();

    const isDialogShow = ref(false);
    const cache: any = { securityList: [], securityGroupRules: [] }; // 缓存：记录编辑前的状态
    const updateCache = () => {
      cache.securityList = cloneDeep(securityList.value);
      cache.securityGroupRules = cloneDeep(securityGroupRules.value);
    };
    const applyCache = () => {
      // 首次confirm前的一切操作应被记录
      if (cache.securityList.length) securityList.value = cloneDeep(cache.securityList); // 深克隆用于应对连续close的情况
      if (cache.securityGroupRules.length) securityGroupRules.value = cloneDeep(cache.securityGroupRules);
    };
    const show = () => {
      isDialogShow.value = true;
      // 回显已确认选择的安全组
      confirmedSecurityGroupCloudList.value.length > 0 && initialState();
    };
    const hide = (apply: boolean) => {
      isDialogShow.value = false;
      // 应用缓存, 恢复编辑前的状态
      apply && applyCache();
    };

    const securityList = ref([]);
    const isSecurityListLoading = ref(false);
    const confirmedSecurityGroupCloudList = ref([]);
    const searchVal = ref('');

    const getSecurityList = async (accountId: string, region: string, vpcId?: string) => {
      isSecurityListLoading.value = true;
      try {
        const rules = [
          { field: 'account_id', op: 'eq', value: accountId },
          { field: 'region', op: 'eq', value: region },
        ];
        if (searchVal.value.length) {
          rules.push({ field: 'name', op: QueryRuleOPEnum.CS, value: searchVal.value });
        }
        if (vpcId) {
          rules.push({ field: 'extension.vpc_id', op: 'json_eq', value: vpcId });
        }
        // todo: 是否要加滚动加载？
        const result = await resourceStore.getCommonList(
          { filter: { op: 'and', rules }, page: { count: false, start: 0, limit: 500 } },
          'security_groups/list',
        );
        securityList.value = result?.data?.details?.map((item: any) => ({ ...item, isChecked: false })) ?? [];
      } catch (error) {
        securityList.value = [];
      } finally {
        // 清空状态
        confirmedSecurityGroupCloudList.value = [];
        securityGroupRules.value = [];
        isSecurityListLoading.value = false;
      }
    };

    // 重新加载安全组列表
    watch(
      [() => props.bizId, () => props.accountId, () => props.region, searchVal],
      async ([bizId, accountId, region]) => {
        if ((!bizId && isServicePage) || !accountId || !region) {
          securityList.value = [];
          return;
        }
        getSecurityList(accountId, region);
      },
    );
    watch(
      () => props.vpcId,
      (val) => {
        if (VendorEnum.AWS === props.vendor && val) {
          getSecurityList(props.accountId, props.region, val);
        }
      },
    );

    const securityGroupRules = ref([]);
    const isRulesTableLoading = ref(false);
    const el = ref<UseDraggableReturn>();
    const selectedSecurityType = ref(SECURITY_GROUP_RULE_TYPE.INGRESS);

    const computedDisabled = computed(() => {
      return !(props.accountId && props.vendor && props.region);
    });

    const computedSecurityGroupRules = computed(() => {
      return securityGroupRules.value.map(({ id, name, data }) => ({
        id,
        name,
        data: data.filter(({ type }: any) => type === selectedSecurityType.value),
      }));
    });

    const securityRulesColumns = useColumns('securityCommon', false, props.vendor).columns.filter(
      ({ field }: { field: string }) => !['updated_at'].includes(field),
    );

    const isAllExpand = ref(true);

    const currentIndex = ref(-1);
    const handleSecurityGroupChange = async (isSelected: boolean, item: any, index: number) => {
      if (isSelected) {
        currentIndex.value = index;
        try {
          isRulesTableLoading.value = true;
          const res = await http.post(
            `/api/v1/cloud/${getBusinessApiPath()}vendors/${props.vendor}/security_groups/${item.id}/rules/list`,
            {
              filter: { op: 'and', rules: [] },
              page: { count: false, start: 0, limit: 500 },
            },
          );
          const arr = res.data?.details || [];
          securityGroupRules.value.push({ id: item.cloud_id, name: item.name, data: arr });
        } finally {
          isRulesTableLoading.value = false;
        }
      } else {
        securityGroupRules.value = securityGroupRules.value.filter(({ id }) => id !== item.cloud_id);
      }
    };

    const initialState = () => {
      securityGroupRules.value.forEach(({ id }, index) => {
        if (!confirmedSecurityGroupCloudList.value.find((item) => item.id === id)) {
          securityGroupRules.value.splice(index, 1);
        }
      });
    };

    const handleConfirm = () => {
      // 记录已确认选择的安全组, 用于回显
      confirmedSecurityGroupCloudList.value = securityGroupRules.value.map(({ id, name }) => ({ id, name })) || [];
      // 数据收集
      props.onSelectedChange(confirmedSecurityGroupCloudList.value.map(({ id }) => id));
      // 缓存确认时的状态
      updateCache();
      hide(false);
    };

    return () => (
      <div>
        {confirmedSecurityGroupCloudList.value.length > 0 ? (
          // 回显已确认选择的安全组
          <div class={'image-selector-selected-block-container'}>
            <div class={'selected-block mr8'}>
              {confirmedSecurityGroupCloudList.value.map(({ id, name }) => (
                <p key={id}>{name}</p>
              ))}
            </div>
            <EditLine class='cursor' fill='#3A84FF' width={13.5} height={13.5} onClick={show} />
          </div>
        ) : (
          <Button onClick={show} disabled={computedDisabled.value || securityList.value.length === 0}>
            <Plus class='f20' />
            {t('选择安全组')}
          </Button>
        )}
        {securityList.value.length || computedDisabled.value ? null : (
          <div class={'security-selector-tips'}>
            {t('无可用的安全组，可')}
            <Button
              theme='primary'
              text
              onClick={() => {
                const url =
                  whereAmI.value === Senarios.business ? '/#/business/security' : '/#/resource/resource?type=security';
                window.open(url, '_blank');
              }}>
              {t('新建安全组')}
            </Button>
          </div>
        )}
        <Dialog
          class={'security-dialog-wrap'}
          isShow={isDialogShow.value}
          onClosed={() => hide(true)}
          onConfirm={handleConfirm}
          title={t('选择安全组')}
          width={'60vw'}>
          <div class={'security-container'}>
            <div class={'security-list g-scroller'}>
              <Input
                class={'search-input'}
                placeholder={t('搜索安全组')}
                type='search'
                clearable
                v-model={searchVal.value}
              />
              <Loading loading={isSecurityListLoading.value} class={'mt8'}>
                <div class={'security-search-list'}>
                  {securityList.value.length ? (
                    securityList.value.map((item, index) => (
                      <div class={'security-search-item'} key={item.id}>
                        <Checkbox
                          v-model={item.isChecked}
                          disabled={
                            props.vendor === VendorEnum.AZURE &&
                            securityGroupRules.value.length > 0 &&
                            currentIndex.value !== index
                          }
                          label={'data.cloud_id'}
                          onChange={(isSelected: boolean) => handleSecurityGroupChange(isSelected, item, index)}>
                          {item.name}
                        </Checkbox>
                      </div>
                    ))
                  ) : (
                    <Exception
                      class='exception-wrap-item exception-part'
                      type='search-empty'
                      scene='part'
                      description={t('搜索为空')}
                    />
                  )}
                </div>
              </Loading>
            </div>
            <div class={'security-group-rules-wrap'}>
              <div class={'security-group-rules-container'}>
                <div class={'security-group-rules-btn-group-container'}>
                  <BkButtonGroup>
                    <Button
                      selected={selectedSecurityType.value === SECURITY_GROUP_RULE_TYPE.EGRESS}
                      onClick={() => (selectedSecurityType.value = SECURITY_GROUP_RULE_TYPE.EGRESS)}>
                      {t('出站规则')}
                    </Button>
                    <Button
                      selected={selectedSecurityType.value === SECURITY_GROUP_RULE_TYPE.INGRESS}
                      onClick={() => (selectedSecurityType.value = SECURITY_GROUP_RULE_TYPE.INGRESS)}>
                      {t('入站规则')}
                    </Button>
                  </BkButtonGroup>
                  <Button style={{ marginRight: '1px' }} onClick={() => (isAllExpand.value = !isAllExpand.value)}>
                    {isAllExpand.value ? (
                      <>
                        <i class='hcm-icon bkhcm-icon-zoomout'></i>
                        <span class='ml8'>{t('全部收起')}</span>
                      </>
                    ) : (
                      <>
                        <i class='hcm-icon bkhcm-icon-fullscreen'></i>
                        <span class='ml8'>{t('全部展开')}</span>
                      </>
                    )}
                  </Button>
                </div>
                <Loading loading={isRulesTableLoading.value}>
                  {/* @ts-ignore */}
                  <VueDraggable
                    ref={el}
                    v-model={securityGroupRules.value}
                    animation={200}
                    handle='.draggable-card-header-draggable-btn'
                    class={'security-group-rules-list g-scroller'}>
                    {computedSecurityGroupRules.value.length ? (
                      <TransitionGroup type='transition' name='fade'>
                        {computedSecurityGroupRules.value.map(({ name, data }, idx) => (
                          <DraggableCard key={idx} title={name} index={idx + 1} isAllExpand={isAllExpand.value}>
                            <Table data={data} columns={securityRulesColumns} showOverflowTooltip stripe={true} />
                          </DraggableCard>
                        ))}
                      </TransitionGroup>
                    ) : (
                      <Exception
                        class='exception-wrap-item exception-part'
                        type='empty'
                        scene='part'
                        description={t('没有数据')}
                      />
                    )}
                  </VueDraggable>
                </Loading>
              </div>
            </div>
          </div>
        </Dialog>
      </div>
    );
  },
});
