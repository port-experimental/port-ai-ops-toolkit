{{- define "copilot-usage-ingestor.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "copilot-usage-ingestor.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{- define "copilot-usage-ingestor.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version -}}
{{- end -}}

{{- define "copilot-usage-ingestor.labels" -}}
helm.sh/chart: {{ include "copilot-usage-ingestor.chart" . }}
app.kubernetes.io/name: {{ include "copilot-usage-ingestor.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "copilot-usage-ingestor.secretName" -}}
{{- if .Values.secret.nameOverride -}}
{{- .Values.secret.nameOverride -}}
{{- else -}}
{{- printf "%s-secret" (include "copilot-usage-ingestor.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
