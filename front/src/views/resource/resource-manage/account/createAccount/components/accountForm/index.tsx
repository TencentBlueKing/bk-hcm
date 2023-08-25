import { Button, Card, Form, Input, Radio } from 'bkui-vue';
import { defineComponent, reactive } from 'vue';
import './index.scss';
import { VendorEnum } from '@/common/constant';
import tcloudVendor from '@/assets/image/vendor-tcloud.png';
import awsVendor from '@/assets/image/vendor-aws.png';
import azureVendor from '@/assets/image/vendor-azure.png';
import gcpVendor from '@/assets/image/vendor-gcp.png';
import huaweiVendor from '@/assets/image/vendor-huawei.png';
import { Success, InfoLine, TextFile } from 'bkui-vue/lib/icon';

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
          showFooter={true}
        >
          {{
            default: () => (
              <>
                <div class={'api-secret-header'}>
                  <p class={'account-form-card-title'}>API 密钥</p>
                  <InfoLine fill='#979BA5' />
                  <p class={'header-text'}>
                    同一个主账号下,只允许接入一次。如后续对API密钥更新,必须是隶属于同一主账号D。
                  </p>
                  <TextFile fill='#3A84FF' />
                  <Button theme='primary' text class={'header-btn'}>
                    接入指引
                  </Button>
                </div>
                <div class={'account-form-card-content'}>
                  <Form formType='vertical'>
                    <FormItem label='SecretID/密钥ID'>
                      <Input class={'api-secret-selector'}/>
                    </FormItem>
                    <FormItem label='SecretKey'>
                      <Input class={'api-secret-selector'}/>
                    </FormItem>
                  </Form>
                </div>
                <Button theme='primary' class={'account-validate-btn'}>
                  账号校验
                </Button>
              </>
            ),
            footer: () => (
              <div class={'api-secret-footer'}>
                <Form class={'api-secret-footer'}>
                  <FormItem
                    label={'主账号ID'}
                    class={'footer-form-item'}
                  >
                    <p class={'api-secret-footer-content'}></p>
                  </FormItem>
                  <FormItem
                    label={'所属账号ID'}
                    description={'666666'}
                    class={'footer-form-item'}
                  >
                    <p class={'api-secret-footer-content'}></p>
                  </FormItem>
                </Form>
              </div>
            ),
          }}
        </Card>

        <Card
          class={'account-form-card'}
          showHeader={false}
        >
          <p class={'account-form-card-title'}>
            其他信息
          </p>
          <div class={'account-form-card-content'}>
            <Form
              formType='vertical'
            >
              <FormItem label='账号别名' class={'api-secret-selector'}>
                <Input/>
              </FormItem>
              <FormItem label='责任人' class={'api-secret-selector'}>
                <Input/>
              </FormItem>
              <FormItem label='备注'>
                <Input type={'textarea'}/>
              </FormItem>
            </Form>
          </div>
        </Card>
      </div>
    );
  },
});
