import { PropType, defineComponent } from 'vue';
import './index.scss';

export default defineComponent({
  props: {
    detail: {
      required: true,
      type: Object as PropType<{
        // 一级帐号名称
        primaryAccountName: string;
        // 一级帐号ID
        primaryAccountId: string;
        // 云厂商
        cloudProvider: string;
        // 帐号邮箱
        accountEmail: string;
        // 主负责人
        mainResponsiblePerson: string;
        // 组织架构
        organizationalStructure: string;
        // 二级帐号个数
        secondaryAccountCount: number;
      }>,
    },
  },
  setup(props) {
    return () => (
      <div class={'account-detail-wrapper'}>
        <p class={'sub-title'}>帐号信息</p>
        <div class={'detail-info'}>
          <div class='item'>
            <span class='label'>一级帐号名称：</span>
            <span class='value'>{props.detail.primaryAccountName}</span>
          </div>
          <div class='item'>
            <span class='label'>一级帐号ID：</span>
            <span class='value'>{props.detail.primaryAccountId}</span>
          </div>
          <div class='item'>
            <span class='label'>云厂商：</span>
            <span class='value'>{props.detail.cloudProvider}</span>
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
            <span class='label'>组织架构：</span>
            <span class='value'>{props.detail.organizationalStructure}</span>
          </div>
          <div class='item'>
            <span class='label'>二级帐号个数：</span>
            <span class='value'>{props.detail.secondaryAccountCount}</span>
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
