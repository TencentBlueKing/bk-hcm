import { useDepartment } from '@/hooks';
import { Department } from '@/typings';
import { Checkbox, Select, Tree } from 'bkui-vue';
import { computed, defineComponent, PropType, watch } from 'vue';

import './organization-select.scss';

export default defineComponent({
  props: {
    disabled: {
      type: Boolean,
    },
    modelValue: {
      type: Array as PropType<number[]>,
      default: (): number[] => [],
    },
  },
  emits: ['input', 'change', 'update:modelValue'],
  setup(props, ctx) {
    const {
      organizationTree,
      // getParentIds,
      expandDepartment,
      recursionCheckChildDept,
      recursionCheckParentDept,
      departmentMap,
      updateDepartment,
      checkedDept,
    } = useDepartment();
    const isLoading = computed(() => !props.modelValue.every((id) => isAllLoaded(id)));
    const dispalyValue = computed(() => {
      if (!isLoading.value) {
        props.modelValue.forEach((id) => {
          const dept = departmentMap.value.get(id);
          if (!dept.checked) {
            handleCheck(true, dept, false);
          }
        });
      }
      const nameValues = props.modelValue.map((id) => departmentMap.value.get(id)?.full_name ?? id);
      return isLoading.value ? [] : nameValues;
    });

    watch(
      () => isLoading.value,
      async (loading) => {
        if (!loading) {
          props.modelValue.forEach((id) => {
            const dept = departmentMap.value.get(id);
            if (!dept.checked) {
              handleCheck(true, dept, false);
            }
          });
        }
      },
    );

    function isAllLoaded(id: number): boolean {
      if (!id) return true;
      const dept = departmentMap.value.get(id);
      if (!dept) return false;
      if (dept.has_children && dept.loaded) return true;
      return isAllLoaded(dept.parent);
    }

    // async function patchGetParentDept(ids: string) {
    //   const pidMap: Record<string, number[]> = await getParentIds(ids);
    //   const pids = Object.values(pidMap).reduce((
    //     list,
    //     ids,
    //   ) => list.concat(ids.slice(1)), []);

    //   await Promise.all(Array.from(new Set(pids)).map((pid: number) => expandDepartment({
    //     id: pid,
    //   })));
    // }

    function handleCheck(checked: boolean, department: Department, update = true) {
      Array.from(departmentMap.value.values()).forEach((e) => {
        // 只能选中一条
        e.checked = e.id === department.id;
        e.indeterminate = e.id === department.id;
      });
      const fullDept = departmentMap.value.get(department.id);
      const { has_children, loaded, parent, children } = fullDept;
      updateDepartment(department.id, {
        checked: !!checked,
        indeterminate: false,
      });
      if (has_children && loaded) {
        recursionCheckChildDept(children, !!checked);
      }
      if (parent) {
        recursionCheckParentDept(department.id, !!checked);
      }

      if (update) {
        updateValue(checkedDept.value);
      }
    }

    function updateValue(val: number[]) {
      ctx.emit('update:modelValue', val);
    }

    function handleChange(newVal: string[]) {
      if (!newVal) {
        return;
      }
      const newMap = newVal.reduce<Record<string, number>>((acc, name) => {
        acc[name] = 1;
        return acc;
      }, {});

      const newIds = checkedDept.value.filter((id) => {
        const dept = departmentMap.value.get(id);
        const valid = newMap[dept?.full_name];
        if (!newMap[dept?.full_name]) {
          handleCheck(false, dept, false);
        }
        return valid;
      });
      updateValue(newIds);
    }

    const handleToggle = (isOpen: Boolean) => {
      if (!isOpen) {
        ctx.emit('change', dispalyValue.value);
      }
    };

    return () => (
      <Select
        {...ctx.attrs}
        disabled={props.disabled}
        customContent
        modelValue={dispalyValue.value}
        multipleMode='tag'
        multiple={false}
        loading={isLoading.value}
        onChange={handleChange}
        onToggle={handleToggle}
        clearable={false}>
        <Tree node-key='id' showNodeTypeIcon={false} onNodeExpand={expandDepartment} data={organizationTree.value}>
          {{
            nodeAction: ({ __attr__, loading, has_children }: any) => (
              <span class='organization-tree-action-span'>
                {has_children &&
                  // eslint-disable-next-line no-nested-ternary
                  (loading ? (
                    <i class='icon hcm-icon bkhcm-icon-loading-circle organization-tree-action-circle'></i>
                  ) : __attr__.isOpen ? (
                    <i class='icon hcm-icon bkhcm-icon-angle-up-fill'></i>
                  ) : (
                    <i class='icon hcm-icon bkhcm-icon-right-shape'></i>
                  ))}
              </span>
            ),
            node: (department: Department) => (
              <span class='flex-row align-items-center' onClick={(e) => e.stopPropagation()}>
                <Checkbox
                  modelValue={department.checked}
                  indeterminate={department.indeterminate}
                  onChange={(checked) => handleCheck(checked, department)}
                />
                <span class='ml8'>{department.name}</span>
              </span>
            ),
          }}
        </Tree>
      </Select>
    );
  },
});
