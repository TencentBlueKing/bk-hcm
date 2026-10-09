import 'vue-router';
import { type RouteMetaConfig } from '@/router/meta';

declare module 'vue-router' {
  // eslint-disable-next-line @typescript-eslint/no-empty-interface
  interface RouteMeta extends RouteMetaConfig {}
}
