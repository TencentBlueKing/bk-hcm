import { defineComponent, onMounted, reactive, ref } from 'vue';
import { Button, Form, Input, Message } from 'bkui-vue';
import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import CommonDialog from '@/components/common-dialog';
import SubnetSelector from '@/views/service/service-apply/components/common/subnet-selector';
import { useBusinessStore } from '@/store';
import { useI18n } from 'vue-i18n';
import { CLB_QUOTA_NAME } from '@/typings';
import './index.scss';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { ResourceTypeEnum } from '@/common/resource-constant';

const { FormItem } = Form;

/**
 * 新增 SNAT IP 对话框
 */
export default defineComponent({
  name: 'AddSnatIpDialog',
  props: { isShow: Boolean, lbInfo: Object, vpcDetail: Object, reloadLbDetail: Function },
  emits: ['update:isShow'],
  setup(props, { emit }) {
    const { t } = useI18n();
    const { getBizsId } = useWhereAmI();
    const businessStore = useBusinessStore();

    const getDefaultFormData = () => ({
      bk_biz_id: getBizsId(),
      cloud_subnet_id: '',
      type: '0', // 0 自动生成 1 手动录入
      ip_count: 0,
      ip_list: [''],
    });
    const formData = reactive(getDefaultFormData());
    const formRef = ref();
    const snatIpQuotaLimit = ref(0);
    const canAdd = ref(true);
    const subnetSelectorRef = ref();
    const isSubmitLoading = ref(false);

    // click-handler - 添加一条 IP
    const handleAddIp = () => {
      if (formData.ip_list.length >= snatIpQuotaLimit.value) {
        Message({ theme: 'warning', message: `最多添加 ${snatIpQuotaLimit.value} 个 IP` });
        canAdd.value = false;
        return;
      }
      formData.ip_list.push('');
    };

    // click-handler - 移除一条 IP
    const handleRemoveIp = (ip: string) => {
      const index = formData.ip_list.findIndex((item) => item === ip);
      formData.ip_list.splice(index, 1);
      canAdd.value = true;
    };

    // submit-handler
    const handleConfirm = async () => {
      await formRef.value.validate();
      try {
        isSubmitLoading.value = true;
        // 整理参数
        const snat_ips =
          formData.type === '0'
            ? Array.from({ length: formData.ip_count }, () => ({ subnet_id: formData.cloud_subnet_id }))
            : formData.ip_list.map((ip) => ({ subnet_id: formData.cloud_subnet_id, ip }));
        await businessStore.createSnatIps(props.lbInfo?.id, props.lbInfo?.vendor, { snat_ips });
        Message({ theme: 'success', message: '新增成功' });
        await props.reloadLbDetail(props.lbInfo?.id);
      } finally {
        isSubmitLoading.value = false;
      }
      emit('update:isShow', false);
    };

    // 获取腾讯云账号负载均衡的配额
    const getTotalSnapIpQuota = async () => {
      const res = await businessStore.getClbQuotas({
        account_id: props.lbInfo?.account_id,
        region: props.lbInfo?.region,
      });
      const snatIpQuota = res.data?.find((item) => item.quota_id === CLB_QUOTA_NAME.TOTAL_SNAT_IP_QUOTA);
      snatIpQuotaLimit.value = snatIpQuota?.quota_limit || 0;
    };

    onMounted(() => {
      getTotalSnapIpQuota();
    });

    return () => (
      <CommonDialog
        title='新增 SNAT IP'
        class='add-snat-ip-dialog'
        isShow={props.isShow}
        width={600}
        onUpdate:isShow={(isShow) => emit('update:isShow', isShow)}>
        {{
          default: () => (
            <>
              <div class='vpc-info'>
                <div class='label'>{t('所属VPC')}</div>
                <div class='value'>
                  {props.vpcDetail?.name}({props.vpcDetail?.cloud_id}){' '}
                  {props.vpcDetail?.extension?.cidr?.map((item: any) => item.cidr)?.join(',')} 共
                  {subnetSelectorRef.value?.subnetList?.length}个子网
                </div>
              </div>
              <Form formType='vertical' model={formData} ref={formRef}>
                <FormItem label='子网' required property='cloud_subnet_id'>
                  <SubnetSelector
                    ref={subnetSelectorRef}
                    v-model={formData.cloud_subnet_id}
                    bizId={formData.bk_biz_id}
                    vpcId={props.vpcDetail?.id}
                    vendor={props.vpcDetail?.vendor}
                    region={props.vpcDetail?.region}
                    accountId={props.vpcDetail?.account_id}
                    zone={props.lbInfo?.zones?.[0]}
                    resourceType={ResourceTypeEnum.CLB}
                  />
                </FormItem>
                <FormItem label='分配IP' required property='type'>
                  <BkRadioGroup type='card' v-model={formData.type}>
                    <BkRadioButton label='0'>{t('自动生成')}</BkRadioButton>
                    <BkRadioButton label='1'>{t('手动录入')}</BkRadioButton>
                  </BkRadioGroup>
                </FormItem>
                {formData.type === '0' ? (
                  <FormItem label='IP数量' required property='ip_count'>
                    <Input
                      type='number'
                      v-model={formData.ip_count}
                      placeholder='0'
                      min={0}
                      max={snatIpQuotaLimit.value}
                    />
                  </FormItem>
                ) : (
                  <FormItem label='IP' required property='ip_list'>
                    {formData.ip_list.map((ip, index) => (
                      <div class='ip-input-container'>
                        <FormItem
                          property={`ip_list.${index}`}
                          required
                          errorDisplayType='tooltips'
                          rules={[
                            {
                              trigger: 'blur',
                              validator: (value: string) =>
                                /^(?!0)(?!.*\.$)((1\d\d|2[0-4]\d|25[0-5]|[1-9]?\d)(\.|$)){4}$/.test(value),
                              message: '请输入正确的SNAT IP格式',
                            },
                            {
                              trigger: 'blur',
                              validator: (value: string) => formData.ip_list.filter((ip) => ip === value).length <= 1,
                              message: '重复的SNAT IP',
                            },
                          ]}>
                          <Input v-model={formData.ip_list[index]} />
                        </FormItem>
                        <Button text onClick={handleAddIp} disabled={!canAdd.value}>
                          <i class='hcm-icon bkhcm-icon-plus-circle-shape'></i>
                        </Button>
                        <Button text onClick={() => handleRemoveIp(ip)}>
                          <i class='hcm-icon bkhcm-icon-minus-circle-shape'></i>
                        </Button>
                      </div>
                    ))}
                  </FormItem>
                )}
              </Form>
            </>
          ),
          footer: () => (
            <div class='footer-container'>
              <div class='quota-wrap'>
                {t('配额')}
                &nbsp;:&nbsp;
                <span class='add-count'>{formData.type === '0' ? formData.ip_count : formData.ip_list.length}</span>
                &nbsp;/&nbsp;
                <span class='total-count'>{snatIpQuotaLimit.value - props.lbInfo?.extension?.snat_ips?.length}</span>
              </div>
              <div class='btn-wrap'>
                <Button class='mr8' theme='primary' onClick={handleConfirm} loading={isSubmitLoading.value}>
                  {t('确定')}
                </Button>
                <Button onClick={() => emit('update:isShow', false)}>{t('取消')}</Button>
              </div>
            </div>
          ),
        }}
      </CommonDialog>
    );
  },
});
