{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "bk-hcm.name" -}}
{{- include "common.names.name" . -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "bk-hcm.fullname" -}}
{{- include "common.names.fullname" . -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "bk-hcm.chart" -}}
{{- include "common.names.chart" . -}}
{{- end -}}

{{/*
Return the proper Docker Image Registry Secret Names
*/}}
{{- define "bk-hcm.imagePullSecrets" -}}
{{ include "common.images.pullSecrets" (dict "images" (list .Values.image) "global" .Values.global) }}
{{- end -}}

{{/*
 Create the name of the service account to use
 */}}
{{- define "bk-hcm.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "bk-hcm.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/*
Wait for pod
*/}}
{{- define "bk-hcm.wait-for-pod-init-container" -}}
{{- $root := first . -}}
{{- $rest := rest . -}}
{{- $name := last . -}}
- name: {{ printf "check-%s"  (index $rest 0) }}
  image: {{ printf "%s/%s:%s" $root.Values.global.imageRegistry $root.Values.k8sWaitFor.repository  $root.Values.k8sWaitFor.tag}}
  imagePullPolicy: {{ $root.Values.global.imagePullPolicy | quote }}
  resources: {{ toYaml $root.Values.k8sWaitFor.resources | nindent 4 }}
  args:
    - pod
    - {{ $name }}
{{- end }}


{{/*
Returns http port for service
*/}}
{{- define "bk-hcm.getHttpServicePort" -}}
{{- $firstPort := first .ports }}
{{- $value := get $firstPort "port" }}
{{- range .ports }}
  {{- if eq .name "http" -}}
    {{- $value = .port }}
  {{- end -}}
{{- end -}}
{{- print $value }}
{{- end -}}
