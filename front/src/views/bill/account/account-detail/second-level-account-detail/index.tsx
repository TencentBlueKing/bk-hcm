import { PropType, defineComponent } from 'vue';
import './index.scss';

export default defineComponent({
  props: {
    detail: {
      type: Object as PropType<{
        // 二级帐号名称
        secondaryAccountName: string;
        // 二级帐号ID
        secondaryAccountId: string;
        // 所属一级帐号
        parentPrimaryAccount: string;
        // 云厂商
        cloudProvider: string;
        // 站点类型
        siteType: string;
        // 帐号邮箱
        accountEmail: string;
        // 主负责人
        mainResponsiblePerson: string;
        // 运营产品
        operatingProduct: string;
      }>,
    },
  },
  setup(props) {
    return () => (
      <div class={'account-detail-wrapper'}>
        <p class={'sub-title'}>帐号信息</p>
        <div class={'detail-info'}>
          <div class='item'>
            <span class='label'>二级帐号名称：</span>
            <span class='value'>{props.detail.secondaryAccountName}</span>
          </div>
          <div class='item'>
            <span class='label'>二级帐号ID：</span>
            <span class='value'>{props.detail.secondaryAccountId}</span>
          </div>
          <div class='item'>
            <span class='label'>所属一级帐号：</span>
            <span class='value'>{props.detail.parentPrimaryAccount}</span>
          </div>
          <div class='item'>
            <span class='label'>云厂商：</span>
            <span class='value'>{props.detail.cloudProvider}</span>
          </div>
          <div class='item'>
            <span class='label'>站点类型：</span>
            <span class='value'>{props.detail.siteType}</span>
          </div>
          <div class='item'>
            <span class='label'>帐号邮箱：</span>
            <span class='value'>{props.detail.accountEmail}</span>
          </div>
          <div class='item'>
            <span class='label'>主负责人：</span>
            <span class='value'>{props.detail.mainResponsiblePerson}</span>
          </div>
          <div class='item'>
            <span class='label'>运营产品：</span>
            <span class='value'>{props.detail.operatingProduct}</span>
          </div>
        </div>
        <p class={'sub-title'}>API 密钥</p>
        <div class={'detail-info'}>
          <div class='item'>
            <span class='label'>云密钥 ID：</span>
            <span class='value'>{'************'}</span>
          </div>
          <div class='item'>
            <span class='label'>云密钥：</span>
            <span class='value'>{'************'}</span>
          </div>
          <div class='item'>
            <span class='label'>所属账号 ID：</span>
            <span class='value'>{'************'}</span>
          </div>
        </div>
      </div>
    );
  },
});
