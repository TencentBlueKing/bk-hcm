<script lang="ts" setup>
/**
 * 业务视图入口
 *
 * 功能模块划分方案说明：
 * - 每个功能模块（如 host、drive、vpc 等）在 views 目录下有独立目录
 * - 模块目录包含 biz.vue（业务视图入口）和 rsc.vue（资源视图入口）
 * - 通用 UI 组件写在模块目录下，供 biz.vue 和 rsc.vue 复用
 * - 组件不带业务逻辑，只负责展示
 *
 * 核心理念：
 * - 功能模块不受上层菜单变化影响
 * - 菜单与路由是可变的上层配置
 * - 路由决定入口指向的是 biz.vue 还是 rsc.vue
 *
 * TODO: 待实现具体业务逻辑
 */
</script>

<template>
  <div class="host-biz-view">
    <!-- 业务视图入口 -->
  </div>
</template>
