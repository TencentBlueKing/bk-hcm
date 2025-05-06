import { defineComponent, reactive, ref, watch } from 'vue';
import { Form, Input, Select, Radio, Button, Dialog, TagInput } from 'bkui-vue';
import { useI18n } from 'vue-i18n';
import {
  GCP_TYPE_STATUS,
  GCP_MATCH_STATUS,
  GCP_SOURCE_LIST,
  GCP_TARGET_LIST,
  GCP_PROTOCOL_LIST,
  GCP_EXECUTION_STATUS,
  COMMON_STATUS,
} from '@/constants';
import './gcp-add.scss';
import { timeFormatter } from '@/common/util';
export default defineComponent({
  name: 'GcpAdd',
  props: {
    loading: {
      type: Boolean,
    },
    isAdd: {
      type: Boolean,
    },
    detail: Object,
    isShow: {
      type: Boolean,
    },
    gcpTitle: {
      type: String,
    },
  },
  emits: ['update:isShow', 'submit'],
  setup(props, ctx) {
    const { t } = useI18n();
    const { FormItem } = Form;
    const { Option } = Select;
    const { Group } = Radio;
    const check = (val: any): boolean => {
      return /^[a-z][a-z-z0-9_-]*$/.test(val);
    };
    const formRef = ref<InstanceType<typeof Form>>(null);
    // eslint-disable-next-line @typescript-eslint/prefer-optional-chain
    // const gcpPorts = computed(() => (state.projectModel[state.operate]
    //   && state.projectModel[state.operate]      // 端口
    //     .find((e: any) => e.protocol === state.protocol)?.ports));
    // // const ports = computed(() => (state.projectModel[state.operate]
    // //   && state.projectModel[state.operate]      // 端口
    // //     .find((e: any) => e.protocol === state.protocol)));
    // // console.log('ports', ports);
    const gcpPorts = ref([]);
    const state = reactive({
      projectModel: {
        id: 0,
        type: 'egress', // 账号类型
        name: 'test', // 名称
        priority: '', // 优先级
        vpc_id: '--', // vpcid
        target_tags: [],
        destination_ranges: [],
        target_service_accounts: [],
        source_tags: [],
        source_service_accounts: [],
        source_ranges: [],
        bk_biz_id: 0, // 业务id
        created_at: '--',
        updated_at: '--',
        disabled: false,
        denied: [],
        allowed: [
          {
            protocol: 'tcp',
            port: ['443'],
          },
        ],
        memo: '',
        log_enable: false,
      },
      operate: 'allowed',
      target: 'destination_ranges',
      source: 'source_ranges',
      protocol: 'tcp',
      formList: [
        {
          label: t('名称'),
          property: 'name',
          component: () => (
            <section class='w450'>
              {props.isAdd ? (
                <Input class='w450' placeholder={t('请输入名称')} v-model={state.projectModel.name} />
              ) : (
                <span>{state.projectModel.name}</span>
              )}
            </section>
          ),
        },
        {
          label: t('业务'),
          property: 'resource-id',
          component: () => (
            <span>{state.projectModel.bk_biz_id === -1 ? '全部' : state.projectModel.bk_biz_id || '--'}</span>
          ),
        },
        {
          label: t('云厂商'),
          property: 'resource-id',
          component: () => <span>{t('谷歌云')}</span>,
        },
        {
          label: t('日志'),
          property: 'log_enable',
          component: () => (
            <>
              {state.projectModel.id ? (
                <span>{state.projectModel.log_enable ? t('开') : t('关')}</span>
              ) : (
                <Group v-model={state.projectModel.log_enable}>
                  {COMMON_STATUS.map((e) => (
                    <Radio label={e.value}>{t(e.label)}</Radio>
                  ))}
                </Group>
              )}
            </>
          ),
        },
        {
          label: 'VPC',
          property: 'vpc_id',
          component: () => <span>{state.projectModel.vpc_id}</span>,
        },
        {
          label: t('优先级'),
          property: 'priority',
          component: () => (
            <Input
              class='w450'
              type='number'
              min={0}
              max={65535}
              placeholder={t('请输入优先级')}
              v-model_number={state.projectModel.priority}
            />
          ),
        },
        {
          label: t('方向'),
          property: 'type',
          component: () => (
            <>
              {state.projectModel.id ? (
                <span>{state.projectModel.type === 'EGRESS' ? t('出站') : t('入站')}</span>
              ) : (
                <Group v-model={state.projectModel.type}>
                  {GCP_TYPE_STATUS.map((e) => (
                    <Radio label={e.value}>{t(e.label)}</Radio>
                  ))}
                </Group>
              )}
            </>
          ),
        },
        {
          label: t('对匹配项执行的操作'),
          property: 'resource-id',
          component: () => (
            <Group v-model={state.operate}>
              {GCP_MATCH_STATUS.map((e) => (
                <Radio label={e.value}>{t(e.label)}</Radio>
              ))}
            </Group>
          ),
        },
        {
          label: t('目标'),
          property: 'target_tags',
          component: () => (
            <section class='flex-row'>
              <Select v-model={state.target}>
                {GCP_TARGET_LIST.map((item) => (
                  <Option key={item.id} value={item.id} label={item.name}>
                    {item.name}
                  </Option>
                ))}
              </Select>
              <TagInput
                class='w450 ml20'
                allow-create
                allow-auto-match
                placeholder={t('请输入目标')}
                list={[]}
                v-model={state.projectModel[state.target]}
              />
            </section>
          ),
        },
        {
          label: t('来源过滤条件'),
          property: 'name',
          component: () => (
            <section class='flex-row'>
              <Select v-model={state.source}>
                {GCP_SOURCE_LIST.map((item) => (
                  <Option key={item.id} value={item.id} label={item.name}>
                    {item.name}
                  </Option>
                ))}
              </Select>
              <TagInput
                class='w450 ml20'
                allow-create
                allow-auto-match
                placeholder={t('请输入过滤条件')}
                list={[]}
                v-model={state.projectModel[state.source]}
              />
            </section>
          ),
        },
        // {
        //   label: t('次要来源过滤条件'),
        //   property: 'name',
        //   component: () => (
        //     <section class="flex-row">
        //         <Select v-model={state.projectModel.name}>
        //         {GCP_SOURCE_LIST.map(item => (
        //             <Option
        //                 key={item.id}
        //                 value={item.id}
        //                 label={item.name}
        //             >
        //                 {item.name}
        //             </Option>
        //         ))
        //         }
        //         </Select>
        //         <Input class="w450 ml20" placeholder={t('请输入名称')} v-model={state.projectModel.name} />
        //     </section>
        //   ),
        // },
        {
          label: t('协议和端口'),
          property: 'name',
          component: () => (
            <section class='flex-row'>
              <Select v-model={state.protocol}>
                {GCP_PROTOCOL_LIST.map((item) => (
                  <Option key={item.id} value={item.id} label={item.name}>
                    {item.name}
                  </Option>
                ))}
              </Select>
              <TagInput
                class='w450 ml20'
                allow-create
                allow-auto-match
                list={[]}
                placeholder={t('请输入端口')}
                v-model={gcpPorts.value}
                onBlur={handleBlur}
              />
            </section>
          ),
        },
        {
          label: t('强制执行'),
          property: 'disabled',
          component: () => (
            <>
              {state.projectModel.id ? (
                <span>{state.projectModel.disabled ? t('已停用') : t('已启用')}</span>
              ) : (
                <Group v-model={state.projectModel.disabled}>
                  {GCP_EXECUTION_STATUS.map((e) => (
                    <Radio label={e.value}>{t(e.label)}</Radio>
                  ))}
                </Group>
              )}
            </>
          ),
        },
        {
          label: t('创建时间'),
          property: 'resource-id',
          component: () => <span>{timeFormatter(state.projectModel.created_at)}</span>,
        },
        {
          label: t('修改时间'),
          property: 'resource-id',
          component: () => <span>{timeFormatter(state.projectModel.updated_at)}</span>,
        },
        {
          label: t('备注'),
          property: 'memo',
          component: () => (
            <Input class='w450' placeholder={t('请输入备注')} type='textarea' v-model={state.projectModel.memo} />
          ),
        },
      ],
      formRules: {
        name: [
          {
            trigger: 'blur',
            message: '名称必须以小写字母开头，后面最多可跟 32个小写字母、数字或连字符，但不能以连字符结尾',
            validator: check,
          },
        ],
      },
      buttonLoading: false,
    });

    watch(
      () => props.isShow,
      (val) => {
        if (val) {
          // @ts-ignore
          state.projectModel = { ...props.detail };
          state.projectModel.denied = state.projectModel.denied || [];
          state.projectModel.destination_ranges = state.projectModel.destination_ranges || [];
          state.projectModel.source_service_accounts = state.projectModel.source_service_accounts || [];
          state.projectModel.source_tags = state.projectModel.source_tags || [];
          state.projectModel.target_service_accounts = state.projectModel.target_service_accounts || [];
          state.projectModel.target_tags = state.projectModel.target_tags || [];
          // console.log('state.projectModel', state.projectModel);
          state.operate = state?.projectModel?.allowed?.length ? 'allowed' : 'denied';
          // eslint-disable-next-line max-len
          gcpPorts.value =
            state.projectModel[state.operate].find((e: any) => e.protocol === state.protocol)?.port || [];
          state.target = GCP_TARGET_LIST.find((e: any) => state.projectModel[e.id]?.length)?.id || 'destination_ranges';
          state.source = GCP_SOURCE_LIST.find((e: any) => state.projectModel[e.id]?.length)?.id;
        }
      },
    );

    watch(
      () => props.loading,
      (val) => {
        state.buttonLoading = val;
      },
    );

    watch(
      () => state.target,
      (newValue, oldValue) => {
        if (newValue !== oldValue) {
          state.projectModel[oldValue] = [];
        }
      },
    );

    watch(
      () => state.source,
      (newValue, oldValue) => {
        if (newValue !== oldValue) {
          state.projectModel[oldValue] = [];
        }
      },
    );

    watch(
      () => state.operate,
      (newValue, oldValue) => {
        if (state.projectModel[oldValue]?.length) {
          state.projectModel[newValue] = state.projectModel[oldValue];
          state.projectModel[oldValue] = [];
        }
      },
    );

    watch(
      () => state.protocol,
      () => {
        gcpPorts.value = state.projectModel[state.operate].find((e: any) => e.protocol === state.protocol)?.port || [];
      },
    );

    const handleBlur = () => {
      const protocolData = state.projectModel[state.operate].map((p: any) => p.protocol);
      if (!protocolData.includes(state.protocol)) {
        if (gcpPorts.value.length) {
          state.projectModel[state.operate].push({
            protocol: state.protocol,
            port: gcpPorts.value,
          });
        }
      } else {
        state.projectModel[state.operate].forEach((e: any, index: number) => {
          if (e.protocol === state.protocol) {
            if (gcpPorts.value.length === 0) {
              state.projectModel[state.operate].splice(index, 1);
            } else {
              e.port = gcpPorts.value;
            }
          }
        });
      }
    };

    const submit = () => {
      state.buttonLoading = true;
      const data = {
        id: state.projectModel.id,
        memo: state.projectModel.memo,
        priority: state.projectModel.priority,
        source_ranges: state.projectModel.source_ranges,
        destination_ranges: state.projectModel.destination_ranges,
        source_tags: state.projectModel.source_tags,
        target_tags: state.projectModel.target_tags,
        target_service_accounts: state.projectModel.target_service_accounts,
        source_service_accounts: state.projectModel.source_service_accounts,
        denied: state.projectModel.denied,
        allowed: state.projectModel.allowed,
      };
      ctx.emit('submit', data);
    };
    const cancel = () => {
      ctx.emit('update:isShow', false);
    };

    return () => (
      <Dialog isShow={props.isShow} title={props.gcpTitle} height={800} size='large' dialog-type='show'>
        <Form model={state.projectModel} labelWidth={140} rules={state.formRules} ref={formRef} class='gcp-form'>
          {state.formList.map((item: any) => (
            <FormItem label={item.label} property={item.property}>
              {item.component()}
            </FormItem>
          ))}
          <footer class='gcp-footer'>
            <Button class='w90' theme='primary' loading={state.buttonLoading} onClick={submit}>
              {t('确认')}
            </Button>
            <Button class='w90 ml20' onClick={cancel}>
              {t('取消')}
            </Button>
          </footer>
        </Form>
      </Dialog>
    );
  },
});
