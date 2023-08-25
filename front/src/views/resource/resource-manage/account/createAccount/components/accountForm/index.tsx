import { Card, Form, Radio } from 'bkui-vue';
import { defineComponent, reactive } from 'vue';
import './index.scss';
import { VendorEnum } from '@/common/constant';
import tcloudVendor from '@/assets/image/vendor-tcloud.png';
import awsVendor from '@/assets/image/vendor-aws.png';
import azureVendor from '@/assets/image/vendor-azure.png';
import gcpVendor from '@/assets/image/vendor-gcp.png';
import huaweiVendor from '@/assets/image/vendor-huawei.png';
import { Success } from 'bkui-vue/lib/icon';

const { FormItem } = Form;

export const VENDORS_INFO = [
  {
    vendor: VendorEnum.TCLOUD,
    name: '腾讯云',
    icon: tcloudVendor,
  },
  {
    vendor: VendorEnum.AWS,
    name: '亚马逊云',
    icon: awsVendor,
  },
  {
    vendor: VendorEnum.AZURE,
    name: '微软云',
    icon: azureVendor,
  },
  {
    vendor: VendorEnum.GCP,
    name: '谷歌云',
    icon: gcpVendor,
  },
  {
    vendor: VendorEnum.HUAWEI,
    name: '华为云',
    icon: huaweiVendor,
  },
];


export default defineComponent({
  setup() {
    const formModel = reactive({
      site: 'china' as 'china'|'international', // 站点
      vendor: VendorEnum.TCLOUD, // 云厂商
    });
    return () => (
      <div class={'account-form'}>
        <Card
          class={'account-form-card'}
          showHeader={false}
        >
          <p class={'account-form-card-title'}>
            账号归属
          </p>
          <div class={'account-form-card-content'}>
            <Form
              formType='vertical'
            >
              <FormItem
                label='厂商选择'
              >
                <div class={'account-vendor-selector'}>
                  {
                    VENDORS_INFO.map(({ vendor, name, icon }) => (
                      <div
                        class={`account-vendor-option ${vendor === formModel.vendor ? 'account-vendor-option-active' : ''}`}
                        onClick={() => formModel.vendor = vendor}
                        >
                        <img src={icon} alt={name} class={'account-vendor-option-icon'}/>
                        <p class={'account-vendor-option-text'}>
                        { name }
                        </p>
                        {
                          formModel.vendor === vendor
                            ? <Success fill='#3A84FF' class={'active-icon'}/>
                            : null
                        }
                      </div>
                    ))
                  }
                </div>
              </FormItem>
              <FormItem
                label='站点种类'
              >
                <Radio
                  label={'china'}
                  v-model={formModel.site}
                >
                  中国站
                </Radio>
                <Radio
                  label={'international'}
                  v-model={formModel.site}
                >
                  国际站
                </Radio>
              </FormItem>
            </Form>
          </div>
        </Card>

        <Card
          class={'account-form-card'}
          showHeader={false}
        >
          <p class={'account-form-card-title'}>
            API 密钥
          </p>
          <div class={'account-form-card-content'}>
            456
          </div>
        </Card>

        <Card
          class={'account-form-card'}
          showHeader={false}
        >
          <p class={'account-form-card-title'}>
            其他信息
          </p>
          <div class={'account-form-card-content'}>
            789
          </div>
        </Card>
      </div>
    );
  },
});
