{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}

{{- define "k8scostoptimizer-name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" | lower -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "k8scostoptimizer-fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{- define "k8scostoptimizer-labels.chart" -}}
app: {{ template "k8scostoptimizer-fullname" . }}
chart: "{{ .Chart.Name }}-{{ .Chart.Version }}"
release: {{ .Release.Name | quote }}
heritage: {{ .Release.Service | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service | quote }}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "k8scostoptimizer-serviceAccountName" -}}
{{- if .Values.k8scostoptimizer.serviceAccount.create -}}
    {{ default (include "k8scostoptimizer-fullname" .) .Values.k8scostoptimizer.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.k8scostoptimizer.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/*
Create the annotations to support helm3
*/}}
{{- define "k8scostoptimizer-helm3.annotations" -}}
meta.helm.sh/release-namespace: {{ .Release.Namespace | quote }}
meta.helm.sh/release-name: {{ .Release.Name | quote }}
{{- end -}}