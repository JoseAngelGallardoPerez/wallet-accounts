{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "wallet-accounts.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "wallet-accounts.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "wallet-accounts.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "wallet-accounts.labels" -}}
helm.sh/chart: {{ include "wallet-accounts.chart" . }}
{{ include "wallet-accounts.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "wallet-accounts.selectorLabels" -}}
app.kubernetes.io/name: {{ include "wallet-accounts.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "wallet-accounts.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "wallet-accounts.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create tag name of the image
*/}}
{{- define "wallet-accounts.imageTag" -}}
{{ .Values.image.tag | default .Chart.AppVersion }}
{{- end }}

{{/*
Create the name of the image repository
*/}}
{{- define "wallet-accounts.imageRepository" -}}
{{ .Values.image.repository | default (printf "velmie/%s" .Chart.Name) }}
{{- end }}

{{/*
Create full image repository name including tag
*/}}
{{- define "wallet-accounts.imageRepositoryWithTag" -}}
{{ include "wallet-accounts.imageRepository" . }}:{{ include "wallet-accounts.imageTag" . }}
{{- end }}

{{/*
Create full database migration image repository name
*/}}
{{- define "wallet-accounts.dbMigrationImageRepositoryWithTag" -}}
{{ include "wallet-accounts.imageRepository" . }}-db-migration:{{ include "wallet-accounts.imageTag" . }}
{{- end }}