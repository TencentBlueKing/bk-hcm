import { defineComponent, PropType, ref, watch } from 'vue';
import { SelectColumn, InputColumn } from '@blueking/ediatable';
import './index.scss';
import useFormModel from '@/hooks/useFormModel';
import { isArray } from 'lodash';

export interface IP {
  ip: string;
  mask: number;
  minMask: number;
  maxMask: number;
}

export const SubnetInput = defineComponent({
  props: {
    ips: Object as PropType<{
      idx: number;
      range: Array<IP>;
    }>,
    modelValue: Object as PropType<IP>,
    disabled: Boolean as PropType<boolean>,
    isSub: Boolean,
  },
  emits: ['changeIdx', 'update:modelValue'],
  setup(props, { expose, emit }) {
    // 用户可填的IP的V4四段表示法及范围
    const [o1, o2, o3, o4] = props.ips.range[props.ips.idx].ip.split('.').map((v) => +v);
    const { formModel, setFormValues, resetForm } = useFormModel({
      ip1: o1,
      ip2: o2,
      ip3: o3,
      ip4: o4,
    });

    const { formModel: ipRange, setFormValues: setIpRange, resetForm: resetIpRange } = useFormModel({
      r1: 0 as any,
      r2: 0 as any,
      r3: 0 as any,
      r4: 0 as any,
    });

    // 用户可填的mask及范围
    const newMask = ref(props.modelValue.mask);

    const maskRange = ref({
      min: 1,
      max: 32,
    });

    // 当前用户可填写IP内容的实例
    const o2Ref = ref();
    const o3Ref = ref();
    const o4Ref = ref();

    watch(
      () => props.ips,
      (ips) => {
        const { idx, range } = ips;
        const [o1, o2, o3, o4] = range[idx].ip.split('.').map((v) => +v);
        setFormValues({
          ip1: o1,
          ip2: o2,
          ip3: o3,
          ip4: o4,
        });
        newMask.value = +range[idx].mask;
        maskRange.value = {
          min: range[idx].minMask,
          max: range[idx].maxMask,
        };
      },
      {
        immediate: true,
        deep: true,
      },
    );

    // 根据用户填写的子网掩码确定各个用户可填写位的数值范围
    watch(
      [() => newMask.value, () => props.ips],
      ([num]) => {
        let arr = props.ips.range[props.ips.idx].ip.split('.').map((v) => +v);
        // 双指针，指针中间部分可自由取0、1
        let L = +props.ips.range[props.ips.idx].mask;
        let R = +num;
        if (R < L) return;
        // 4段10进制共4段：0，1，2，3 - 判断左右指针分别在第几段
        let i = Math.floor((L - 1) / 8);
        let j = Math.floor((R - 1) / 8);
        const range = [-1, -1, -1, -1] as any;
        // 双指针窗口外均不能改动
        for (let k = 0; k < range.length; k++) {
          if (k < i || k > j) range[k] = -1;
        }
        // 双指针落在同一段
        if (i === j) {
          range[i] = new Array(2 ** (R - L)).fill(0).map((_, idx) => {
            const len = 8 * (j + 1);
            const dx = len - R;
            const tmp = (idx << dx) + arr[i];
            return tmp;
          });
        }
        // 双指针落在不同段，中间可能隔了一个段
        if (i < j) {
          for (let k = i; k <= j; k++) {
            // 左指针所在段
            if (k === i) {
              const len = 8 * (i + 1);
              range[k] = new Array(2 ** (len - L)).fill(0).map((_, idx) => {
                const tmp = idx + arr[i];
                return tmp;
              });
            }
            // 中间段
            if (k > i && k < j) range[k] = '0-255';
            // 右指针所在段
            if (k === j) {
              const len = 8 * j;
              const dx = R - len;
              range[k] = new Array(2 ** dx).fill(0).map((_, idx) => {
                const dy = 8 * (j + 1) - R;
                const tmp = idx << dy;
                return tmp;
              });
            }
          }
        }
        setIpRange({
          r1: range[0],
          r2: range[1],
          r3: range[2],
          r4: range[3],
        });
      },
      {
        immediate: true,
      },
    );

    watch(
      [() => formModel, () => newMask.value],
      ([val]) => {
        const { ip1, ip2, ip3, ip4 } = val;
        const res = {
          ip: [ip1, ip2, ip3, ip4].join('.'),
          mask: newMask.value,
          minMask: newMask.value,
          maxMask: maskRange.value.max,
        };
        emit('update:modelValue', res);
      },
      {
        deep: true,
      },
    );

    expose({
      getValue: async () => {
        await Promise.all([o2Ref.value.getValue(), o3Ref.value.getValue(), o4Ref.value.getValue()]);
        return formModel;
      },
      reset: () => {
        resetForm();
        resetIpRange();
        newMask.value = props.ips.range[props.ips.idx].mask;
      }
    });

    return () => (
      <div class={'subnet-input-wrapper'}>
        <div class={'item-wrapper'}>
          <SelectColumn
            disabled={props.disabled || props.ips.range.length < 2}
            class={'w100'}
            modelValue={formModel.ip1}
            list={props.ips.range
              .map((item) => item.ip.split('.')[0])
              .map((v, idx) => ({
                label: v,
                value: idx,
                key: v,
              }))}
            onChange={(idx) => {
              emit('changeIdx', idx);
            }}
          />
        </div>
        .
        <div class={'item-wrapper'}>
          <InputColumn
            type='number'
            key={String(ipRange.r2)}
            v-model={formModel.ip2}
            class={'w100'}
            disabled={ipRange.r2 === -1}
            min={0}
            max={255}
            ref={o2Ref}
            rules={[
              {
                validator: (value: string) => value == '0' || Boolean(value),
                message: '不能为空',
              },
              {
                validator: (value: string) => {
                  const val = +value;
                  if (ipRange.r2 === '0-255' || ipRange.r2 === -1) return true;
                  if (isArray(ipRange.r2) && ipRange.r2.includes(val)) return true;
                  return false;
                },
                message: `范围:${ipRange.r2}`,
              },
            ]}
          />
        </div>
        .
        <div class={'item-wrapper'}>
          <InputColumn
            type='number'
            key={String(ipRange.r3)}
            v-model={formModel.ip3}
            class={'w100'}
            disabled={ipRange.r3 === -1}
            min={0}
            max={255}
            ref={o3Ref}
            rules={[
              {
                validator: (value: string) => value == '0' || Boolean(value),
                message: '不能为空',
              },
              {
                validator: (value: string) => {
                  const val = +value;
                  if (ipRange.r3 === '0-255') return true;
                  if (isArray(ipRange.r3) && ipRange.r3.includes(val)) return true;
                  return false;
                },
                message: `范围:${ipRange.r3}`,
              },
            ]}
          />
        </div>
        .
        <div class={'item-wrapper'}>
          <InputColumn
            type='number'
            key={String(ipRange.r4)}
            v-model={formModel.ip4}
            class={'w100'}
            disabled={ipRange.r4 === -1}
            min={0}
            max={255}
            ref={o4Ref}
            rules={[
              {
                validator: (value: string) => value == '0' || Boolean(value),
                message: '不能为空',
              },
              {
                validator: (value: string) => {
                  const val = +value;
                  if (ipRange.r4 === '0-255') return true;
                  if (isArray(ipRange.r4) && ipRange.r4.includes(val)) return true;
                  return false;
                },
                message: `范围:${ipRange.r4}`,
              },
            ]}
          />
        </div>
        /
        <div class={'item-wrapper'}>
          <InputColumn
            v-model={newMask.value}
            type='number'
            class={'w100'}
            min={maskRange.value.min}
            max={maskRange.value.max}
          />
        </div>
      </div>
    );
  },
});
