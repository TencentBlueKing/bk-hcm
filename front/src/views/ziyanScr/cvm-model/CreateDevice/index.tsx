import { defineComponent, ref, reactive, computed } from 'vue';
import './index.scss';
import { Message } from 'bkui-vue';
import AreaSelector from '../../hostApplication/components/AreaSelector';
import ZoneSelector from '../../hostApplication/components/ZoneSelector';
import apiService from '@/api/scrApi';
import ModalFooter from '@/components/modal/modal-footer.vue';

export interface ICvmDeviceCreateModel {
  zone: string;
  device_family: string;
  core_type: '小核心' | '中核心' | '大核心';
  device_type: string;
  cpu_core: number;
  memory: number;
  device_class: string;
  device_type_class: 'SpecialType' | 'CommonType';
  technical_class: string;
  disable?: boolean;
  region: string;
}

export default defineComponent({
  name: 'AllhostInventoryManager',
  props: { isShow: Boolean },
  emits: ['update:isShow', 'submit-success', 'hidden'],
  setup(props, { emit }) {
    const deviceSizeNames = { 小核心: '小核心', 中核心: '中核心', 大核心: '大核心' };
    const deviceFamilyOptions = {
      标准型: '标准型',
      高IO型: '高IO型',
      大数据型: '大数据型',
      计算型: '计算型',
    };
    const deviceTypeClassOptions = {
      CommonType: '通用机型',
      SpecialType: '专用机型',
    };

    const isShow = computed({
      get() {
        return props.isShow;
      },
      set(val) {
        emit('update:isShow', val);
      },
    });

    const formRef = ref();
    const formModel = reactive<Omit<ICvmDeviceCreateModel, 'region' | 'zone'> & { region: string[]; zone: string[] }>({
      zone: [],
      region: [],
      device_family: '标准型',
      core_type: '小核心',
      device_type: '',
      cpu_core: 0,
      memory: 1,
      device_class: '',
      device_type_class: 'CommonType',
      technical_class: '',
    });
    const selectedZones = ref<Array<{ value: string; label: string; region: string }>>([]);

    const handleRegionChange = () => {
      formModel.zone = [];
      selectedZones.value = [];
    };

    const handleZoneChange = (_value: string | string[], detail: any) => {
      if (!detail) {
        selectedZones.value = [];
        return;
      }

      selectedZones.value = detail;
    };

    const isSubmitLoading = ref(false);

    const handleConfirm = async () => {
      await formRef.value.validate();
      isSubmitLoading.value = true;
      try {
        // 以地域为分组组装 device_types
        const regionToZones = new Map<string, string[]>();
        for (const z of selectedZones.value) {
          const r = z.region;
          if (!r) continue;
          const list = regionToZones.get(r) || [];
          list.push(z.value);
          regionToZones.set(r, list);
        }

        const deviceTypes: ICvmDeviceCreateModel[] = [];
        for (const r of formModel.region) {
          const zones = regionToZones.get(r) || [];
          for (const z of zones) {
            deviceTypes.push({
              device_type: formModel.device_type,
              device_class: formModel.device_class,
              device_family: formModel.device_family,
              core_type: formModel.core_type,
              cpu_core: formModel.cpu_core,
              memory: formModel.memory,
              device_type_class: formModel.device_type_class,
              technical_class: formModel.technical_class,
              region: r,
              zone: z,
            });
          }
        }

        const res = await apiService.createCvmDevice({ device_types: deviceTypes }, { globalError: false });

        if (res.code === 0) {
          emit('submit-success');
          isShow.value = false;
        } else {
          Message({ theme: 'error', message: res.message });
        }
      } finally {
        isSubmitLoading.value = false;
      }
    };

    const handleClosed = () => {
      isShow.value = false;
    };

    return () => (
      <bk-dialog
        v-model:isShow={isShow.value}
        class='common-dialog'
        close-icon={false}
        title='创建新机型'
        width={600}
        onHidden={() => emit('hidden')}>
        {{
          default: () => (
            <bk-form model={formModel} ref={formRef} class='cvm-device-create-form'>
              <bk-form-item label='地域' property='region' required>
                <AreaSelector
                  ref='areaSelector'
                  v-model={formModel.region}
                  params={{ resourceType: 'QCLOUDCVM' }}
                  multiple
                  onChange={handleRegionChange}
                  class='i-form-control'
                />
              </bk-form-item>
              <bk-form-item label='园区' property='zone' required>
                <ZoneSelector
                  v-model={formModel.zone}
                  params={{ resourceType: 'QCLOUDCVM', region: formModel.region }}
                  separateCampus={false}
                  multiple
                  onChange={handleZoneChange}
                  class='i-form-control'
                />
              </bk-form-item>
              <bk-form-item label='实例族' property='device_family' required>
                <hcm-form-enum v-model={formModel.device_family} option={deviceFamilyOptions} class='i-form-control' />
              </bk-form-item>
              <bk-form-item label='机型' property='device_type' required>
                <bk-input v-model={formModel.device_type} class='i-form-control' />
              </bk-form-item>
              <bk-form-item label='机型类型' property='device_class' required>
                <bk-input v-model={formModel.device_class} class='i-form-control' />
              </bk-form-item>
              <bk-form-item label='技术分类' property='technical_class' required>
                <bk-input v-model={formModel.technical_class} class='i-form-control' />
              </bk-form-item>
              <bk-form-item label='机型分类' property='device_type_class' required>
                <hcm-form-enum
                  v-model={formModel.device_type_class}
                  option={deviceTypeClassOptions}
                  class='i-form-control'
                />
              </bk-form-item>
              <bk-form-item label='核心类型' property='core_type' required>
                <hcm-form-enum v-model={formModel.core_type} option={deviceSizeNames} class='i-form-control' />
              </bk-form-item>
              <bk-form-item label='CPU(核)' property='cpu_core' required>
                <bk-input type='number' v-model={formModel.cpu_core} min={0} class='i-form-control' />
              </bk-form-item>
              <bk-form-item label='内存(G)' property='memory' required>
                <bk-input type='number' v-model={formModel.memory} min={0} class='i-form-control' />
              </bk-form-item>
            </bk-form>
          ),
          footer: () => (
            <div class='create-device-dialog-footer'>
              <ModalFooter
                disabled={false}
                loading={isSubmitLoading.value}
                onConfirm={handleConfirm}
                onClosed={handleClosed}
              />
            </div>
          ),
        }}
      </bk-dialog>
    );
  },
});
