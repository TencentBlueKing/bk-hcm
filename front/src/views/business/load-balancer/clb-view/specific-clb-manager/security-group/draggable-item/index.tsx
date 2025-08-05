import { PropType, defineComponent } from 'vue';
import routerAction from '@/router/utils/action';
import { MENU_BUSINESS_SECURITY_GROUP_MANAGEMENT } from '@/constants/menu-symbol';
import QueryString from 'qs';

import { Button, InfoBox } from 'bkui-vue';
import './index.scss';

export default defineComponent({
  props: {
    securityItem: Object as PropType<{
      cloud_id: string;
      id: string;
      name: string;
    }>,
    idx: Number,
    securitySearchVal: String,
    handleUnbind: Function,
    selectedSecuirtyGroupsSet: Set,
  },
  setup(props) {
    // 高亮命中关键词
    const getHighLightNameText = (name: string, rootCls: string) => {
      return (
        <div
          class={rootCls}
          v-html={name?.replace(
            new RegExp(props.securitySearchVal, 'g'),
            `<span class='search-result-highlight'>${props.securitySearchVal}</span>`,
          )}></div>
      );
    };

    const openSecurityGroupManagementPage = () => {
      routerAction.open({
        name: MENU_BUSINESS_SECURITY_GROUP_MANAGEMENT,
        query: {
          filter: QueryString.stringify(
            { cloud_id: props.securityItem.cloud_id },
            { arrayFormat: 'comma', encode: false, allowEmptyArrays: true },
          ),
        },
      });
    };

    return () => (
      <div
        key={props.securityItem.cloud_id}
        class={
          props.selectedSecuirtyGroupsSet.has(props.securityItem.id)
            ? 'config-security-item-new'
            : 'config-security-item'
        }>
        <i class={'hcm-icon bkhcm-icon-grag-fill mr8 draggable-card-header-draggable-btn'}></i>

        <div class={'config-security-item-idx'}>{props.idx + 1}</div>
        <span class={'config-security-item-name'}>
          {props.securitySearchVal ? getHighLightNameText(props.securityItem.name, '') : props.securityItem.name}
        </span>
        <span class={'config-security-item-id'}>({props.securityItem.cloud_id})</span>
        <div class={'config-security-item-edit-block'}>
          <Button text theme='primary' class={'mr27'} onClick={openSecurityGroupManagementPage}>
            去编辑
            <i class='icon hcm-icon bkhcm-icon-jump-fill ml5'></i>
          </Button>
          <Button
            class={'mr24'}
            text
            theme='danger'
            onClick={() => {
              InfoBox({
                infoType: 'warning',
                title: '确定解绑该安全组',
                cancelText: '取消',
                async onConfirm() {
                  await props.handleUnbind(props.securityItem.id);
                },
              });
            }}>
            <i class='hcm-icon bkhcm-icon-jiebang remove-icon'></i>
            解绑
          </Button>
        </div>
      </div>
    );
  },
});
