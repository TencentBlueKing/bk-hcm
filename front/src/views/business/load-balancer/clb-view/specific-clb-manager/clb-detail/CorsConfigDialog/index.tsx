import { defineComponent } from 'vue';
import CommonDialog from '@/components/common-dialog';
import { useRegionsStore } from '@/store/useRegionsStore';
import { useI18n } from 'vue-i18n';
import './index.scss';

export default defineComponent({
  name: 'CorsConfigDialog',
  props: {
    isShow: Boolean,
    lbInfo: Object,
    listenerNum: Number,
    vpcDetail: Object,
    targetVpcDetail: Object,
  },
  emits: ['update:isShow'],
  setup(props, { emit }) {
    const { t } = useI18n();
    const { getRegionName } = useRegionsStore();

    return () => (
      <CommonDialog
        isShow={props.isShow}
        onUpdate:isShow={(isShow) => emit('update:isShow', isShow)}
        title='后端服务配置'
        dialogType='show'
        class='cors-config-dialog'>
        {/* {props.listenerNum && (
          <Alert theme='warning'>{t('该负载均衡已绑定后端服务，请解绑后端服务后再修改地域')}</Alert>
        )} */}
        <div class='config-item'>
          <div class='label'>{t('CLB 所在地域')}</div>
          <div class='value'>{getRegionName(props.lbInfo?.vendor, props.lbInfo?.region)}</div>
        </div>
        <div class='config-item'>
          <div class='label'>{t('CLB 所属网络')}</div>
          <div class='value'>
            {props.vpcDetail?.name}({props.vpcDetail?.cloud_id})
          </div>
        </div>
        <div class='config-item'>
          <div class='label'>{t('后端服务地域')}</div>
          <div class='value'>{getRegionName(props.lbInfo?.vendor, props.lbInfo?.extension?.target_region)}</div>
        </div>
        <div class='config-item'>
          <div class='label'>{t('后端服务网络')}</div>
          <div class='value'>
            {props.targetVpcDetail?.name}({props.targetVpcDetail?.cloud_id})
          </div>
        </div>
      </CommonDialog>
    );
  },
});
