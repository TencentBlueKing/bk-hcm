<script setup lang="ts">
import { watch, useId, ref, getCurrentInstance } from 'vue';
import isEqual from 'lodash/isEqual';
import { useAuthStore } from '@/store/auth';
import CombineRequest from '@blueking/combine-request';
import { type IAuthSign, getAuthDefs, getAuthResources } from '@/common/auth-service';
import { type IVerifyResult } from '@/store/auth';
import usePermissionDialog from '@/hooks/use-permission-dialog';

export interface IAuthProps {
  sign: IAuthSign | IAuthSign[];
  tag?: string;
}

defineOptions({ name: 'hcm-auth' });

const props = withDefaults(defineProps<IAuthProps>(), {
  tag: 'div',
});

const emit = defineEmits<(e: 'done') => void>();

const instance = getCurrentInstance();

// 组件唯一id
const id = useId();

// 鉴权状态数据
const isAuthorized = ref(false);
const verified = ref(false);
const noPerm = ref(true);

// 鉴权结果的permission数据
let permission: IVerifyResult['permission'];

const authStore = useAuthStore();

const permissionDialog = usePermissionDialog();

const combineRequest = CombineRequest.setup(Symbol.for('hcm-auth'), async (signCompList) => {
  // sign的唯一性列表，同一sign单个与数组形式将被会认为不同
  const uniqueSign: (IAuthSign | IAuthSign[])[] = [];

  // 组件与鉴权结果数据组，
  // 第一个元素为组件id，
  // 第二个元素为鉴权结果的数据读取范围，对应鉴权结果的数组index
  // 第三个元素为与当前组件sign一致的组件id，包括当前组件id
  const compAuthResultGroup: [string, number[], string[]][] = [];

  let lastLength = 0;
  (signCompList as [IAuthSign | IAuthSign[], string][]).forEach(([sign, compId]) => {
    // 每次先为当前compId创建一条记录
    const currentGroup: [string, number[], string[]] = [compId, [], [compId]];
    compAuthResultGroup.push(currentGroup);

    // sign在列表中的位置
    const signIndex = uniqueSign.findIndex((val) => isEqual(val, sign));
    if (signIndex === -1) {
      // 设置起始点，初始为0否则为上一个sign的结束点，也就是最后的长度
      currentGroup?.[1]?.push(lastLength);

      // sign放入列表
      uniqueSign.push(sign);

      // 设置结束点，起始点 + sign的长度非数组时为1
      lastLength = Array.isArray(sign) ? lastLength + sign.length : lastLength + 1;
      currentGroup?.[1]?.push(lastLength);
    } else {
      // sign已经存在，则直接使用一致的读取范围即可
      const exist = compAuthResultGroup[signIndex];

      // 加入一致的组件id
      exist[2].push(compId);

      // 分别更新鉴权结果的读取范围与sign一致的组件id
      [, currentGroup[1], currentGroup[2]] = exist;
    }
  });

  const { results, permission } = await authStore.verify(uniqueSign.flat());

  const compAuthMap = new Map<string, [IVerifyResult['results'], string[]]>();
  for (const [compId, range, sameSignComp] of compAuthResultGroup) {
    compAuthMap.set(compId, [results.slice(...range), sameSignComp]);
  }
  return { compAuthMap, permission };
});

watch(
  () => props.sign,
  async (sign, oldSign) => {
    if (!isEqual(sign, oldSign)) {
      // 按[sign, 组件id]格式添加，sign可能是单元素也可能是数组
      combineRequest.add([sign, id]);

      // 获取合并请求后的返回结果，包含本次所有需要鉴权的组件id与鉴权结果及组件id与实例的映射
      const requestResult = await combineRequest.getPromise();

      // 取出permission
      permission = requestResult.permission;

      const [authorizedList] = requestResult.compAuthMap.get(id);

      // 设置鉴权状态
      const authorized = authorizedList.every((item) => item.authorized);
      isAuthorized.value = authorized;
      noPerm.value = !authorized;
      verified.value = true;
    }
  },
  { immediate: true },
);

const handleClick = async () => {
  if (isAuthorized.value) {
    return;
  }

  // 获取当前组件sign的权限定义
  const authIds = getAuthDefs(props.sign).map((item) => item.id);

  // 获取当前组件sign的资源配置
  const resources = getAuthResources(props.sign);

  // 取得当前组件sign所对应的permission数据
  const { actions, ...others } = permission;

  const currentActions = actions
    // 先通过id过滤，只保留当前组件所对应的action
    .filter((item) => authIds.includes(item.id))
    .map((item) => ({
      ...item,
      related_resource_types: item.related_resource_types.map((resourceItem) => ({
        ...resourceItem,
        // TODO: 支持多层级
        // 只保留当前组件所对应需要申请的资源实例
        // 目前的业务场景没有多层级，可以先打平处理，为满足内部数据格式需要手动构造为二层结构且过滤掉空数据兼容未指定实例的情况
        instances: [
          resourceItem?.instances
            ?.flat()
            .filter((instance) =>
              resources.some((resource) =>
                [String(resource.bk_biz_id), String(resource.resource_id)].includes(String(instance.id)),
              ),
            ),
        ].filter(Boolean),
      })),
    }));

  const compPermission = {
    actions: currentActions,
    ...others,
  };

  // 传入permission及相关配置，done为点击已申请时触发的回调函数，当绑定了done事件时则触发事件，否则置空使用dialog默认刷新逻辑
  permissionDialog.show(compPermission, {
    done: instance?.vnode?.props?.onDone ? () => emit('done') : null,
  });
};
</script>

<template>
  <component :is="tag" :class="['hcm-auth', { disabled: noPerm, verified }]" @click="handleClick">
    <slot v-bind="{ noPerm }" :class="{ locked: noPerm }"></slot>
  </component>
</template>

<style lang="scss" scoped>
.hcm-auth {
  display: inline-block;
  pointer-events: none;

  &.verified {
    pointer-events: auto;
  }

  &.disabled {
    position: relative;
    cursor: url('@/assets/image/lock.svg'), not-allowed;

    /* 透明覆盖层接收点击与保持lock手势 */
    &::after {
      content: '';
      position: absolute;
      top: 0;
      left: 0;
      right: 0;
      bottom: 0;
      z-index: 1;
    }
  }
}
</style>
