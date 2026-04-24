## 1. 新增 auth-register.sh 脚本

- [x] 1.1 在容器镜像源码中新建 `auth-register.sh`，脚本从环境变量读取参数（`BK_APP_CODE`、`BK_AUTH_USER_NAME`、`BK_TENANT_ID`、`BK_AUTHSERVER_HOST`、`BK_AUTHSERVER_ENDPOINT`）
- [x] 1.2 脚本中通过 `/proc/sys/kernel/random/uuid`（或 `uuidgen`）生成随机 Request ID
- [x] 1.3 脚本中使用环境变量引用构造 curl 命令，向 `http://$BK_AUTHSERVER_ENDPOINT/api/v1/auth/init/authcenter` 发起 POST 请求，Header 和 JSON body 均通过变量传入
- [x] 1.4 脚本中检查 curl 响应的 `.code` 字段，非 `"0"` 时打印错误并以非零退出码退出
- [x] 1.5 为脚本设置可执行权限（`chmod +x`），并在 Dockerfile 中将其 COPY 至 `/data/bin/auth-register.sh`

## 2. 修改 Helm chart

- [x] 2.1 修改 `docs/support-file/helm/templates/authserver/auth-register-job.yaml`：将 `command` 字段替换为 `command: ["/data/bin/auth-register.sh"]`，删除 `sh -c` 内联 bash here-doc
- [x] 2.2 在 `auth-register-job.yaml` 容器 spec 中新增 `env:` 块，注入以下环境变量：`BK_APP_CODE`、`BK_AUTH_USER_NAME`、`BK_TENANT_ID`、`BK_AUTHSERVER_HOST`、`BK_AUTHSERVER_ENDPOINT`
