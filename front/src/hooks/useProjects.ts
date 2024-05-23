import { viewTypes } from '@/router/router-config';
import { useProjectStore } from '@/stores/projects';
import { computed, onBeforeMount } from 'vue';
import { useRoute, useRouter } from 'vue-router';

export function useProjectList(all = false) {
  const router = useRouter();
  const route = useRoute();

  const projectStore = useProjectStore();
  const projects = computed(() => (all ? projectStore.devopsProjects : projectStore.metricProjects));
  const currentProjectId = computed(() => route.params.projectId as string);
  const currentProject = computed(() =>
    projects.value.list.find((project) => project.project_id === currentProjectId.value),
  );

  onBeforeMount(async () => {
    if (!projects.value.fetched) {
      all ? projectStore.fetchDevopsProjects() : projectStore.fetchMetricProjects();
    }
  });

  function handleProjectChange(projectId: string) {
    router.push({
      ...route,
      params: {
        viewType: viewTypes[0],
        ...route.params,
        projectId,
      },
    });
  }

  function getAdminByDeptId(dept_id: string) {
    return projectStore.orgAdminMap[dept_id]?.username ?? 'unknow';
  }

  return {
    projects,
    currentProjectId,
    currentProject,
    projectStore,
    handleProjectChange,
    getAdminByDeptId,
  };
}
