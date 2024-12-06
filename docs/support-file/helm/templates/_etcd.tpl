{{/* vim: set filetype=mustache: */}}
{{/*
内建 Etcd 名称
*/}}
{{- define "bk-hcm.etcdName" -}}
{{- if .Values.etcd.fullnameOverride -}}
{{- .Values.etcd.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default "etcd" .Values.etcd.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{- define "bk-hcm.etcdSecretName" -}}
{{- if .Values.externalEtcd.tls.enabled -}}
    {{- if .Values.externalEtcd.tls.certBase64Encoded }}
        {{- printf "%s-etcd-certs" (include "bk-hcm.fullname" .) -}}
    {{- else }}
        {{- .Values.externalEtcd.tls.existingSecret -}}
    {{- end -}}
{{- end -}}
{{- end -}}

{{/*
生成etcd yaml配置
*/}}
{{- define "bk-hcm.etcdConfig" -}}
{{- if .Values.etcd.enabled -}}
endpoints:
  - {{ include "bk-hcm.etcdName" . }}:2379
dialTimeoutMS:
username: root
password: {{ .Values.etcd.auth.rbac.rootPassword }}
tls:
  insecureSkipVerify:
  certFile:
  keyFile:
  caFile:
  password:
{{- else -}}
endpoints:
  - {{ .Values.externalEtcd.host }}:{{ .Values.externalEtcd.port }}
dialTimeoutMS:
username: {{ .Values.externalEtcd.username }}
password: {{ .Values.externalEtcd.password }}
{{- if .Values.externalEtcd.tls.enabled -}}
tls:
  insecureSkipVerify: {{ .Values.externalEtcd.tls.insecureSkipVerify }}
  certFile: "/data/hcm/etc/certs/{{ .Values.externalEtcd.tls.certFileName }}"
  keyFile: "/data/hcm/etc/certs/{{ .Values.externalEtcd.tls.keyFileName }}"
  caFile: "/data/hcm/etc/certs/{{ .Values.externalEtcd.tls.caCertFileName }}"
  password:
{{- end -}}
{{- end -}}
{{- end -}}
