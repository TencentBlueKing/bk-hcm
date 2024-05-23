import { useDepartmentStore } from '@/store/departments';
import { Department } from '@/typings';
import { computed, ref, watch } from 'vue';

export function useDepartment() {
  const departmentStore = useDepartmentStore();

  const departmentMap = ref<Map<number, Department>>(generateDeptTreeMap());
  const organizationTree = computed(() =>
    Array.from(departmentMap.value.values()).filter((department) => !department.parent),
  );
  const checkedDept = computed(() => getCheckedDept(organizationTree.value));

  if (departmentStore.departmentMap.size === 0) {
    departmentStore.getRootDept();
  }

  watch(
    () => departmentStore.departmentMap,
    (deptMap) => {
      departmentMap.value = generateDeptTreeMap(deptMap);
    },
    {
      deep: true,
    },
  );

  function generateDeptTreeMap(deptMap?: Map<number, Department>) {
    const originDepartmentMap = deptMap ?? departmentStore.departmentMap;
    const deptList = Array.from(originDepartmentMap.values());
    const newDepartmentMap: Map<number, Department> = deptList.reduce((acc, department) => {
      const curDept = departmentMap?.value?.get(department.id);
      const parent = departmentMap?.value?.get(department.parent);
      const isChecked = parent?.checked;

      acc.set(department.id, {
        ...department,
        ...(department.has_children
          ? {
              children: [],
              async: true,
            }
          : {}),
        checked: (isChecked || curDept?.checked) ?? false,
        indeterminate: curDept?.indeterminate ?? false,
      });
      return acc;
    }, new Map());

    Array.from(newDepartmentMap.values()).forEach((dept) => {
      const parent = newDepartmentMap.get(dept.parent);
      if (Array.isArray(parent?.children)) {
        parent.children.push(dept);
        parent.loaded = true;
        parent.isOpen = false;
      }
    });
    return newDepartmentMap;
  }

  async function expandDepartment({ id }: Partial<Department>) {
    const dept = departmentMap.value.get(id);
    if (!dept || (dept.has_children && !dept.loaded)) {
      await departmentStore.getChildDept(id);
      updateDepartment(id, {
        isOpen: true,
      });
    }
  }

  function updateDepartment(id: number, params: Partial<Department>) {
    const dept = departmentMap.value.get(id);
    if (dept) {
      departmentMap.value.set(id, Object.assign(dept, params));
    }
  }

  function patchUpdateDepartment(deptMap: Map<number, Department>) {
    Array.from(deptMap.entries()).forEach((value) => {
      updateDepartment(value[0], value[1]);
    });
  }

  function recursionCheckChildDept(list: Department[], checked: boolean) {
    return list.forEach((item) => {
      updateDepartment(item.id, {
        checked,
        indeterminate: false,
      });
      if (item.has_children && item.loaded) {
        recursionCheckChildDept(item.children, checked);
      }
    });
  }

  function recursionCheckParentDept(id: number, checked: boolean): Record<number, Department> {
    const curDept = departmentMap.value.get(id);
    const parent = departmentMap.value.get(curDept.parent);
    if (parent) {
      const indeterminate = isHalf(parent.children);

      departmentMap.value.set(
        parent.id,
        Object.assign(parent, {
          checked: checked && !indeterminate,
          indeterminate,
        }),
      );

      if (parent.parent) {
        return recursionCheckParentDept(parent.id, checked);
      }
    }
  }

  function isHalf(children: Department[]) {
    let checkedLength = 0;
    let halfLength = 0;
    children.forEach((item) => {
      if (item.checked) checkedLength += 1;
      if (item.indeterminate) halfLength += 1;
    });

    return (checkedLength > 0 && checkedLength < children.length) || halfLength > 0;
  }

  function getCheckedDept(list = organizationTree.value) {
    return list.reduce((acc: number[], item: Department) => {
      if (item.checked && !item.indeterminate) {
        // acc.push(item.id);
        acc = [item.id];
      } else if (item.indeterminate) {
        // acc.push(...getCheckedDept(item.children));
        acc = getCheckedDept(item.children);
      }
      return acc;
    }, []);
  }

  return {
    getParentIds: departmentStore.getParentIds,
    organizationTree,
    departmentMap,
    expandDepartment,
    updateDepartment,
    patchUpdateDepartment,
    recursionCheckChildDept,
    recursionCheckParentDept,
    checkedDept,
  };
}
