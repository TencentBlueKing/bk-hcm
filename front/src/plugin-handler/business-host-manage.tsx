import { h, ref, withDirectives, Ref } from 'vue';
import { Button, Dropdown, Message, Checkbox, bkTooltips } from 'bkui-vue';
import { useResourceStore } from '@/store';
import { CLOUD_HOST_STATUS } from '@/common/constant';
import defaultUseColumns from '@/views/resource/resource-manage/hooks/use-columns';
import HostOperations, {
  HOST_RUNNING_STATUS,
  HOST_SHUTDOWN_STATUS,
} from '@/views/business/host/children/host-operations';
import defaultUseTableListQuery from '@/hooks/useTableListQuery';
import type { PropsType } from '@/hooks/useTableListQuery';
import Confirm, { confirmInstance } from '@/components/confirm';

const { DropdownMenu, DropdownItem } = Dropdown;
const resourceStore = useResourceStore();

type UseColumnsParams = {
  type?: string;
  isSimpleShow?: boolean;
  vendor?: string;
  extra?: {
    isLoading: Ref<boolean>;
    triggerApi: () => void;
  };
};

const useColumns = ({ type = 'cvms', isSimpleShow = false, vendor, extra }: UseColumnsParams) => {
  const currentOperateRowIndex = ref(-1);

  // 回收参数「云硬盘/EIP 随主机回收」
  const isRecycleDiskWithCvm = ref(false);
  const isRecycleEipWithCvm = ref(false);

  // 重置回收参数
  const resetRecycleSingleCvmParams = () => {
    isRecycleDiskWithCvm.value = false;
    isRecycleEipWithCvm.value = false;
  };

  // 主机相关操作 - 单个操作
  const handleCvmOperate = async (label: string, type: string, data: any) => {
    // 判断当前主机是否可以执行对应操作
    if (cvmInfo.value[type].status.includes(data.status)) return;
    resetRecycleSingleCvmParams();
    let infoboxContent;
    if (type === 'recycle') {
      // 请求 cvm 所关联的资源(硬盘, eip)个数
      const {
        data: [target],
      } = await resourceStore.getRelResByCvmIds({ ids: [data.id] });
      const { disk_count, eip_count, eip } = target;
      infoboxContent = h('div', { style: { textAlign: 'justify' } }, [
        h('div', { style: { marginBottom: '10px' } }, [
          `当前操作主机为：${data.name}`,
          h('br'),
          `共关联 ${disk_count - 1} 个数据盘，${eip_count} 个弹性 IP${eip ? '('.concat(eip.join(','), ')') : ''}`,
        ]),
        h('div', null, [
          h(
            Checkbox,
            {
              checked: isRecycleDiskWithCvm.value,
              onChange: (checked: boolean) => (isRecycleDiskWithCvm.value = checked),
            },
            '云硬盘随主机回收',
          ),
          h(
            Checkbox,
            {
              checked: isRecycleEipWithCvm.value,
              onChange: (checked: boolean) => (isRecycleEipWithCvm.value = checked),
            },
            '弹性 IP 随主机回收',
          ),
        ]),
      ]);
    } else {
      infoboxContent = `当前操作主机为：${data.name}`;
    }
    Confirm(`确定${label}`, infoboxContent, async () => {
      confirmInstance.hide();
      extra.isLoading.value = true;
      try {
        if (type === 'recycle') {
          await resourceStore.recycledCvmsData({
            infos: [{ id: data.id, with_disk: isRecycleDiskWithCvm.value, with_eip: isRecycleEipWithCvm.value }],
          });
        } else {
          await resourceStore.cvmOperate(type, { ids: [data.id] });
        }
        Message({ message: t('操作成功'), theme: 'success' });
        extra.triggerApi();
      } finally {
        extra.isLoading.value = false;
      }
    });
  };

  const operationDropdownList = [
    { label: '开机', type: 'start' },
    { label: '关机', type: 'stop' },
    { label: '重启', type: 'reboot' },
  ];

  // 操作的相关信息
  const cvmInfo = ref({
    start: { op: '开机', loading: false, status: HOST_RUNNING_STATUS },
    stop: {
      op: '关机',
      loading: false,
      status: HOST_SHUTDOWN_STATUS,
    },
    reboot: { op: '重启', loading: false, status: HOST_SHUTDOWN_STATUS },
    recycle: { op: '回收', loading: false, status: HOST_SHUTDOWN_STATUS },
  });

  const getBkToolTipsOption = (data: any) => {
    return {
      content: `当前主机处于 ${CLOUD_HOST_STATUS[data.status]} 状态`,
      disabled: !cvmInfo.value.stop.status.includes(data.status),
    };
  };

  const { columns, generateColumnsSettings } = defaultUseColumns(type, isSimpleShow, vendor);

  return {
    columns: [
      ...columns,
      {
        label: '操作',
        width: 120,
        showOverflowTooltip: false,
        render: ({ data, index }: { data: any; index: number }) => {
          return h('div', { class: 'operation-column' }, [
            withDirectives(
              h(
                Button,
                {
                  text: true,
                  theme: 'primary',
                  class: 'mr10',
                  onClick: () => {
                    handleCvmOperate('回收', 'recycle', data);
                  },
                  // TODO: 权限
                  disabled: cvmInfo.value.stop.status.includes(data.status),
                },
                '回收',
              ),
              [[bkTooltips, getBkToolTipsOption(data)]],
            ),
            h(
              Dropdown,
              {
                trigger: 'click',
                popoverOptions: {
                  renderType: 'shown',
                  onAfterShow: () => (currentOperateRowIndex.value = index),
                  onAfterHidden: () => (currentOperateRowIndex.value = -1),
                },
              },
              {
                default: () =>
                  h(
                    'div',
                    {
                      class: [`more-action${currentOperateRowIndex.value === index ? ' current-operate-row' : ''}`],
                    },
                    h('i', { class: 'hcm-icon bkhcm-icon-more-fill' }),
                  ),
                content: () =>
                  h(
                    DropdownMenu,
                    null,
                    operationDropdownList.map(({ label, type }) => {
                      return withDirectives(
                        h(
                          DropdownItem,
                          {
                            key: type,
                            onClick: () => handleCvmOperate(label, type, data),
                            extCls: `more-action-item${
                              cvmInfo.value[type].status.includes(data.status) ? ' disabled' : ''
                            }`,
                          },
                          label,
                        ),
                        [
                          [
                            bkTooltips,
                            {
                              content: `当前主机处于 ${CLOUD_HOST_STATUS[data.status]} 状态`,
                              disabled: !cvmInfo.value[type].status.includes(data.status),
                            },
                          ],
                        ],
                      );
                    }),
                  ),
              },
            ),
          ]);
        },
      },
    ],
    generateColumnsSettings,
  };
};

const useTableListQuery = (
  props: PropsType,
  type = 'cvms',
  apiMethod?: Function,
  apiName = 'list',
  args: any = {},
  extraResolveData?: (...args: any) => Promise<any>,
) => {
  return defaultUseTableListQuery(props, type, apiMethod, apiName, args, extraResolveData);
};

const pluginHandler = {
  useColumns,
  useTableListQuery,
  HostOperations,
};

export default pluginHandler;

export type PluginHandlerType = typeof pluginHandler;
