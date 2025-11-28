import { defineComponent, onMounted, ref, watch, nextTick, computed, reactive, useTemplateRef } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import './index.scss';
import { Input, Button, Sideslider, Message, Dropdown, Alert } from 'bkui-vue';
import { Plus } from 'bkui-vue/lib/icon';
import CommonCard from '@/components/CommonCard';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import BusinessSelector from '@/components/business-selector/index.vue';
import AreaSelector from '../AreaSelector';
import ZoneTagSelector from '@/components/zone-tag-selector/index.vue';
import CvmSystemDisk from '@/views/ziyanScr/components/cvm-system-disk/form.vue';
import CvmDataDisk from '@/views/ziyanScr/components/cvm-data-disk/form.vue';
import NetworkInfoCollapsePanel from '../network-info-collapse-panel/index.vue';
import AntiAffinityLevelSelect from '../AntiAffinityLevelSelect';
import FormCvmImageSelector from '@/views/ziyanScr/components/ostype-selector/form-cvm-image-selector.vue';
import applicationSideslider from '../application-sideslider/index.vue';
import WName from '@/components/w-name';
import HostApplyTips from './host-apply-tips/common-tips.vue';
import HostApplySpringPoolTips from './host-apply-tips/spring-pool-tips.vue';
import ReqTypeValue from '@/components/display-value/req-type-value.vue';
import DeviceTypeCvmSelector from '@/components/device-type-selector/cvm-apply/cvm-apply.vue';
import {
  MENU_SERVICE_HOST_APPLICATION,
  MENU_BUSINESS_TICKET_MANAGEMENT,
  MENU_BUSINESS_HOST_MANAGEMENT,
} from '@/constants/menu-symbol';
import { useAccountStore, useUserStore } from '@/store';
import usePlanStore from '@/store/usePlanStore';
import useCvmChargeType from '@/views/ziyanScr/hooks/use-cvm-charge-type';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import useFormModel from '@/hooks/useFormModel';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import apiService from '@/api/scrApi';
import { VendorEnum, GLOBAL_BIZS_KEY } from '@/common/constant';
import { VerifyStatus, VerifyStatusMap } from './constants';
import { ChargeType } from '@/typings/plan';
import { cloneDeep } from 'lodash';
import { timeFormatter, expectedDeliveryTime } from '@/common/util';
import http from '@/http';

import { useItDeviceType } from '@/views/ziyanScr/cvm-produce/component/create-order/use-it-device-type';
import { ICvmSystemDisk } from '@/views/ziyanScr/components/cvm-system-disk/typings';
// 滚服项目
import RollingServerTips from './host-apply-tips/rolling-server-tips.vue';
import RollingServerCpuCoreLimits from '@/views/ziyanScr/rolling-server/cpu-core-limits/index.vue';
// 小额绿通
import GreenChannelTips from './host-apply-tips/green-channel-tips.vue';
import GreenChannelCpuCoreLimits from './green-channel/cpu-core-limits.vue';
// 机房裁撤
import DissolveCpuCoreLimits from './dissolve/cpu-core-limits.vue';
import ShortRentalTips from './host-apply-tips/short-rental-tips.vue';

import type { ICvmDeviceTypeFormData } from '@/components/device-type-selector/typings';
import { RequirementType } from '@/store/config/requirement';
import { useAffinityCheck } from './affinity-check/use-affinity-check';
import AffinityCheckResultDialog from './affinity-check/affinity-check-result-dialog.vue';

// 资源利用率
import ResourceUsageRateCard from '@/components/resource-usage-rate/resource-usage-rate-card.vue';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;
const { DropdownMenu, DropdownItem } = Dropdown;

export default defineComponent({
  components: {
    applicationSideslider,
  },
  props: {
    isbusiness: {
      type: Boolean,
      default: false,
    },
  },
  setup(props) {
    const accountStore = useAccountStore();

    const { cvmChargeTypes } = useCvmChargeType();

    const IDCPMformRef = ref();
    const QCLOUDCVMformRef = ref();

    const networkInfoPanelRef = useTemplateRef<typeof NetworkInfoCollapsePanel>('network-info-panel');

    const router = useRouter();
    const route = useRoute();
    const addResourceRequirements = ref(false);
    const isLoading = ref(false);
    const title = ref('增加资源需求');
    const CVMapplication = ref(false);
    const { getBizsId, whereAmI, isBusinessPage } = useWhereAmI();
    const planStore = usePlanStore();
    const isNeedVerfiy = ref(false);
    const isVerifyFailed = ref(false);
    const {
      formModel: cpuAmount,
      setFormValues: setCpuAmount,
      resetForm: resetCpuAmount,
    } = useFormModel({
      prepaid: 0,
      postpaid: 0,
    });
    const order = ref({
      loading: false,
      submitting: false,
      saving: false,
      model: {
        bkBizId: undefined as number,
        bkUsername: '',
        requireType: 1,
        enableNotice: false,
        expectTime: expectedDeliveryTime(),
        remark: '',
        follower: [] as any,
        suborders: [] as any,
      },
      rules: {
        bkBizId: [{ required: true, message: '请选择业务', trigger: 'change' }],
        requireType: [{ required: true, message: '请选择需求类型', trigger: 'change' }],
        expectTime: [{ required: true, message: '请填写期望交付时间', trigger: 'change' }],
      },
    });

    // 添加按钮侧边栏公共表单对象
    const resourceForm = ref({
      resourceType: 'QCLOUDCVM', // 主机类型
      remark: '', // 备注
      enable_disk_check: false,
      region: '', // 地域
      zone: '', // 园区
      zones: [], // 园区，cvm支持多可用区，使用此字段
      charge_type: cvmChargeTypes.PREPAID,
      charge_months: 36, // 计费时长
      bk_asset_id: '', // 继承套餐的机器代表固资号
    });
    // 侧边栏腾讯云CVM
    const QCLOUDCVMForm = ref({
      spec: {
        device_type: '', // 机型
        replicas: 1, // 需求数量
        anti_affinity_level: 'ANTI_NONE',
        vpc: '', //  vpc
        subnet: '', //  子网
        image_id: '', // 镜像
        system_disk: { disk_type: '', disk_size: 0, disk_num: 1 },
        data_disk: [],
        network_type: 'TENTHOUSAND',
        inherit_instance_id: '', // 继承套餐的机器代表实例ID
        cpu: undefined,
        res_assign: undefined,
      },
    });

    const {
      affinityCheck,
      affinityCheckResult,
      isLoading: isAffinityCheckLoading,
      isResultDialogShow: isAffinityCheckResultDialogShow,
    } = useAffinityCheck();

    const formRef = ref();
    const IDCPMIndex = ref(-1);
    const QCLOUDCVMIndex = ref(-1);
    const resourceFormRef = ref();
    const dropdownMenuShowState = reactive({
      idc: false,
      cvm: false,
    });
    const { columns: CloudHostcolumns, generateColumnsSettings } = useColumns('CloudHost');
    let cloudHostSetting = generateColumnsSettings(CloudHostcolumns);
    const { columns: PhysicalMachinecolumns } = useColumns('PhysicalMachine');
    const cloudTableColumns = ref([]);

    // 特殊需求类型（滚服项目、小额绿通）-状态
    const isRollingServer = computed(() => order.value.model.requireType === RequirementType.RollServer);
    const isGreenChannel = computed(() => order.value.model.requireType === RequirementType.GreenChannel);
    const isSpringPool = computed(() => order.value.model.requireType === RequirementType.SpringResPool);
    const isShortRental = computed(() => order.value.model.requireType === RequirementType.ShortRental);
    const isDissolve = computed(() => order.value.model.requireType === RequirementType.Dissolve);
    const isRollingServerLike = computed(() => isRollingServer.value || isSpringPool.value);
    const isSpecialRequirement = computed(() => isRollingServer.value || isGreenChannel.value);

    const handleDeviceTypeChange = (
      data: Partial<ICvmDeviceTypeFormData>,
      from: 'confirm' | 'auto',
      changed?: { zones: boolean },
    ) => {
      if (from === 'confirm') {
        const { deviceTypeList, inheritInstanceId, inheritAssetId } = data;
        QCLOUDCVMForm.value.spec.cpu = deviceTypeList?.[0]?.cpu_amount;
        QCLOUDCVMForm.value.spec.inherit_instance_id = inheritInstanceId;
        resourceForm.value.bk_asset_id = inheritAssetId;

        // zones有变更
        if (changed?.zones) {
          QCLOUDCVMForm.value.spec.subnet = '';
          // 选择了多个可用区
          if (data.zones?.length !== 1 || data.zones?.[0] === 'all') {
            QCLOUDCVMForm.value.spec.vpc = '';
          }
        }
      }
    };

    const currentSpecDeviceType = computed(() => QCLOUDCVMForm.value.spec.device_type);
    const { currentCloudInstanceConfig, isItDeviceType } = useItDeviceType(true, currentSpecDeviceType, () => {
      const { region, zones, charge_type: chargeType } = resourceForm.value;
      return { region, zone: zones, chargeType };
    });

    const PhysicalMachineoperation = ref({
      label: '操作',
      width: 100,
      render: ({ row, index }: any) => {
        return (
          <div class='operation-column'>
            <Button text theme='primary' class='mr10' onClick={() => clonelist(row, 'IDCPM')}>
              克隆
            </Button>
            <Dropdown
              trigger='manual'
              isShow={IDCPMIndex.value === index && dropdownMenuShowState.idc}
              popoverOptions={{
                renderType: 'shown',
                onAfterHidden: () => {
                  IDCPMIndex.value = -1;
                  dropdownMenuShowState.idc = false;
                },
                forceClickoutside: true,
              }}>
              {{
                default: () => (
                  <div
                    class={`more-action${IDCPMIndex.value === index ? ' current-operate-row' : ''}`}
                    onClick={() => {
                      IDCPMIndex.value = index;
                      dropdownMenuShowState.idc = true;
                    }}>
                    <i class='hcm-icon bkhcm-icon-more-fill' />
                  </div>
                ),
                content: () => (
                  <DropdownMenu>
                    <DropdownItem
                      key='retry'
                      onClick={() => {
                        modifylist(row, index, 'IDCPM');
                        dropdownMenuShowState.idc = false;
                      }}>
                      修改
                    </DropdownItem>
                    <DropdownItem
                      key='stop'
                      onClick={() => {
                        deletelist(index, 'IDCPM');
                        dropdownMenuShowState.idc = false;
                      }}>
                      删除
                    </DropdownItem>
                  </DropdownMenu>
                ),
              }}
            </Dropdown>
          </div>
        );
      },
    });
    const CloudHostoperation = ref({
      label: '操作',
      fixed: 'right',
      width: 120,
      render: ({ row, index }: any) => {
        return (
          <div class='operation-column'>
            <Button text theme='primary' class='mr10' onClick={() => clonelist(row, 'QCLOUDCVM')}>
              克隆
            </Button>
            <Dropdown
              trigger='manual'
              isShow={QCLOUDCVMIndex.value === index && dropdownMenuShowState.cvm}
              popoverOptions={{
                renderType: 'shown',
                onAfterHidden: () => {
                  QCLOUDCVMIndex.value = -1;
                  dropdownMenuShowState.cvm = false;
                },
                forceClickoutside: true,
              }}>
              {{
                default: () => (
                  <div
                    class={`more-action${QCLOUDCVMIndex.value === index ? ' current-operate-row' : ''}`}
                    onClick={() => {
                      QCLOUDCVMIndex.value = index;
                      dropdownMenuShowState.cvm = true;
                    }}>
                    <i class='hcm-icon bkhcm-icon-more-fill' />
                  </div>
                ),
                content: () => (
                  <DropdownMenu>
                    <DropdownItem
                      key='retry'
                      onClick={() => {
                        modifylist(row, index, 'QCLOUDCVM');
                        dropdownMenuShowState.cvm = false;
                      }}>
                      修改
                    </DropdownItem>
                    <DropdownItem
                      key='stop'
                      onClick={() => {
                        deletelist(index, 'QCLOUDCVM');
                        dropdownMenuShowState.cvm = false;
                      }}>
                      删除
                    </DropdownItem>
                  </DropdownMenu>
                ),
              }}
            </Dropdown>
          </div>
        );
      },
    });
    const CVMVerifyColumns = [
      {
        field: 'verify_result',
        label: '预检状态',
        minWidth: 90,
        isDefaultShow: true,
        render({ cell }: { cell: VerifyStatus }) {
          return <span class={`status-${cell}`}>{VerifyStatusMap[cell] || '待校验'}</span>;
        },
        isHidden: isSpecialRequirement.value,
      },
      {
        field: 'reason',
        label: '预检信息',
        minWidth: 150,
        isDefaultShow: true,
        render({ cell }: { cell: string }) {
          return <span v-bk-tooltips={{ content: cell, disabled: !cell?.length }}>{cell || '--'}</span>;
        },
        isHidden: isSpecialRequirement.value,
        showOverflowTooltip: false,
      },
    ];

    // 主机类型列表
    const resourceTypes = ref([
      {
        value: 'IDCPM',
        label: 'IDC_物理机',
      },

      {
        value: 'QCLOUDCVM',
        label: '腾讯云_CVM',
      },
    ]);

    // 侧边栏物理机CVM
    const pmForm = ref({
      spec: {
        device_type: '', // 机型
        raid_type: '', // RAID 类型
        os_type: '', // 操作系统
        isp: '', // 经营商
        anti_affinity_level: '',
        network_type: 'TENTHOUSAND',
        replicas: 1, // 需求数量
      },
      rules: {
        device_type: [{ required: true, message: '请选择机型', trigger: 'change' }],
        region: [{ required: true, message: '请选择地域', trigger: 'change' }],
        replicas: [{ required: true, message: '请输入需求数量', trigger: 'blur' }],
        os_type: [{ required: true, message: '请选择操作系统', trigger: 'change' }],
        anti_affinity_level: [{ required: true, message: '请选择反亲和性', trigger: 'change' }],
      },
      options: {
        deviceTypes: [],
        osTypes: [],
        raidType: [],
        regions: [],
        zones: [],
        isps: [],
      },
    });
    // 云组件table
    const cloudTableData = ref([]);
    // 物理机table
    const physicalTableData = ref([]);

    // 云地域变更时, 清空zone, vpc, subnet, device_type
    const handleRegionChange = () => {
      if (resourceForm.value.resourceType === 'QCLOUDCVM') {
        QCLOUDCVMForm.value.spec.vpc = '';
        resourceForm.value.zones = [];
      } else {
        resourceForm.value.zone = '';
        pmForm.value.spec.device_type = '';
      }
    };

    const onQcloudAffinityChange = (val: any) => {
      pmForm.value.spec.anti_affinity_level = val;
    };

    // 获取 IDCPM机型列表
    const IDCPMOsTypes = async () => {
      const { info } = await apiService.getOsTypes();
      pmForm.value.options.osTypes = info || [];
    };
    // 获取 IDCPM机型列表
    const IDCPMIsps = async () => {
      const { info } = await apiService.getIsps();
      pmForm.value.options.isps = info || [];
    };
    // 获取 IDCPM机型列表
    const IDCPMDeviceTypes = async () => {
      const { info } = await apiService.getIDCPMDeviceTypes();
      pmForm.value.options.deviceTypes = info || [];
    };

    // 获取可用的IDCPM列表
    const IDCPMlist = () => {
      IDCPMDeviceTypes();
      IDCPMOsTypes();
      IDCPMIsps();
    };
    const onResourceTypeChange = (resourceType: string) => {
      resourceForm.value.region = '';
      resourceForm.value.zone = '';
      resourceForm.value.zones = [];
      const { osTypes, deviceTypes, isps } = pmForm.value.options;
      if (resourceType === 'IDCPM' && osTypes.length === 0 && deviceTypes.length === 0 && isps.length === 0) {
        IDCPMlist();
      }
    };

    // 监听物理机机型变化
    watch(
      () => pmForm.value.spec.device_type,
      () => {
        // RAID 类型
        pmForm.value.spec.raid_type =
          pmForm.value.options.deviceTypes?.find((item) => item.device_type === pmForm.value.spec.device_type)?.raid ||
          '';
      },
    );

    const clonelist = (originRow: any, resourceType: string) => {
      const cloneRow = cloneDeep(originRow);

      if (resourceType === 'QCLOUDCVM') {
        // 克隆后，需要重新进行需求预检
        Object.assign(cloneRow, { verify_result: '', reason: '' });
        cloudTableData.value.push(cloneRow);
      } else {
        physicalTableData.value.push(cloneRow);
      }
    };

    const modifyindex = ref(0);
    const modifyresourceType = ref('');
    const modifylist = (originRow: any, index: number, resourceType: string) => {
      const cloneRow = cloneDeep(originRow);

      // 控制一键申请的slider是否显示
      CVMapplication.value = false;

      resourceForm.value.resourceType = resourceType;
      modifyresourceType.value = resourceType;

      const { anti_affinity_level, bk_asset_id, remark, replicas, spec } = cloneRow;
      const { region, zone, zones, charge_type, charge_months } = spec;

      Object.assign(resourceForm.value, { bk_asset_id, region, zone, zones, remark });

      if (resourceType === 'QCLOUDCVM') {
        QCLOUDCVMForm.value.spec = { ...spec, anti_affinity_level, replicas: +replicas };
        Object.assign(resourceForm.value, { charge_type, charge_months });
      } else {
        pmForm.value.spec = { ...spec, anti_affinity_level, replicas: +replicas };
      }

      title.value = '修改资源需求';
      modifyindex.value = index;
      addResourceRequirements.value = true;
    };
    const deletelist = (index: number, resourceType: string) => {
      if (resourceType === 'QCLOUDCVM') {
        cloudTableData.value.splice(index, 1);
      } else {
        physicalTableData.value.splice(index, 1);
      }
    };

    const resolveSpecDataDiskInReApply = (spec: any) => {
      const { data_disk, disk_type, disk_size } = spec;
      if (spec.zones === null && spec.zone === 'cvm_separate_campus') {
        spec.zones = ['all'];
      }
      // 兼容旧单据数据
      if (!data_disk) {
        return disk_type ? { ...spec, data_disk: [{ disk_type, disk_size, disk_num: 1 }] } : { ...spec, data_disk: [] };
      }
      return spec;
    };

    const isQueryAutoComplete = ref(false);
    const unReapply = async () => {
      isQueryAutoComplete.value = true;
      // 来源于单据-再次申请
      if (route?.query?.order_id) {
        const data = await apiService.getOrderDetail(+route?.query?.order_id);

        const {
          bk_biz_id: bkBizId,
          bk_username: bkUsername,
          require_type: requireType,
          enable_notice: enableNotice,
          expect_time: expectTime,
          remark,
          follower,
          suborders,
        } = data ?? {};

        order.value.model = {
          bkBizId,
          bkUsername,
          requireType,
          enableNotice,
          expectTime,
          remark,
          follower: follower || [],
          suborders,
        };

        suborders.forEach(({ resource_type, remark, replicas, spec, applied_core }: any) => {
          const data = {
            resource_type,
            remark,
            replicas: +replicas,
            spec: resolveSpecDataDiskInReApply(spec),
            applied_core,
          };

          resource_type === 'QCLOUDCVM' ? cloudTableData.value.push(data) : physicalTableData.value.push(data);
        });
      }

      if (route?.query?.id) {
        assignment(route?.query);

        const { from } = route?.query;
        // 来源为业务/服务下CVM库存一键或资源预测申请时，需要回填需求类型
        if (from === 'businessCvmInventory' || from === 'serviceCvmInventory' || from === 'businessResourcePlan') {
          order.value.model.requireType = Number(route.query.require_type);
          isOneClickApplication.value = true;
        }

        addResourceRequirements.value = true;
      }
    };
    onMounted(() => {
      unReapply();
    });
    watch(
      () => order.value.model.requireType,
      () => {
        device.value.filter.require_type = order.value.model.requireType;
      },
    );
    const device = ref({
      filter: {
        require_type: 1,
        region: [],
        zone: [],
        device_type: [],
        device_group: ['标准型'],
        cpu: '',
        mem: '',
        enable_capacity: true,
      },
      options: {
        require_types: [],
        regions: [],
        zones: [],
        device_groups: ['标准型', '高IO型', '大数据型', '计算型'],
        device_types: [],
        cpu: [],
        mem: [],
      },
      page: {
        limit: 50,
        start: 0,
        total: 0,
        sort: '-capacity_flag',
      },
    });

    // 一键申请按钮点击事件
    const handleApplication = () => {
      CVMapplication.value = true;
      isOneClickApplication.value = true;
    };

    const computedBiz = computed(() => {
      return whereAmI.value === Senarios.business ? getBizsId() : order.value.model.bkBizId;
    });

    const assignment = (data: any) => {
      resourceForm.value.resourceType = 'QCLOUDCVM';
      QCLOUDCVMForm.value.spec = {
        device_type: data.device_type, // 机型
        vpc: '', //  vpc
        subnet: '', //  子网
        replicas: 1,
        anti_affinity_level: 'ANTI_NONE',
        image_id: 'img-fjxtfi0n', // 镜像
        system_disk: { disk_type: '', disk_size: 0, disk_num: 1 },
        data_disk: [],
        network_type: 'TENTHOUSAND',
        inherit_instance_id: '',
        cpu: data.cpu,
        res_assign: data.res_assign,
      };
      resourceForm.value.region = data.region;
      resourceForm.value.zone = data.zone;
      resourceForm.value.zones = data.zone ? [data.zone] : data.zones;
      resourceForm.value.charge_months = 36;
    };

    // 是否处理一键申请模式
    const isOneClickApplication = ref(false);

    // 应用一键申请
    const onOneClickApply = (row: any, val: boolean) => {
      isOneClickApplication.value = true;

      // 控制一键申请slider的显示状态
      CVMapplication.value = val;

      // 初始化配置表单的数据
      assignment(row);

      // 显示配置（添加/修改）资源需求的slider
      title.value = '增加资源需求';
      addResourceRequirements.value = true;
    };

    // 配置（添加/修改）资源需求的slider的close回调
    const ARtriggerShow = (isShow: boolean) => {
      emptyForm();
      addResourceRequirements.value = isShow;

      // 是否继续显示一键申请slider
      CVMapplication.value = isOneClickApplication.value && !isQueryAutoComplete.value;

      nextTick(() => {
        resourceFormRef.value?.clearValidate();
        QCLOUDCVMformRef.value?.clearValidate();
        IDCPMformRef.value?.clearValidate();
      });
    };

    // 一键申请的slider的close回调
    const CAtriggerShow = (isShow: boolean) => {
      CVMapplication.value = isShow;
      isOneClickApplication.value = false;
    };

    const emptyForm = () => {
      resourceForm.value = {
        resourceType: 'QCLOUDCVM',
        region: '', // 地域
        zone: '', // 园区
        zones: [],
        remark: '',
        enable_disk_check: false,
        charge_type: cvmChargeTypes.PREPAID,
        charge_months: 36,
        bk_asset_id: resourceForm.value.bk_asset_id, // 继承套餐的机器固资号不用清除
      };
      QCLOUDCVMForm.value = {
        spec: {
          device_type: '', // 机型
          replicas: 1,
          vpc: '', //  vpc
          subnet: '', //  子网
          anti_affinity_level: 'ANTI_NONE',
          image_id: '', // 镜像
          system_disk: { disk_type: '', disk_size: 0, disk_num: 1 },
          data_disk: [],
          network_type: 'TENTHOUSAND',
          inherit_instance_id: QCLOUDCVMForm.value.spec.inherit_instance_id, // 继承套餐的机器实例id不用清除
          cpu: undefined,
          res_assign: undefined,
        },
      };
      pmForm.value.spec = {
        device_type: '', // 机型
        raid_type: '', // RAID 类型
        os_type: '', // 操作系统
        anti_affinity_level: '',
        replicas: 1,
        isp: '', // 运营商
        network_type: 'TENTHOUSAND',
      };
    };
    const cloudResourceForm = () => {
      const {
        resourceType: resource_type,
        remark,
        enable_disk_check,
        region,
        zone,
        zones,
        charge_type,
        charge_months,
        bk_asset_id,
      } = resourceForm.value;

      return {
        bk_asset_id,
        resource_type,
        remark,
        enable_disk_check,
        anti_affinity_level: QCLOUDCVMForm.value.spec.anti_affinity_level,
        replicas: +QCLOUDCVMForm.value.spec.replicas,
        source: 'business',
        spec: {
          ...QCLOUDCVMForm.value.spec,
          region,
          zone,
          zones,
          charge_type,
          charge_months,
        },
      };
    };
    const PMResourceForm = () => {
      return {
        resource_type: resourceForm.value.resourceType,
        remark: resourceForm.value.remark,
        anti_affinity_level: pmForm.value.spec.anti_affinity_level,
        replicas: +pmForm.value.spec.replicas,
        source: 'business',
        spec: {
          region: resourceForm.value.region,
          zone: resourceForm.value.zone,
          ...pmForm.value.spec,
        },
      };
    };
    const QCLOUDCVMformRules = computed(() => ({
      device_type: [{ required: true, message: '请选择机型', trigger: 'change' }],
      image_id: [{ required: true, message: '请选择镜像', trigger: 'change' }],
      replicas: [
        { required: true, message: '请输入需求数量', trigger: 'blur' },
        // 临时规则双十一后可能需要去除
        {
          validator: (value: number) => !(isRollingServerLike.value && value > 100),
          message: '注意：因云接口限制，单次的机器数最大值为100，超过后请手动克隆为多条配置',
          trigger: 'change',
        },
      ],
      system_disk: [
        {
          validator: (value: ICvmSystemDisk) => !!value.disk_type,
          message: '请选择系统盘类型',
          trigger: 'change',
          required: true,
        },
      ],
      data_disk: [
        {
          validator: (value: { disk_type: string; disk_size: number; disk_num: number }[]) => {
            if (value.length === 0) return true;
            return value.every((item) => item.disk_type && item.disk_size && item.disk_num);
          },
          message: '数据盘信息不能为空',
          trigger: 'change',
        },
      ],
      subnet: [
        {
          validator: (value: string) => (QCLOUDCVMForm.value.spec.vpc ? !!value : true),
          message: '选择 VPC 后必须选择子网',
          trigger: 'change',
        },
      ],
    }));
    const resourceFormrules = ref({
      resourceType: [{ required: true, message: '请选择主机类型', trigger: 'change' }],
      region: [{ required: true, message: '请选择云地域', trigger: 'change' }],
      zone: [{ required: true, message: '请选择可用区', trigger: 'change' }],
    });
    const handleSubmit = async () => {
      await resourceFormRef.value.validate();
      if (resourceForm.value.resourceType === 'QCLOUDCVM') {
        try {
          await QCLOUDCVMformRef.value.validate();
        } catch (error) {
          networkInfoPanelRef.value?.handleToggle(true);
          return Promise.reject(error);
        }
      } else {
        await IDCPMformRef.value.validate();
      }

      if (title.value === '增加资源需求') {
        if (resourceForm.value.resourceType === 'QCLOUDCVM') {
          cloudTableData.value.push(cloudResourceForm());
        } else {
          physicalTableData.value.push(PMResourceForm());
        }
        emptyForm();
      } else {
        if (modifyresourceType.value === 'QCLOUDCVM') {
          if (modifyresourceType.value === resourceForm.value.resourceType) {
            cloudTableData.value[modifyindex.value] = cloudResourceForm();
          } else {
            cloudTableData.value.splice(modifyindex.value, 1);
            physicalTableData.value.push(PMResourceForm());
          }
          emptyForm();
        } else {
          if (modifyresourceType.value === resourceForm.value.resourceType) {
            physicalTableData.value[modifyindex.value] = PMResourceForm();
          } else {
            physicalTableData.value.splice(modifyindex.value, 1);
            cloudTableData.value.push(cloudResourceForm());
          }
          emptyForm();
        }
      }
      modifyindex.value = 0;
      addResourceRequirements.value = false;
      nextTick(() => {
        resourceFormRef.value?.clearValidate();
        QCLOUDCVMformRef.value?.clearValidate();
        IDCPMformRef.value?.clearValidate();
      });
    };
    const isUncommit = computed(() => {
      return route?.query?.order_id && +route?.query?.unsubmitted === 1;
    });
    const handleSaveOrSubmit = async (type: 'save' | 'submit') => {
      await formRef.value.validate();
      const suborders = [...cloudTableData.value, ...physicalTableData.value];
      isLoading.value = true;

      try {
        const basePath = `${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa`;
        const taskPath = type === 'submit' ? 'task/create/apply' : 'task/update/apply/ticket';
        let url = null;
        let bk_biz_id;
        if (props.isbusiness) {
          bk_biz_id = accountStore.bizs;
          url = `${basePath}/bizs/${accountStore.bizs}/${taskPath}`;
        } else {
          bk_biz_id = order.value.model.bkBizId === 'all' ? undefined : order.value.model.bkBizId;
          url = `${basePath}/${taskPath}`;
        }
        await http.post(url, {
          order_id: isUncommit.value ? +route?.query.order_id : undefined,
          bk_biz_id,
          bk_username: useUserStore().username,
          require_type: order.value.model.requireType,
          expect_time: timeFormatter(order.value.model.expectTime),
          remark: order.value.model.remark,
          follower: order.value.model.follower,
          suborders,
        });
        const message = `${type === 'submit' ? '申请成功' : '保存成功'}`;
        Message({
          theme: 'success',
          message,
        });
        // 合代码之后完善跳转路由
        if (props.isbusiness) {
          router.replace({
            name: MENU_BUSINESS_TICKET_MANAGEMENT,
            query: { [GLOBAL_BIZS_KEY]: bk_biz_id, type: 'host_apply' },
          });
        } else {
          router.replace({ name: MENU_SERVICE_HOST_APPLICATION });
        }
      } finally {
        isLoading.value = false;
      }
    };

    const handleVerify = async () => {
      await formRef.value.validate();
      const suborders = cloudTableData.value;
      isLoading.value = true;
      try {
        const { data } = await planStore.verify_resource_demand({
          bk_biz_id: !isSpringPool.value ? +computedBiz.value : 931,
          require_type: order.value.model.requireType,
          suborders,
        });
        for (let i = 0; i < cloudTableData.value.length; i++) {
          cloudTableData.value[i].verify_result = data.verifications[i].verify_result;
          cloudTableData.value[i].reason = data.verifications[i].reason;
        }
        isNeedVerfiy.value = data.verifications.reduce((acc, cur) => {
          acc ||= cur.verify_result !== 'PASS';
          return acc;
        }, false);
        Message({
          theme: isNeedVerfiy.value ? 'warning' : 'success',
          message: isNeedVerfiy.value ? '校验不通过' : '校验通过',
        });
      } finally {
        isLoading.value = false;
      }
    };

    const handleAffinityCheck = () => {
      affinityCheck({
        bk_biz_id: !isSpringPool.value ? +computedBiz.value : 931,
        specs: cloudTableData.value.map((item) => ({
          region: item.spec.region,
          zones: item.spec.zones,
          device_type: item.spec.device_type,
          replicas: item.spec.replicas,
        })),
      });
    };

    const handleCancel = () => {
      if (props.isbusiness) {
        router.push({
          name: MENU_BUSINESS_HOST_MANAGEMENT,
        });
      } else {
        router.go(-1);
      }
    };

    // vpc变更时，置空subnet
    const handleVpcChange = () => {
      QCLOUDCVMForm.value.spec.subnet = '';
    };

    watch(
      isSpecialRequirement,
      (val) => {
        if (!val) {
          resourceForm.value.bk_asset_id = '';
          QCLOUDCVMForm.value.spec.inherit_instance_id = '';
          cloudTableColumns.value = [...CloudHostcolumns, ...CVMVerifyColumns, CloudHostoperation.value];
          cloudHostSetting = generateColumnsSettings(cloudTableColumns.value);
          return;
        }
        // 滚服项目、小额绿通, 默认为云主机, 禁用选择
        resourceForm.value.resourceType = 'QCLOUDCVM';
        // 动态更新云主机列字段
        cloudTableColumns.value = [...CloudHostcolumns, CloudHostoperation.value];
        cloudHostSetting = generateColumnsSettings(cloudTableColumns.value);
      },
      {
        immediate: true,
      },
    );

    // 清空配置清单
    const clearResRequirements = () => {
      cloudTableData.value = [];
      physicalTableData.value = [];
    };

    const isOnlyPhysicalRequirements = computed(() => {
      return cloudTableData.value.length === 0 && physicalTableData.value.length > 0;
    });

    const isRequirementsEmpty = computed(() => {
      return !cloudTableData.value.length && !physicalTableData.value.length;
    });

    // 需求核数
    const replicasCpuCores = computed(() =>
      cloudTableData.value.reduce((prev, curr) => {
        const { replicas, spec, applied_core } = curr;
        // 如果 applied_core(接口值) 有值，则优先使用；若无值，则使用前端计算值
        if (applied_core !== undefined) return prev + applied_core;

        return prev + replicas * spec.cpu;
      }, 0),
    );

    // 滚服项目、小额绿通 - cpu需求限额
    const rollingServerCpuCoreLimitsRef = useTemplateRef<typeof RollingServerCpuCoreLimits>(
      'rolling-server-cpu-core-limits',
    );
    const availableCpuCoreQuota = computed(() => {
      let val = 0;
      if (isRollingServerLike.value) val = rollingServerCpuCoreLimitsRef.value?.availableCpuCoreQuota ?? val;
      if (isGreenChannel.value) val = greenChannelCpuCoreLimitsRef.value?.availableCpuCoreQuota ?? val;
      if (isDissolve.value) val = dissolveCpuCoreLimitsRef.value?.availableCpuCoreQuota ?? val;
      return val;
    });
    const isCpuCoreExceeded = computed(() => replicasCpuCores.value > availableCpuCoreQuota.value);

    // 小额绿通
    const greenChannelCpuCoreLimitsRef = useTemplateRef<typeof GreenChannelCpuCoreLimits>(
      'green-channel-cpu-core-limits',
    );
    // 机房裁撤
    const dissolveCpuCoreLimitsRef = useTemplateRef<typeof DissolveCpuCoreLimits>('dissolve-cpu-core-limits');

    const addButtonDisabledState = computed(() => {
      let disabled = false;
      let content = '';
      if (
        (isRollingServer.value || isGreenChannel.value || isSpringPool.value || isDissolve.value) &&
        availableCpuCoreQuota.value <= 0
      ) {
        content = `已超过${
          // eslint-disable-next-line no-nested-ternary
          isRollingServerLike.value
            ? isRollingServer.value
              ? '滚服项目'
              : '春保资源池'
            : isGreenChannel.value
            ? '小额绿通'
            : '机房裁撤'
        }的CPU可用额度，不允许添加`;
        disabled = true;
      }
      return { content, disabled };
    });

    const applyButtonDisabledState = computed(() => {
      if (isRollingServer.value) {
        return { disabled: true, content: '滚服项目暂不支持一键申请' };
      }
      if ((isGreenChannel.value || isSpringPool.value || isDissolve.value) && availableCpuCoreQuota.value <= 0) {
        return { disabled: true, content: '已超过CPU可用额度，不允许添加' };
      }
      return { disabled: false, content: '' };
    });

    const submitButtonDisabledState = computed(() => {
      if (isRequirementsEmpty.value) {
        return { disabled: true, content: '资源需求不能为空' };
      }
      if ((isRollingServer.value || isGreenChannel.value || isDissolve.value) && isCpuCoreExceeded.value) {
        let name = '滚服项目';
        if (isGreenChannel.value) {
          name = '小额绿通';
        } else if (isDissolve.value) {
          name = '机房裁撤';
        }
        return { disabled: true, content: `当前所需的CPU总核数超过${name}CPU限额，请调整后再重试。` };
      }
      return { disabled: false, content: '' };
    });

    const affinityCheckButtonDisabledState = computed(() => {
      if (isRequirementsEmpty.value) {
        return { disabled: true, content: '资源需求不能为空' };
      }
      if (isOnlyPhysicalRequirements.value) {
        return { disabled: true, content: '暂不支持物理机亲和性检查' };
      }
      return { disabled: false, content: '' };
    });

    watch(
      () => cloudTableData.value,
      async (val) => {
        resetCpuAmount();
        for (const item of val) {
          const { replicas, spec, applied_core } = item;
          const { cpu, charge_type } = spec;

          // 如果 applied_core(接口值) 有值，则优先使用；若无值，则使用前端计算值
          if (ChargeType.POSTPAID_BY_HOUR === charge_type) {
            const postpaid =
              applied_core !== undefined ? cpuAmount.postpaid + applied_core : cpuAmount.postpaid + cpu * replicas;
            setCpuAmount({ ...cpuAmount, postpaid });
          }
          if (ChargeType.PREPAID === charge_type) {
            const prepaid =
              applied_core !== undefined ? cpuAmount.prepaid + applied_core : cpuAmount.prepaid + cpu * replicas;
            setCpuAmount({ ...cpuAmount, prepaid });
          }
        }
        isNeedVerfiy.value = val.reduce((acc, cur) => {
          acc ||= cur.verify_result !== 'PASS';
          return acc;
        }, false);
        isVerifyFailed.value = val.reduce((acc, cur) => {
          acc ||= cur.verify_result === 'FAILED';
          return acc;
        }, false);
      },
      {
        deep: true,
      },
    );

    return () => (
      <div class='host-application-form-wrapper'>
        {!props.isbusiness && <DetailHeader>新增申请</DetailHeader>}
        <div class={props.isbusiness ? '' : 'apply-form-wrapper'}>
          {/* 申请单据表单 */}
          <bk-form
            form-type='vertical'
            label-width='150'
            model={order.value.model}
            rules={order.value.rules}
            ref={formRef}>
            <CommonCard title={() => '基本信息'} class='mb12'>
              <div class='basic-top-row'>
                {!props.isbusiness && (
                  <bk-form-item label='所属业务' required property='bkBizId' class='mr24'>
                    <BusinessSelector
                      class='item-warp-component'
                      v-model={order.value.model.bkBizId}
                      autoSelect
                      authed
                      url-key='scr_apply_host_bizs'
                      apiMethod={apiService.getCvmApplyAuthBizList}
                      base64Encode
                    />
                  </bk-form-item>
                )}

                <bk-form-item label='需求类型' required property='requireType'>
                  <hcm-form-req-type
                    appearance='card'
                    v-model={order.value.model.requireType}
                    // 春保资源池不显示
                    filter={(list: any) => list.filter((item: any) => item.require_type !== 8)}
                    onChange={() => {
                      // 手动更改时，需要清空已保存的需求
                      clearResRequirements();
                    }}
                  />
                </bk-form-item>
                <div class='alert-content'>
                  {(function () {
                    if (isRollingServer.value) return <RollingServerTips />;
                    if (isGreenChannel.value) return <GreenChannelTips />;
                    if (isSpringPool.value) return <HostApplySpringPoolTips />;
                    if (isShortRental.value) return <ShortRentalTips />;
                    return <HostApplyTips requireType={order.value.model.requireType} />;
                  })()}
                </div>
              </div>
              <div class='flex-row align-content-center'>
                <bk-form-item
                  label='期望交付时间'
                  required
                  property='expectTime'
                  class='mr24'
                  description={() => (
                    <span>
                      期望申领时间默认为当月，在资源预测额度充足时，过单后会立即申领。如希望审批时按指定时间过单后生产，请联系
                      <WName name={'ICR'} />
                      (IEG资源服务助手)确认{' '}
                    </span>
                  )}>
                  <bk-date-picker
                    class='item-warp-component'
                    v-model={order.value.model.expectTime}
                    clearable
                    type='datetime'></bk-date-picker>
                </bk-form-item>
                <bk-form-item label='关注人'>
                  <hcm-form-user class='item-warp-component' v-model={order.value.model.follower} />
                </bk-form-item>
              </div>
            </CommonCard>

            <ResourceUsageRateCard class='mb12' bizId={computedBiz.value} />

            <CommonCard
              title={() => (
                <div class='flex-row align-items-center'>
                  <span class='mr5'>配置清单</span>
                  <i
                    class={'hcm-icon bkhcm-icon-info-line'}
                    v-bk-tooltips={{
                      content: (
                        <div>
                          <div>自研云主机购买，经过以下步骤后交付给业务</div>
                          <div>1.提交参数后，云梯生产主机</div>
                          <div>2.资源平台对系统初始化，包括GSE agent安装，磁盘格式化等</div>
                          <div>3.转交到业务</div>
                        </div>
                      ),
                    }}></i>
                </div>
              )}
              class='mb12 config-ticket-card'>
              <div class='mb12 tools-wrapper'>
                <Button
                  class='button'
                  theme='primary'
                  outline
                  onClick={() => {
                    addResourceRequirements.value = true;
                    isOneClickApplication.value = false;
                    title.value = '增加资源需求';
                    IDCPMlist();
                  }}
                  loading={dissolveCpuCoreLimitsRef.value?.cpuCoreSummaryLoading}
                  disabled={
                    !dissolveCpuCoreLimitsRef.value?.cpuCoreSummaryLoading && addButtonDisabledState.value.disabled
                  }
                  v-bk-tooltips={{
                    content: addButtonDisabledState.value.content,
                    disabled:
                      dissolveCpuCoreLimitsRef.value?.cpuCoreSummaryLoading || !addButtonDisabledState.value.disabled,
                  }}>
                  <Plus class={'prefix-icon'} />
                  添加
                </Button>
                <Button
                  class='button'
                  onClick={handleApplication}
                  loading={dissolveCpuCoreLimitsRef.value?.cpuCoreSummaryLoading}
                  disabled={
                    !dissolveCpuCoreLimitsRef.value?.cpuCoreSummaryLoading && applyButtonDisabledState.value.disabled
                  }
                  v-bk-tooltips={{
                    content: applyButtonDisabledState.value.content,
                    disabled:
                      dissolveCpuCoreLimitsRef.value?.cpuCoreSummaryLoading || !applyButtonDisabledState.value.disabled,
                  }}>
                  一键申请
                </Button>
                {/* 滚服项目-cpu需求限额，春保资源池复用滚服 */}
                {isRollingServerLike.value && (
                  <RollingServerCpuCoreLimits
                    ref='rolling-server-cpu-core-limits'
                    bizId={computedBiz.value}
                    replicasCpuCores={replicasCpuCores.value}
                  />
                )}
                {/* 小额绿通-cpu需求限额 */}
                {isGreenChannel.value && (
                  <GreenChannelCpuCoreLimits
                    ref='green-channel-cpu-core-limits'
                    replicasCpuCores={replicasCpuCores.value}
                    bizId={computedBiz.value}
                  />
                )}
                {/* 机房裁撤-cpu需求限额 */}
                {isDissolve.value && (
                  <DissolveCpuCoreLimits
                    ref='dissolve-cpu-core-limits'
                    replicasCpuCores={replicasCpuCores.value}
                    bizId={computedBiz.value}
                    isBusinessPage={isBusinessPage}
                  />
                )}
              </div>
              <bk-form-item label='云主机'>
                <p class={'statistics'}>
                  <span class={'label'}>包年包月CPU总数：</span>
                  <span class={'value'}>{cpuAmount.prepaid}核</span>
                  <span class={'ml24 label'}>按量计费CPU总数：</span>
                  <span class={'value'}>{cpuAmount.postpaid}核</span>
                </p>
                <bk-table
                  align='left'
                  row-hover='auto'
                  columns={cloudTableColumns.value}
                  settings={cloudHostSetting.value}
                  data={cloudTableData.value}
                  show-overflow-tooltip
                />
              </bk-form-item>
              {physicalTableData.value.length > 0 && (
                <bk-form-item label='物理机'>
                  <bk-table
                    align='left'
                    row-hover='auto'
                    columns={[...PhysicalMachinecolumns, PhysicalMachineoperation.value]}
                    data={physicalTableData.value}
                    show-overflow-tooltip
                  />
                </bk-form-item>
              )}
            </CommonCard>

            <CommonCard title={() => '备注'}>
              <bk-form-item label='申请备注'>
                <Input
                  type='textarea'
                  v-model={order.value.model.remark}
                  rows={3}
                  maxlength={255}
                  resize={false}
                  placeholder='请输入申请单备注'></Input>
              </bk-form-item>
            </CommonCard>

            {!isSpecialRequirement.value && isVerifyFailed.value && (
              <CommonCard title={() => '需求预检'}>
                <Alert theme='danger' showIcon={false} class={'mb24'}>
                  资源需求超过资源预测的剩余额度，请查看预检信息的报错明细，处理建议：
                  <br />
                  1.调整所需的资源，修改机型或者调整需求数量
                  <br />
                  2.增加资源预测报备后再重试，去
                  <Button
                    theme='primary'
                    text
                    onClick={() => window.open(`#/business/resource-plan?bizs=${computedBiz.value}`, '_blank')}>
                    查看资源预测
                  </Button>
                </Alert>
              </CommonCard>
            )}

            <bk-form-item class={'mt16 form-button-row'}>
              {!isSpecialRequirement.value ? (
                // 非滚服、非小额绿通
                <>
                  {!!cloudTableData.value.length && isNeedVerfiy.value ? (
                    <Button class='mr16' theme='primary' loading={isLoading.value} onClick={handleVerify}>
                      需求校验
                    </Button>
                  ) : (
                    <>
                      <Button
                        class='mr16'
                        theme='primary'
                        loading={isLoading.value}
                        disabled={submitButtonDisabledState.value.disabled}
                        v-bk-tooltips={{
                          content: submitButtonDisabledState.value.content,
                          disabled: !submitButtonDisabledState.value.disabled,
                        }}
                        onClick={() => handleSaveOrSubmit('submit')}>
                        提交
                      </Button>
                      <Button
                        class={'mr16'}
                        loading={isLoading.value}
                        disabled={!physicalTableData.value.length && !cloudTableData.value.length}
                        v-bk-tooltips={{
                          content: '资源需求不能为空',
                          disabled: physicalTableData.value.length || cloudTableData.value.length,
                        }}
                        onClick={() => handleSaveOrSubmit('save')}>
                        保存
                      </Button>
                    </>
                  )}
                </>
              ) : (
                // 滚服、小额绿通
                <>
                  <Button
                    class='mr16'
                    theme='primary'
                    loading={isLoading.value}
                    disabled={submitButtonDisabledState.value.disabled}
                    v-bk-tooltips={{
                      content: submitButtonDisabledState.value.content,
                      disabled: !submitButtonDisabledState.value.disabled,
                    }}
                    onClick={() => handleSaveOrSubmit('submit')}>
                    提交
                  </Button>
                  {/* 滚服、小额绿通不支持保存 */}
                  <Button
                    class={'mr16'}
                    loading={isLoading.value}
                    disabled={true}
                    v-bk-tooltips={{ content: `${isRollingServer.value ? '滚服项目' : '小额绿通'}暂不支持保存` }}>
                    保存
                  </Button>
                </>
              )}

              <Button
                class='mr16'
                loading={isAffinityCheckLoading.value}
                disabled={affinityCheckButtonDisabledState.value.disabled}
                v-bk-tooltips={{
                  content: affinityCheckButtonDisabledState.value.content,
                  disabled: !affinityCheckButtonDisabledState.value.disabled,
                }}
                onClick={handleAffinityCheck}>
                亲和性检查
              </Button>

              <Button onClick={handleCancel}>取消</Button>
            </bk-form-item>
          </bk-form>

          {/* 亲和性检查结果 */}
          <AffinityCheckResultDialog
            v-model:isShow={isAffinityCheckResultDialogShow.value}
            result={affinityCheckResult.value}
          />

          {/* 增加/修改资源需求 */}
          <Sideslider
            class='add-resource-requirements-sideslider'
            width={1200}
            isShow={addResourceRequirements.value}
            title={title.value}
            onClosed={() => {
              ARtriggerShow(false);
            }}>
            {{
              default: () => (
                <div class={'sideslider-layer'}>
                  <CommonCard title={() => '基本信息'}>
                    <bk-form
                      model={resourceForm.value}
                      rules={resourceFormrules}
                      ref={resourceFormRef}
                      form-type='vertical'>
                      <bk-form-item label='主机类型' required property='resourceType'>
                        <bk-select
                          class={'selection-box'}
                          v-model={resourceForm.value.resourceType}
                          onChange={onResourceTypeChange}
                          // 滚服项目、小额绿通只支持云主机
                          disabled={isSpecialRequirement.value || isRollingServerLike.value}
                          v-bk-tooltips={{
                            content: `${
                              // eslint-disable-next-line no-nested-ternary
                              isRollingServerLike.value
                                ? isRollingServer.value
                                  ? '滚服项目'
                                  : '春保资源池'
                                : '小额绿通'
                            }只支持云主机`,
                            disabled: !(isSpecialRequirement.value || isRollingServerLike.value),
                          }}>
                          {resourceTypes.value.map((resType: { value: any; label: any }) => (
                            <bk-option key={resType.value} value={resType.value} label={resType.label}></bk-option>
                          ))}
                        </bk-select>
                      </bk-form-item>
                      <bk-form-item label='云地域' required property='region'>
                        <AreaSelector
                          class={'selection-box'}
                          v-model={resourceForm.value.region}
                          params={{ resourceType: resourceForm.value.resourceType }}
                          popoverOptions={{ boundary: 'parent' }}
                          onChange={handleRegionChange}></AreaSelector>
                      </bk-form-item>
                      {resourceForm.value.resourceType === 'IDCPM' && (
                        <bk-form-item label='可用区' required property='zone'>
                          <ZoneTagSelector
                            class={'selection-box'}
                            key={resourceForm.value.region}
                            style={{ width: '760px' }}
                            v-model={resourceForm.value.zone}
                            vendor={VendorEnum.ZIYAN}
                            region={resourceForm.value.region}
                            resourceType={resourceForm.value.resourceType}
                            separateCampus={true}
                            emptyText='请先选择云地域'
                            minWidth={184}
                            maxWidth={184}
                            autoExpand={'selected'}
                          />
                        </bk-form-item>
                      )}
                    </bk-form>
                  </CommonCard>

                  <CommonCard
                    title={() => '实例配置'}
                    class={resourceForm.value.resourceType === 'IDCPM' ? 'idcpm-card' : 'not-idcpm-card'}>
                    <>
                      {resourceForm.value.resourceType !== 'IDCPM' ? (
                        <>
                          <bk-form
                            model={QCLOUDCVMForm.value.spec}
                            rules={QCLOUDCVMformRules.value}
                            ref={QCLOUDCVMformRef}
                            form-type='vertical'>
                            <bk-form-item label='机型配置' required property='device_type'>
                              {addResourceRequirements.value && (
                                <DeviceTypeCvmSelector
                                  v-model={QCLOUDCVMForm.value.spec.device_type}
                                  v-model:zones={resourceForm.value.zones}
                                  v-model:chargeType={resourceForm.value.charge_type}
                                  v-model:chargeMonths={resourceForm.value.charge_months}
                                  v-model:res-assign-type={QCLOUDCVMForm.value.spec.res_assign}
                                  bizId={computedBiz.value}
                                  vendor={VendorEnum.ZIYAN}
                                  requireType={order.value.model.requireType}
                                  region={resourceForm.value.region}
                                  assetId={resourceForm.value.bk_asset_id}
                                  instanceId={QCLOUDCVMForm.value.spec.inherit_instance_id}
                                  disabled={resourceForm.value.region === ''}
                                  isEditing={isOneClickApplication.value || title.value === '修改资源需求'}
                                  editMode={false}
                                  onChange={handleDeviceTypeChange}></DeviceTypeCvmSelector>
                              )}
                            </bk-form-item>
                            <bk-form-item label='镜像' required property='image_id'>
                              <FormCvmImageSelector
                                class={'commonCard-form-select'}
                                v-model={QCLOUDCVMForm.value.spec.image_id}
                                region={[resourceForm.value.region]}
                                disabled={resourceForm.value.region === ''}
                                popoverOptions={{ boundary: 'parent' }}
                              />
                            </bk-form-item>
                            <bk-form-item property='system_disk' required>
                              {{
                                label: () => (
                                  <>
                                    系统盘
                                    <i
                                      class='hcm-icon bkhcm-icon-prompt text-gray cursor ml4'
                                      v-bk-tooltips={{ content: '系统盘大小范围为50G-1000G' }}></i>
                                  </>
                                ),
                                default: () => (
                                  <CvmSystemDisk
                                    v-model={QCLOUDCVMForm.value.spec.system_disk}
                                    isItDeviceType={isItDeviceType.value}
                                    currentCloudInstanceConfig={currentCloudInstanceConfig.value}
                                  />
                                ),
                              }}
                            </bk-form-item>
                            <bk-form-item property='data_disk'>
                              {{
                                label: () => (
                                  <>
                                    数据盘
                                    <i
                                      class='hcm-icon bkhcm-icon-prompt text-gray cursor ml4'
                                      v-bk-tooltips={{ content: '数据盘大小范围为20G-32000G，且为10的倍数' }}></i>
                                  </>
                                ),
                                default: () => (
                                  <CvmDataDisk
                                    v-model={QCLOUDCVMForm.value.spec.data_disk}
                                    currentCloudInstanceConfig={currentCloudInstanceConfig.value}
                                  />
                                ),
                              }}
                            </bk-form-item>
                            <bk-form-item label='需求数量' required property='replicas'>
                              <hcm-form-number
                                class='commonCard-form-select'
                                v-model={QCLOUDCVMForm.value.spec.replicas}
                                min={1}
                                max={1000}
                              />
                            </bk-form-item>
                            <NetworkInfoCollapsePanel
                              ref={'network-info-panel'}
                              class='network-info-collapse-panel'
                              v-model:vpc={QCLOUDCVMForm.value.spec.vpc}
                              v-model:subnet={QCLOUDCVMForm.value.spec.subnet}
                              region={resourceForm.value.region}
                              zone={resourceForm.value.zones?.[0]}
                              vpcProperty={'vpc'}
                              subnetProperty={'subnet'}
                              disabledVpc={
                                resourceForm.value.zones?.length !== 1 || resourceForm.value.zones?.[0] === 'all'
                              }
                              disabledSubnet={
                                resourceForm.value.zones?.length !== 1 || resourceForm.value.zones?.[0] === 'all'
                              }
                              onChangeVpc={handleVpcChange}
                            />
                            <bk-form-item label='备注'>
                              <Input
                                type='textarea'
                                class={'commonCard-form-select'}
                                rows={3}
                                maxlength={255}
                                v-model={resourceForm.value.remark}
                                autosize
                                resize={false}></Input>
                            </bk-form-item>
                          </bk-form>
                        </>
                      ) : (
                        <>
                          <bk-form
                            model={pmForm.value.spec}
                            rules={pmForm.value.rules}
                            class={'example-form'}
                            ref={IDCPMformRef}
                            form-type='vertical'>
                            <bk-form-item label='机型' required property='device_type'>
                              <bk-select
                                v-model={pmForm.value.spec.device_type}
                                default-first-option
                                class='select-model'
                                filterable>
                                {pmForm.value.options.deviceTypes.map((deviceType: { device_type: any }) => (
                                  <bk-option
                                    key={deviceType.device_type}
                                    value={deviceType.device_type}
                                    label={deviceType.device_type}></bk-option>
                                ))}
                              </bk-select>
                            </bk-form-item>
                            <bk-form-item label='RAID 类型' class={'form-item-raid'}>
                              <div class={'raidText'}> {pmForm.value.spec.raid_type || '-'}</div>
                            </bk-form-item>
                            <bk-form-item label='操作系统' required property='os_type'>
                              <bk-select class={'commonCard-form-select'} v-model={pmForm.value.spec.os_type}>
                                {pmForm.value.options.osTypes.map((osType) => (
                                  <bk-option key={osType} value={osType} label={osType}></bk-option>
                                ))}
                              </bk-select>
                            </bk-form-item>
                            <bk-form-item label='运营商'>
                              <bk-select class={'commonCard-form-select'} v-model={pmForm.value.spec.isp}>
                                <bk-option key='无' value='' label='无'></bk-option>
                                {pmForm.value.options.isps.map((isp) => (
                                  <bk-option key={isp} value={isp} label={isp}></bk-option>
                                ))}
                              </bk-select>
                            </bk-form-item>
                            <div class='commonCard-form'>
                              <bk-form-item label='需求数量' required property='replicas'>
                                <hcm-form-number class='input-demand' v-model={pmForm.value.spec.replicas} min={1} />
                              </bk-form-item>
                              <bk-form-item
                                label='反亲和性'
                                required
                                class={'select-Affinity'}
                                property='anti_affinity_level'>
                                <AntiAffinityLevelSelect
                                  v-model={pmForm.value.spec.anti_affinity_level}
                                  params={{
                                    resourceType: resourceForm.value.resourceType,
                                    hasZone: resourceForm.value.zone !== '',
                                  }}
                                  onAffinitychange={onQcloudAffinityChange}></AntiAffinityLevelSelect>
                              </bk-form-item>
                            </div>
                            <bk-form-item label='备注' property='remark'>
                              <Input
                                class={'commonCard-form-select'}
                                type='textarea'
                                v-model={resourceForm.value.remark}
                                rows={3}
                                maxlength={255}
                                resize={false}
                                placeholder='请输入备注'></Input>
                            </bk-form-item>
                          </bk-form>
                        </>
                      )}
                    </>
                  </CommonCard>
                </div>
              ),
              footer: () => (
                <>
                  <Button theme='primary' onClick={handleSubmit}>
                    保存需求
                  </Button>
                  <Button class='ml16' onClick={() => ARtriggerShow(false)}>
                    取消
                  </Button>
                </>
              ),
            }}
          </Sideslider>

          {/* CVM一键申请 */}
          <Sideslider
            class='common-sideslider cvm-apply-sideslider'
            width={960}
            isShow={CVMapplication.value}
            title='一键申请主机'
            onClosed={() => {
              CAtriggerShow(false);
            }}>
            {{
              header: () => (
                <div class='custom-header-content'>
                  一键申请主机
                  <ReqTypeValue
                    value={device.value.filter.require_type}
                    display={{ appearance: 'tag' }}
                    {...{ theme: 'info' }}
                  />
                </div>
              ),
              default: () => (
                <applicationSideslider
                  isShow={CVMapplication.value}
                  requireType={device.value.filter.require_type}
                  bizId={computedBiz.value}
                  onApply={onOneClickApply}
                />
              ),
            }}
          </Sideslider>
        </div>
      </div>
    );
  },
});
